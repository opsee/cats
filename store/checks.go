package store

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/checks"
)

type dbCheck struct {
	*schema.Check
	*schema.Target
}

type checkStore struct {
	sqlx.Ext
}

func NewCheckStore(q sqlx.Ext) CheckStore {
	return &checkStore{q}
}

// GetState creates a State object populated by the check's settings and
// by the current state if it exists. If it the state is unknown, then it
// assumes a present state of OK.
func (q *checkStore) GetAndLockState(customerId, checkId string) (*checks.State, error) {
	state := &checks.State{}
	err := sqlx.Get(q, state, "SELECT states.state_id, states.customer_id, states.check_id, states.state_name, states.time_entered, states.last_updated, checks.min_failing_count, checks.min_failing_time, states.failing_count, states.response_count FROM check_states AS states JOIN checks ON (checks.id = states.check_id) WHERE states.customer_id = $1 AND checks.id = $2 AND checks.deleted = false FOR UPDATE OF states", customerId, checkId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		// Get the check so that we can get MinFailingCount and MinFailingTime
		// Return an error if the check doesn't exist
		// check, err := store.GetCheck(customerId, checkId)
		check := &schema.Check{}
		err := sqlx.Get(q, check, "SELECT id, customer_id, min_failing_count, min_failing_time FROM checks WHERE customer_id = $1 AND id = $2 and deleted = false", customerId, checkId)
		if err != nil {
			return nil, err
		}

		state = &checks.State{
			CheckId:         checkId,
			CustomerId:      customerId,
			Id:              checks.StateOK,
			State:           checks.StateOK.String(),
			TimeEntered:     time.Now(),
			LastUpdated:     time.Now(),
			MinFailingCount: check.MinFailingCount,
			MinFailingTime:  time.Duration(check.MinFailingTime) * time.Second,
			FailingCount:    0,
		}
	}

	state.MinFailingTime = state.MinFailingTime * time.Second

	return state, nil
}

func (q *checkStore) UpdateState(state *checks.State) error {
	row := q.QueryRowx("SELECT sum(failing_count), sum(response_count) FROM check_state_memos WHERE check_id=$1 AND customer_id=$2", state.CheckId, state.CustomerId)
	if err := row.Err(); err != nil {
		return err
	}
	var failingCount, responseCount int
	row.Scan(&failingCount, &responseCount)
	state.FailingCount = int32(failingCount)
	state.ResponseCount = int32(responseCount)

	return nil
}

func (q *checkStore) PutState(state *checks.State) error {
	_, err := sqlx.NamedExec(q, "INSERT INTO check_states (check_id, customer_id, state_id, state_name, time_entered, last_updated, failing_count, response_count) VALUES (:check_id, :customer_id, :state_id, :state_name, :time_entered, :last_updated, :failing_count, :response_count) ON CONFLICT (check_id) DO UPDATE SET state_id = :state_id, state_name = :state_name, time_entered = :time_entered, last_updated = :last_updated, failing_count = :failing_count, response_count = :response_count", state)
	if err != nil {
		return err
	}

	return nil
}

func (q *checkStore) PutMemo(memo *checks.ResultMemo) error {
	_, err := sqlx.NamedExec(q, "INSERT INTO check_state_memos AS csm (check_id, customer_id, bastion_id, failing_count, response_count, last_updated) VALUES (:check_id, :customer_id, :bastion_id, :failing_count, :response_count, :last_updated) ON CONFLICT (check_id, bastion_id) DO UPDATE SET failing_count = :failing_count, response_count = :response_count, last_updated = :last_updated WHERE csm.check_id = :check_id AND csm.bastion_id = :bastion_id", memo)
	if err != nil {
		return err
	}

	return nil
}

func (q *checkStore) GetMemo(checkId, bastionId string) (*checks.ResultMemo, error) {
	memo := &checks.ResultMemo{}
	err := sqlx.Get(q, memo, "SELECT * FROM check_state_memos WHERE check_id = $1 AND bastion_id = $2 LIMIT 1", checkId, bastionId)
	if err != nil {
		return nil, err
	}

	return memo, nil
}

// GetLiveBastions returns a list of bastions whose timestamps do not differ by
// greater than 120 seconds. It's important not to simply look at NOW() - 120 seconds
// because in periods of time where we aren't processing results, this could cause us
// to throw out results from "live" bastions.
func (q *checkStore) GetLiveBastions(customerID, checkID string) (bastions []string, err error) {
	var memos []*checks.ResultMemo
	err = sqlx.Select(q, &memos, "SELECT * FROM check_state_memos WHERE check_id = $1 ORDER BY last_updated DESC", checkID)
	if err != nil {
		return bastions, err
	}
	bastions = make([]string, 0, len(memos))
	// The newest timestamp is the upper bound of our 120 second window. So in O(n) we can
	// found our set of "live" bastions.
	var (
		lowerBound time.Time
		bIdx       int
	)
	for _, m := range memos {
		if m.LastUpdated.After(lowerBound) {
			bIdx++
			bastions = bastions[:bIdx]
			bastions[bIdx-1] = m.BastionId
		}
	}
	return bastions, nil
}

// CreateStateTransitionLogEntry creates and stores a StateTransitionLogEntry, returning the created
// log entry or an error.
func (q *checkStore) CreateStateTransitionLogEntry(checkId, customerId string, fromState, toState checks.StateId) (*checks.StateTransitionLogEntry, error) {
	var logEntryID int
	err := q.QueryRowx("INSERT INTO check_state_transitions (check_id, customer_id, from_state, to_state) VALUES ($1, $2, $3, $4) RETURNING id", checkId, customerId, fromState, toState).Scan(&logEntryID)
	if err != nil {
		return nil, err
	}

	logEntry := &checks.StateTransitionLogEntry{}
	err = q.QueryRowx("SELECT * FROM check_state_transitions WHERE id=$1", logEntryID).StructScan(logEntry)
	if err != nil {
		return nil, err
	}
	return logEntry, nil
}

func (q *checkStore) GetCheckCount(customerId string) (int32, error) {
	var count int32 = 0
	err := sqlx.Get(q, &count, "select count(1) from checks where customer_id = $1 and deleted = false", customerId)
	if err != nil && err != sql.ErrNoRows {
		return count, err
	}

	return count, nil
}

// GetStateTransitionLogEntries returns state transition log entries between a start and end time
// log entry or an error.
func (q *checkStore) GetCheckStateTransitionLogEntries(checkId, customerId string, from, to time.Time) ([]*checks.StateTransitionLogEntry, error) {
	var logEntries []*checks.StateTransitionLogEntry

	err := sqlx.Select(q, &logEntries, "SELECT * FROM check_state_transitions WHERE check_id=$1 AND customer_id=$2 AND created_at BETWEEN $3 AND $4", checkId, customerId, from, to)
	if err != nil {
		return nil, err
	}

	return logEntries, nil
}

// GetChecks gets all checks for a customer
func (q *checkStore) GetChecks(user *schema.User) (checks []*schema.Check, err error) {
	dbcs := []dbCheck{}
	err = sqlx.Select(q, &dbcs, "SELECT id, COALESCE(interval, 30) AS interval, check_spec, customer_id, name, execution_group_id, min_failing_count, min_failing_time, target_name, target_type, target_id FROM checks WHERE customer_id=$1 AND deleted=false", user.CustomerId)
	if err != nil {
		return nil, err
	}

	checks = make([]*schema.Check, len(dbcs))
	for i, c := range dbcs {
		check := c.Check
		check.Target = c.Target
		checks[i] = check

		err = sqlx.Select(q, &check.Assertions, `SELECT key, value, operand, relationship FROM assertions WHERE check_id=$1 AND customer_id=$2`, check.Id, check.CustomerId)
		if err != nil {
			return nil, err
		}
	}
	return checks, nil
}

// GetCheck gets a single check for a customer
func (q *checkStore) GetCheck(user *schema.User, checkId string) (check *schema.Check, err error) {
	bullshit := &dbCheck{}
	err = q.QueryRowx("SELECT id, COALESCE(interval, 30) AS interval, check_spec, customer_id, name, execution_group_id, min_failing_count, min_failing_time, target_name, target_type, target_id FROM checks WHERE id=$1 AND customer_id=$2 AND deleted=false", checkId, user.CustomerId).StructScan(bullshit)
	if err != nil {
		return nil, err
	}

	check = bullshit.Check
	check.Target = bullshit.Target

	err = sqlx.Select(q, &check.Assertions, `SELECT key, value, operand, relationship FROM assertions WHERE check_id=$1 AND customer_id=$2`, check.Id, check.CustomerId)
	if err != nil {
		return nil, err
	}

	return check, nil
}

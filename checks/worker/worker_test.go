package worker

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/checks"
	"github.com/opsee/cats/store"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func mockResult(responseCount, failingCount int) *schema.CheckResult {
	ts := &opsee_types.Timestamp{}
	ts.Scan(time.Now())
	r := &schema.CheckResult{
		CheckId:    "check-id",
		CustomerId: "11111111-1111-1111-1111-111111111111",
		BastionId:  "61f25e94-4f6e-11e5-a99f-4771161a3518",
		Responses:  make([]*schema.CheckResponse, responseCount),
		Timestamp:  ts,
	}

	for i := 0; i < responseCount; i++ {
		r.Responses[i] = &schema.CheckResponse{}
		if i >= failingCount {
			r.Responses[i].Passing = true
		}
	}

	return r
}

type fakeStore struct {
	fail bool
}

func (s *fakeStore) PutResult(result *schema.CheckResult) error {
	if s.fail {
		return errors.New("")
	}

	return nil
}

func (s *fakeStore) GetResultsByCheckId(checkId string) ([]*schema.CheckResult, error) {
	if s.fail {
		return nil, errors.New("")
	}

	return nil, nil
}

func TestDeletedCheck(t *testing.T) {
	db := testSetupFixtures()
	db.MustExec("update checks set deleted = true")
	result := mockResult(2, 1)

	wrkr := NewCheckWorker(db, result)
	_, err := wrkr.Execute()
	assert.Nil(t, err)
	// make sure no check state has been created
	r, err := db.Queryx("select * from check_states")
	assert.Nil(t, err)
	defer r.Close()
	assert.Equal(t, false, r.Next())
}

func TestExistingState(t *testing.T) {
	db := testSetupFixtures()
	result := mockResult(2, 1)
	ts := &opsee_types.Timestamp{}
	ts.Scan(time.Now().Add(30 * time.Second))
	result.Timestamp = ts

	state := &checks.State{
		CheckId:     "check-id",
		CustomerId:  "11111111-1111-1111-1111-111111111111",
		Id:          checks.StateOK,
		State:       "OK",
		TimeEntered: time.Now(),
		LastUpdated: time.Now(),
	}

	err := store.PutState(db, state)
	assert.Nil(t, err)

	wrkr := NewCheckWorker(db, result)
	_, err = wrkr.Execute()
	assert.Nil(t, err)

	tx, err := db.Beginx()
	assert.Nil(t, err)
	state, err = store.GetAndLockState(tx, result.CustomerId, result.CheckId)
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, "FAIL_WAIT", state.State)
	assert.Equal(t, int32(1), state.FailingCount)
	assert.Equal(t, int32(2), state.ResponseCount)
	tx.Commit()
}

func testSetupFixtures() *sqlx.DB {
	db, err := sqlx.Open("postgres", viper.GetString("postgres_conn"))
	if err != nil {
		panic(err)
	}

	db.MustExec("DELETE FROM checks")
	db.MustExec("DELETE FROM check_states")
	db.MustExec("DELETE FROM check_state_memos")

	check := &schema.Check{
		Id:               "check-id",
		CustomerId:       "11111111-1111-1111-1111-111111111111",
		ExecutionGroupId: "11111111-1111-1111-1111-111111111111",
		Name:             "check",
		MinFailingCount:  1,
		MinFailingTime:   90,
	}
	_, err = sqlx.NamedExec(db, "INSERT INTO checks (id, min_failing_count, min_failing_time, customer_id, execution_group_id, name, target_type, target_id) VALUES (:id, :min_failing_count, :min_failing_time, :customer_id, :execution_group_id, :name, 'target-id', 'target-type')", check)
	if err != nil {
		panic(err)
	}

	return db
}

func TestMain(m *testing.M) {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	os.Exit(m.Run())
}

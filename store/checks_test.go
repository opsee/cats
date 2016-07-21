package store

import (
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/cats/checks"
	"github.com/opsee/cats/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTransactionIsolation(t *testing.T) {
	db, err := sqlx.Open("postgres", viper.GetString("postgres_conn"))
	assert.Nil(t, err)

	checkStore := NewCheckStore(db)

	// First we make sure we have memos and state.
	state := &checks.State{
		CheckId:     "check-id",
		CustomerId:  "11111111-1111-1111-1111-111111111111",
		Id:          checks.StateOK,
		State:       checks.StateOK.String(),
		TimeEntered: time.Now(),
		LastUpdated: time.Now(),
	}
	err = checkStore.PutState(state)
	assert.Nil(t, err)

	err = checkStore.PutMemo(&checks.ResultMemo{
		BastionId:     "61f25e94-4f6e-11e5-a99f-4771161a3518",
		CustomerId:    "11111111-1111-1111-1111-111111111111",
		CheckId:       "check-id",
		FailingCount:  0,
		ResponseCount: 2,
	})
	assert.Nil(t, err)

	err = checkStore.PutMemo(&checks.ResultMemo{
		BastionId:     "61f25e94-4f6e-11e5-a99f-4771161a3517",
		CustomerId:    "11111111-1111-1111-1111-111111111111",
		CheckId:       "check-id",
		FailingCount:  0,
		ResponseCount: 2,
	})
	assert.Nil(t, err)

	// There is some non-determinism here... we don't know which goroutine is
	// going to get the row lock first. Huzzah, testing actual concurrency.
	// So, whatever happens, the operation we're doing does have to have a
	// predictable output.

	fakeworker := func(wg *sync.WaitGroup, bastionId string) {
		defer wg.Done()

		tx, err := db.Beginx()
		assert.Nil(t, err)

		checkStore := NewCheckStore(tx)

		state, err = checkStore.GetAndLockState("11111111-1111-1111-1111-111111111111", "check-id")
		assert.Nil(t, err)
		assert.NotNil(t, state)

		err = checkStore.PutMemo(&checks.ResultMemo{
			BastionId:     bastionId,
			CustomerId:    "11111111-1111-1111-1111-111111111111",
			CheckId:       "check-id",
			FailingCount:  2,
			ResponseCount: 2,
		})
		assert.Nil(t, err)

		assert.Nil(t, checkStore.UpdateState(state))
		assert.Nil(t, state.Transition(nil))
		assert.Nil(t, checkStore.PutState(state))
		tx.Commit()
	}
	wg := &sync.WaitGroup{}
	// no guarantee which one locks first, but the outcome should be the same.
	wg.Add(2)
	go fakeworker(wg, "61f25e94-4f6e-11e5-a99f-4771161a3517")
	go fakeworker(wg, "61f25e94-4f6e-11e5-a99f-4771161a3518")
	wg.Wait()

	tx, err := db.Beginx()
	assert.Nil(t, err)
	checkStore = NewCheckStore(tx)

	defer tx.Rollback()
	state, err = checkStore.GetAndLockState("11111111-1111-1111-1111-111111111111", "check-id")
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, int32(4), state.FailingCount)
	assert.Equal(t, int32(4), state.ResponseCount)
}

func TestCheckCount(t *testing.T) {
	assert := assert.New(t)

	withCheckFixtures(func(cs CheckStore) {
		count, err := cs.GetCheckCount(testutil.Checks["1"].CustomerId)
		if err != nil {
			t.Fatal(err)
		}

		assert.EqualValues(2, count)

		// non existent customer returns 0
		count, err = cs.GetCheckCount("11111111-1111-1111-1111-111111111222")
		if err != nil {
			t.Fatal(err)
		}

		assert.EqualValues(0, count)
	})
}

func withCheckFixtures(testFun func(CheckStore)) {
	db, err := sqlx.Open("postgres", viper.GetString("postgres_conn"))
	if err != nil {
		panic(err)
	}

	db.MustExec("delete from checks")
	db.MustExec("delete from check_states")
	db.MustExec("delete from check_state_memos")
	db.MustExec("delete from assertions")

	tx, err := db.Beginx()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	for _, t := range testutil.Checks {
		rows, err := sqlx.NamedQuery(
			tx,
			`INSERT INTO checks (id, min_failing_count, min_failing_time, customer_id,
			 execution_group_id, name, target_type, target_id) VALUES (:id, :min_failing_count,
			 :min_failing_time, :customer_id, :execution_group_id, :name, 'target-id', 'target-type')`,
			t,
		)
		if err != nil {
			panic(err)
		}

		var sub struct {
			Id int
		}
		for rows.Next() {
			rows.StructScan(&sub)
		}
		rows.Close()
	}

	// create a new check store with the transaction and give it to our test funcs
	testFun(&checkStore{tx})
}

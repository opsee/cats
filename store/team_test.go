package store

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/schema"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTeamGet(t *testing.T) {
	assert := assert.New(t)

	withTeamFixtures(func(q TeamStore) {
		// getting active teams works
		team, err := q.Get("11111111-1111-1111-1111-111111111111")
		assert.Nil(err)
		assert.NotNil(team)
		assert.Equal("barbell brigade death squad crew", team.Name)
		assert.Equal("basic", team.Subscription)

		// we don't return inactive teams
		team, err = q.Get("00000000-0000-0000-0000-000000000000")
		assert.Nil(err)
		assert.Nil(team)
	})
}

func withTeamFixtures(testFun func(TeamStore)) {
	db, err := sqlx.Open("postgres", viper.GetString("postgres_conn"))
	if err != nil {
		panic(err)
	}

	db.MustExec("delete from customers")
	db.MustExec("delete from users")
	db.MustExec("delete from signups")

	tx, err := db.Beginx()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	// we use a custom func to insert so that we can control the ids going in
	insertTeam := func(t *schema.Team, active bool) {
		_, err = sqlx.NamedExec(
			tx,
			fmt.Sprintf(`insert into customers (id, name, active, subscription, stripe_customer_id, stripe_subscription_id, subscription_quantity) values
			(:id, :name, %t, :subscription, :stripe_customer_id, :stripe_subscription_id, :subscription_quantity)`, active),
			t,
		)
		if err != nil {
			panic(err)
		}
	}

	insertTeam(&schema.Team{
		Id:                   "11111111-1111-1111-1111-111111111111",
		Name:                 "barbell brigade death squad crew",
		Subscription:         "basic",
		StripeCustomerId:     "cus_8oux3kULDWgU8F",
		StripeSubscriptionId: "",
		SubscriptionQuantity: int32(3),
	}, true)
	insertTeam(&schema.Team{
		Id:                   "00000000-0000-0000-0000-000000000000",
		Name:                 "INACTIVE",
		Subscription:         "basic",
		StripeCustomerId:     "",
		StripeSubscriptionId: "",
		SubscriptionQuantity: int32(3),
	}, false)

	// create a new team store with the transaction and give it to our test funcs
	testFun(NewTeamStore(tx))
}

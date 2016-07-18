package store

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/cats/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestTeamGet(t *testing.T) {
	assert := assert.New(t)

	withTeamFixtures(func(q TeamStore) {
		// getting active teams works
		team, err := q.Get("11111111-1111-1111-1111-111111111111")
		assert.NoError(err)
		assert.NotNil(team)
		assert.Equal("barbell brigade death squad crew", team.Name)
		assert.Equal("beta", team.Subscription)
		assert.Equal(4, len(team.Users))

		// we don't return inactive teams
		team, err = q.Get("00000000-0000-0000-0000-000000000000")
		assert.Nil(err)
		assert.Nil(team)
	})
}

func TestTeamUpdate(t *testing.T) {
	assert := assert.New(t)

	withTeamFixtures(func(q TeamStore) {
		team, err := q.Get("11111111-1111-1111-1111-111111111111")
		assert.Nil(err)
		assert.NotNil(team)

		// change subscription to team plan, increase quantity
		team.Subscription = "team_monthly"
		team.SubscriptionQuantity = 5
		// hey, update the name while we're at it
		team.Name = "money"

		err = q.Update(team)
		assert.NoError(err)
		assert.Equal("money", team.Name)
		assert.Equal("team_monthly", team.Subscription)
		assert.EqualValues(5, team.SubscriptionQuantity)
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

	for k, t := range testutil.Teams {
		_, err = sqlx.NamedExec(
			tx,
			fmt.Sprintf(`insert into customers (id, name, active, subscription, stripe_customer_id, stripe_subscription_id, subscription_quantity) values
			(:id, :name, %t, :subscription, :stripe_customer_id, :stripe_subscription_id, :subscription_quantity)`, k == "active"),
			t,
		)
		if err != nil {
			panic(err)
		}
	}

	for _, u := range testutil.Users {
		_, err = sqlx.NamedExec(
			tx,
			`insert into users (email, active, verified, customer_id, status, perms, password_hash, name) values
						(:email, :active, :verified, :customer_id, :status, :perms, 'blah', 'plasss')`,
			u,
		)
		if err != nil {
			panic(err)
		}
	}

	for _, u := range testutil.Invites {
		_, err = sqlx.NamedExec(
			tx,
			`insert into signups (email, claimed, perms, customer_id, activated, name) values
						(:email, :claimed, :perms, :customer_id, :activated, 'ploppy')`,
			u,
		)
		if err != nil {
			panic(err)
		}
	}

	// create a new team store with the transaction and give it to our test funcs
	testFun(NewTeamStore(tx))
}

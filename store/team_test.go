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
		team.Subscription = "team"
		team.SubscriptionQuantity = 5
		// hey, update the name while we're at it
		team.Name = "money"

		err = q.Update(team)
		assert.NoError(err)
		assert.Equal("money", team.Name)
		assert.Equal("team", team.Subscription)
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
		Subscription:         "beta",
		StripeCustomerId:     "cus_8oux3kULDWgU8F",
		StripeSubscriptionId: "sub_8owgXA5pkRDs31",
		SubscriptionQuantity: int32(3),
	}, true)
	insertTeam(&schema.Team{
		Id:                   "00000000-0000-0000-0000-000000000000",
		Name:                 "INACTIVE",
		Subscription:         "beta",
		StripeCustomerId:     "",
		StripeSubscriptionId: "",
		SubscriptionQuantity: int32(3),
	}, false)

	// insert users / signups
	insertUser := func(u *schema.User) {
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
	insertSignup := func(u *Signup) {
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

	insertUser(&schema.User{
		Email:      "opsee+active+admin@opsee.com",
		Active:     true,
		Verified:   true,
		CustomerId: "11111111-1111-1111-1111-111111111111",
		Status:     "active",
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
	})
	insertUser(&schema.User{
		Email:      "opsee+active+edit@opsee.com",
		Active:     true,
		Verified:   true,
		CustomerId: "11111111-1111-1111-1111-111111111111",
		Status:     "active",
		Perms:      &schema.UserFlags{Admin: false, Edit: true, Billing: false},
	})
	insertUser(&schema.User{
		Email:      "opsee+inactive@opsee.com",
		Active:     false,
		Verified:   true,
		CustomerId: "11111111-1111-1111-1111-111111111111",
		Status:     "active",
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
	})
	insertSignup(&Signup{
		Email:      "opsee+invited+admin+pending@opsee.com",
		Claimed:    false,
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
		CustomerId: "11111111-1111-1111-1111-111111111111",
	})
	insertSignup(&Signup{
		Email:      "opsee+invited+noperms+pending@opsee.com",
		Claimed:    false,
		Perms:      &schema.UserFlags{Admin: false, Edit: false, Billing: false},
		CustomerId: "11111111-1111-1111-1111-111111111111",
	})
	insertSignup(&Signup{
		Email:      "opsee+invited+admin+claimed@opsee.com",
		Claimed:    true,
		Perms:      &schema.UserFlags{Admin: true, Edit: true, Billing: true},
		CustomerId: "11111111-1111-1111-1111-111111111111",
	})

	// create a new team store with the transaction and give it to our test funcs
	testFun(NewTeamStore(tx))
}

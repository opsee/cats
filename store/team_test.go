package store

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/schema"
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
		assert.Equal("beta", team.SubscriptionPlan)
		assert.Equal("active", team.SubscriptionStatus)
		assert.EqualValues(3, team.SubscriptionQuantity)
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
		team.SubscriptionPlan = "team_monthly"
		team.SubscriptionQuantity = 5
		team.SubscriptionStatus = "trialing"
		// hey, update the name while we're at it
		team.Name = "money"

		err = q.Update(team)
		assert.NoError(err)
		assert.Equal("money", team.Name)
		assert.Equal("team_monthly", team.SubscriptionPlan)
		assert.EqualValues(5, team.SubscriptionQuantity)
		assert.Equal("trialing", team.SubscriptionStatus)
	})
}

func TestTeamCreate(t *testing.T) {
	assert := assert.New(t)

	withTeamFixtures(func(q TeamStore) {
		team := &schema.Team{
			Name:                 "etixx quickstep",
			SubscriptionPlan:     "free",
			SubscriptionQuantity: 0,
			SubscriptionStatus:   "active",
			StripeCustomerId:     "idk",
			StripeSubscriptionId: "iidk2",
		}
		err := q.Create(team)
		assert.Nil(err)
		assert.Equal("free", team.SubscriptionPlan)
		assert.EqualValues(0, team.SubscriptionQuantity)
		assert.Equal("active", team.SubscriptionStatus)
		assert.Equal("idk", team.StripeCustomerId)
		assert.Equal("iidk2", team.StripeSubscriptionId)
	})
}

func TestTeamDelete(t *testing.T) {
	assert := assert.New(t)

	withTeamFixtures(func(q TeamStore) {
		err := q.Delete(&schema.Team{Id: "11111111-1111-1111-1111-111111111111"})
		assert.NoError(err)

		t, err := q.Get("11111111-1111-1111-1111-111111111111")
		assert.NoError(err)
		assert.Nil(t)
	})
}

func TestTeamList(t *testing.T) {
	assert := assert.New(t)

	withTeamFixtures(func(q TeamStore) {
		teams, meta, err := q.List(1, 20)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(1, len(teams))
		assert.EqualValues(1, meta.Total)
		assert.EqualValues(1, meta.Page)
		assert.EqualValues(20, meta.PerPage)

		teams, meta, err = q.List(0, 20)
		assert.Error(err)

		teams, meta, err = q.List(1, 2000)
		assert.Error(err)
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
		rows, err := sqlx.NamedQuery(
			tx,
			`insert into subscriptions (plan, stripe_customer_id, stripe_subscription_id, quantity, status) values
			(:subscription_plan, :stripe_customer_id, :stripe_subscription_id, :subscription_quantity, :subscription_status) returning id`,
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

		_, err = sqlx.NamedExec(
			tx,
			fmt.Sprintf(`insert into customers (id, name, active, subscription_id) values
			(:id, :name, %t, %d)`, k == "active", sub.Id),
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
	testFun(&teamStore{tx, db})
}

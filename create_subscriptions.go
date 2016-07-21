package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/cats/store"
	"github.com/opsee/cats/subscriptions"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
)

func main() {
	viper.AutomaticEnv()
	stripe.Key = viper.GetString("stripe_key")

	db, err := sqlx.Open("postgres", viper.GetString("postgres_conn"))
	if err != nil {
		panic(err)
	}

	ts := store.NewTeamStore(db)
	id := viper.GetString("id")
	if id == "" {
		panic("no id provided")
	}

	email := viper.GetString("email")
	if email == "" {
		panic("no email provided")
	}

	team, err := ts.Get(id)
	if err != nil {
		panic(err)
	}

	team.SubscriptionPlan = "beta"
	team.SubscriptionQuantity = 0

	err = subscriptions.Create(team, email, "", time.Now().Add(4*24*time.Hour).Unix())
	if err != nil {
		panic(err)
	}

	err = ts.UpdateSubscription(team)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s updated\n", email)
}

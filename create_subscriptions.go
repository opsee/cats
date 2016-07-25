package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/opsee/basic/schema"
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
	cs := store.NewCheckStore(db)

	teams, meta, err := ts.List(1, 1000)
	if err != nil {
		panic(err)
	}

	if meta.Total > 1000 {
		panic("more than 1000 teams")
	}

	trialEnd, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Aug 1, 2016 at 8:00am (PDT)")
	if err != nil {
		panic(err)
	}

	for _, team := range teams {
		// once we're in the loop, let's try not to panic on error
		team, err := ts.Get(team.Id)
		if err != nil {
			fmt.Sprintf("\ncouldn't get team from database - skipping %s\n", team.Id)
			continue
		}

		// find the first admin for the team
		var admin *schema.User
		for _, u := range team.Users {
			if err := u.CheckActiveStatus(); err == nil {
				if u.HasPermission("admin") {
					admin = u
					break
				}
			}
		}

		// if there's a team with no admin, well that sucks, so skip it
		if admin == nil {
			fmt.Sprintf("\nno team admin for team - skipping %s\n", team.Id)
			continue
		}

		// only add to stripe if they're not already there
		if team.StripeCustomerId == "" || team.StripeSubscriptionId == "" {
			team.SubscriptionPlan = "beta"

			checkCount, err := cs.GetCheckCount(team.Id)
			if err != nil {
				fmt.Sprintf("\ncouldn't get check count - skipping %s\n", team.Id)
				continue
			}

			team.SubscriptionQuantity = checkCount

			err = subscriptions.Create(team, admin.Email, "", trialEnd.Unix())
			if err != nil {
				fmt.Sprintf("\ncouldn't create subscription - skipping %s\n", team.Id)
				continue
			}

			err = ts.UpdateSubscription(team)
			if err != nil {
				fmt.Sprintf("\ncouldn't update subscription - skipping %s\n", team.Id)
				continue
			}

			fmt.Sprint(".")
		}

		time.Sleep(500 * time.Millisecond)
	}
}

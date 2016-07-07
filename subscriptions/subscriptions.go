package subscriptions

import (
	"fmt"

	"github.com/opsee/basic/schema"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

// An opsee subscription plan
type Plan string

const (
	BetaPlan      Plan = "beta"
	DeveloperPlan      = "developer"
	TeamPlan           = "team"
)

var (
	Plans = []Plan{
		BetaPlan,
		DeveloperPlan,
		TeamPlan,
	}
)

// Validates that the plan is supported.
func (p Plan) Validate() error {
	for _, pp := range Plans {
		if p == pp {
			return nil
		}
	}

	return fmt.Errorf("%s is not a valid subscription plan", string(p))
}

// Create a new customer and subscription in stripe. The `user` parameter must be a user with
// billing permissions. `tokenSource` is an optional string token that represents a payment
// source. If supplied, it will be applied to the stripe customer as the payment source for the
// subscription.
func Create(team *schema.Team, user *schema.User, quantity uint64, tokenSource string) error {
	if err := user.CheckPermission("billing"); err != nil {
		return err
	}

	if err = Plan(team.Subscription).Validate(); err != nil {
		return err
	}

	params := &stripe.CustomerParams{
		Email: user.Email,
		Plan:  team.Subscription,
	}

	if source != "" {
		sp, err := stripe.SourceParamsFor(source)
		if err != nil {
			return err
		}

		err = params.SetSource(sp)
		if err != nil {
			return err
		}
	}

	_, err := customer.New(params)
	return err
}

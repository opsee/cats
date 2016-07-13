package subscriptions

import (
	"fmt"

	"github.com/opsee/basic/schema"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
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

	if err := Plan(team.Subscription).Validate(); err != nil {
		return err
	}

	params := &stripe.CustomerParams{
		Email:    user.Email,
		Plan:     team.Subscription,
		Quantity: quantity,
	}

	if tokenSource != "" {
		sp, err := stripe.SourceParamsFor(tokenSource)
		if err != nil {
			return err
		}

		err = params.SetSource(sp)
		if err != nil {
			return err
		}
	}

	params.AddMeta("customer-id", team.Id)
	params.AddMeta("team-name", team.Name)

	response, err := customer.New(params)
	if err != nil {
		return err
	}

	team.StripeCustomerId = response.ID
	if response.Subs == nil {
		return fmt.Errorf("empty subscription list from stripe")
	}

	if len(response.Subs.Values) < 1 {
		return fmt.Errorf("empty subscription list from stripe")
	}

	subscription := response.Subs.Values[0]
	team.StripeSubscriptionId = subscription.ID

	return nil
}

// Update a customer's subscription in stripe
func Update(team *schema.Team, user *schema.User, quantity uint64, tokenSource string) error {
	if err := user.CheckPermission("billing"); err != nil {
		return err
	}

	if err := Plan(team.Subscription).Validate(); err != nil {
		return err
	}

	if team.StripeCustomerId == "" {
		return fmt.Errorf("team missing stripe customer_id")
	}

	if team.StripeSubscriptionId == "" {
		return fmt.Errorf("team missing stripe subscription_id")
	}

	params := &stripe.SubParams{
		Customer: team.StripeCustomerId,
		Plan:     team.Subscription,
		Quantity: quantity,
	}

	if tokenSource != "" {
		params.Token = tokenSource
	}

	if _, err := sub.Update(team.StripeSubscriptionId, params); err != nil {
		return err
	}

	return nil
}

// Cancel a customer's subscription in stripe
func Cancel(team *schema.Team, user *schema.User) error {
	if err := user.CheckPermission("billing"); err != nil {
		return err
	}

	if team.StripeSubscriptionId == "" {
		return fmt.Errorf("team missing stripe subscription_id")
	}

	if _, err := sub.Cancel(team.StripeSubscriptionId, nil); err != nil {
		return err
	}

	return nil
}

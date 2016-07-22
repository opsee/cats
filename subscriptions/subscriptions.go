package subscriptions

import (
	"fmt"

	"github.com/opsee/basic/schema"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/sub"
)

// An opsee subscription plan
type Plan string

const (
	FreePlan      Plan = "free"
	BetaPlan           = "beta"
	DeveloperPlan      = "developer_monthly"
	TeamPlan           = "team_monthly"
)

var (
	Plans = []Plan{
		FreePlan,
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

// Fetches subscription, payment source, and invoice info and populates the team struct
func Get(team *schema.Team) error {
	if team.StripeCustomerId == "" {
		return fmt.Errorf("team missing stripe customer_id")
	}

	if team.StripeSubscriptionId == "" {
		return fmt.Errorf("team missing stripe subscription_id")
	}

	subby, err := sub.Get(team.StripeSubscriptionId, nil)
	if err != nil {
		return err
	}

	team.SubscriptionTrialStart = opsee_types.NewTimestamp(subby.TrialStart)
	team.SubscriptionTrialEnd = opsee_types.NewTimestamp(subby.TrialEnd)
	if subby.Plan != nil {
		team.SubscriptionPlanAmount = int32(subby.Plan.Amount)
	}

	// credit card info
	if subby.Customer != nil && subby.Customer.Sources != nil && len(subby.Customer.Sources.Values) > 0 {
		source := subby.Customer.Sources.Values[0]

		if source.Card != nil {
			team.CreditCardInfo = &schema.CreditCardInfo{
				Name:     source.Card.Name,
				Last4:    source.Card.LastFour,
				ExpMonth: int32(source.Card.Month),
				ExpYear:  int32(source.Card.Year),
				Brand:    string(source.Card.Brand),
			}
		}

	}

	// upcoming invoice
	nextInv, err := invoice.GetNext(&stripe.InvoiceParams{
		Customer: team.StripeCustomerId,
	})
	if err != nil {
		return err
	}
	team.NextInvoice = copyInvoice(nextInv)

	// all invoices
	var invoices []*schema.Invoice
	invoiceIter := invoice.List(&stripe.InvoiceListParams{
		Customer: team.StripeCustomerId,
	})
	for invoiceIter.Next() {
		invoices = append(invoices, copyInvoice(invoiceIter.Invoice()))
	}
	team.Invoices = invoices

	return nil
}

// Create a new customer and subscription in stripe. The `user` parameter must be a user with
// billing permissions. `tokenSource` is an optional string token that represents a payment
// source. If supplied, it will be applied to the stripe customer as the payment source for the
// subscription.
func Create(team *schema.Team, email string, tokenSource string, trialEnd int64) error {
	if err := Plan(team.SubscriptionPlan).Validate(); err != nil {
		return err
	}

	params := &stripe.CustomerParams{
		Email:    email,
		Plan:     team.SubscriptionPlan,
		Quantity: uint64(team.SubscriptionQuantity),
	}

	if trialEnd != 0 {
		params.TrialEnd = trialEnd
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
	team.SubscriptionStatus = string(subscription.Status)

	return nil
}

// Update a customer's subscription in stripe
func Update(team *schema.Team, tokenSource string) error {
	if err := Plan(team.SubscriptionPlan).Validate(); err != nil {
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
		Plan:     team.SubscriptionPlan,
		Quantity: uint64(team.SubscriptionQuantity),
	}

	if tokenSource != "" {
		params.Token = tokenSource
	}

	_, err := sub.Update(team.StripeSubscriptionId, params)
	if err != nil {
		return err
	}

	return nil
}

// Cancel a customer's subscription in stripe
func Cancel(team *schema.Team) error {
	if team.StripeSubscriptionId == "" {
		return fmt.Errorf("team missing stripe subscription_id")
	}

	if _, err := sub.Cancel(team.StripeSubscriptionId, nil); err != nil {
		return err
	}

	return nil
}

func copyInvoice(stripeInvoice *stripe.Invoice) *schema.Invoice {
	invoice := &schema.Invoice{}
	invoice.Date = opsee_types.NewTimestamp(stripeInvoice.Date)
	invoice.Amount = int32(stripeInvoice.Total)
	invoice.Paid = stripeInvoice.Paid
	return invoice
}

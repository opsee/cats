package subscriptions

import (
	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/mailer"
	log "github.com/opsee/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
)

func HandleEvent(team *schema.Team, event *stripe.Event) error {
	log.Infof("subscriptions.HandleEvent: %#v -- %#v", event, event.Data.Obj)

	switch event.Type {
	case "customer.subscription.trial_will_end":
		// only send this to non-free plan people
		if team.SubscriptionPlan == "free" {
			return nil
		}

		// if they've added a payment source, then no worries, don't send an email
		if team.CreditCardInfo != nil {
			return nil
		}

		mailBillingUsers(team, "warning-minus-three", map[string]interface{}{})

	case "invoice.payment_failed":
		// only send this to non-free plan people
		// no idea why this would happen
		if team.SubscriptionPlan == "free" {
			return nil
		}

		// if we're trialing and payment fails, then the trial must have expired
		// so we'll send them a special email
		if team.SubscriptionStatus == string(sub.Trialing) {
			mailBillingUsers(team, "trial-expired", map[string]interface{}{})
			team.SubscriptionStatus = string(sub.PastDue)
			return nil
		}

		// this is just a regular payment failure
		switch event.GetObjValue("attempt_count") {
		case "1":
			mailBillingUsers(team, "warning-zero", map[string]interface{}{})
		case "2":
			mailBillingUsers(team, "warning-three", map[string]interface{}{})
		case "3":
			mailBillingUsers(team, "warning-seven", map[string]interface{}{})
		}

		return nil

	case "invoice.payment_succeeded":
		// make sure our db is reflecting the right status if they have a cc
		if team.CreditCardInfo != nil {
			team.SubscriptionStatus = string(sub.Active)
		}
	}

	return nil
}

func mailBillingUsers(team *schema.Team, template string, vars map[string]interface{}) {
	withBillingUsers(team, func(u *schema.User) {
		logger := log.WithFields(log.Fields{"template": template, "email": u.Email})
		_, err := mailer.Send(u.Email, u.Name, template, vars)
		if err != nil {
			logger.WithError(err).Error("couldn't send email to mandrill")
		}

		logger.Info("sent email")
	})
}

func withBillingUsers(team *schema.Team, billFunc func(*schema.User)) {
	for _, u := range team.Users {
		if u.HasPermission("admin") || u.HasPermission("billing") {
			billFunc(u)
		}
	}
}

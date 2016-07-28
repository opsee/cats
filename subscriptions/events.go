package subscriptions

import (
	"github.com/opsee/basic/schema"
	"github.com/opsee/cats/mailer"
	log "github.com/opsee/logrus"
	"github.com/stripe/stripe-go"
)

func HandleEvent(team *schema.Team, event *stripe.Event) error {
	log.Infof("subscriptions.HandleEvent: %#v -- %#v", event, event.Data.Obj)

	switch event.Type {
	case "customer.subscription.trial_will_end":
		for _, u := range team.Users {
			if u.HasPermission("admin") || u.HasPermission("billing") {
				_, err := mailer.Send(u.Email, u.Name, "warning-minus-three", map[string]interface{}{})
				if err != nil {
					log.WithError(err).Error("couldn't send email to mandrill")
				}
			}
		}
	}

	return nil
}

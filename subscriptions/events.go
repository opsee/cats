package subscriptions

import (
	// "github.com/opsee/cats/mailer"
	log "github.com/opsee/logrus"
	"github.com/stripe/stripe-go"
)

func HandleEvent(event *stripe.Event) error {
	log.Infof("subscriptions.HandleEvent: %#v -- %#v", event, *event.Data.Obj)

	switch event.Type {
	case "customer.subscription.trial_will_end":
		
	}
	
	return nil
}

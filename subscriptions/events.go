package subscriptions

import (
	// "github.com/opsee/cats/mailer"
	log "github.com/opsee/logrus"
	"github.com/stripe/stripe-go"
)

func HandleEvent(event *stripe.Event) error {
	log.Infof("subscriptions.HandleEvent: %s", event.Type)

	// mailer.Send()
	return nil
}

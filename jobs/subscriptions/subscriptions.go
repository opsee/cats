package subscriptions

import (
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/subscriptions"
	"github.com/opsee/gmunch"
	log "github.com/opsee/logrus"
	"github.com/stripe/stripe-go"
	"golang.org/x/net/context"
)

type Job struct {
	event   *gmunch.Event
	context context.Context
	service opsee.CatsServer
}

func New(service opsee.CatsServer, evt *gmunch.Event) *Job {
	return &Job{
		event:   evt,
		context: context.Background(),
		service: service,
	}
}

func (j *Job) Context() context.Context {
	return j.context
}

func (j *Job) Execute() (interface{}, error) {
	log.Infof("job: %s", j.event.Name)

	event := &stripe.Event{}
	if err := j.event.Decoder().Decode(event); err != nil {
		log.WithError(err).Errorf("couldn't decode gmunch event: %#v", j.event)
		return nil, err
	}

	stripeCustomerId := event.GetObjValue("customer")
	if stripeCustomerId == "" {
		log.Info("no customer id, skipping stripe event")
		return nil, nil
	}

	if !event.Live {
		// for debugging, this is 'computer@markmart.in'
		stripeCustomerId = "cus_8szeJAcdhSXmUY"
	}

	teamResponse, err := j.service.GetTeam(context.Background(), &opsee.GetTeamRequest{
		Team: &schema.Team{
			StripeCustomerId: stripeCustomerId,
		},
	})

	if err != nil {
		log.WithError(err).Error("couldn't get team")
		return nil, err
	}

	if err := subscriptions.HandleEvent(teamResponse.Team, event); err != nil {
		log.WithError(err).Error("couldn't handle subscription event")
		return nil, err
	}

	return struct{}{}, nil
}

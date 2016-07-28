package subscriptions

import (
	"github.com/opsee/cats/subscriptions"
	"github.com/opsee/gmunch"
	log "github.com/opsee/logrus"
	"github.com/stripe/stripe-go"
	"golang.org/x/net/context"
)

type Job struct {
	event   *gmunch.Event
	context context.Context
}

func New(evt *gmunch.Event) *Job {
	return &Job{
		event:   evt,
		context: context.Background(),
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
	
	stripeCustomerId := e.GetObjValue("customer")
	if stripeCustomerId == "" {
		
	}

	if err := subscriptions.HandleEvent(event); err != nil {
		return nil, err
	}

	return struct{}{}, nil
}

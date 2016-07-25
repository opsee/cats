package service

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/opsee/basic/tp"
	log "github.com/opsee/logrus"
	"github.com/stripe/stripe-go"
	"golang.org/x/net/context"
)

var (
	StripeWebhookPassword string
)

const (
	stripeEventKey = iota
)

func (s *service) stripeHookDecoder() tp.DecodeFunc {
	return func(ctx context.Context, rw http.ResponseWriter, r *http.Request, p httprouter.Params) (context.Context, int, error) {
		_, password, ok := r.BasicAuth()
		if !ok {
			log.Error("failed to decode basic auth")
			return ctx, http.StatusUnauthorized, fmt.Errorf("authentication failed")
		}

		if subtle.ConstantTimeEq(int32(len(StripeWebhookPassword)), int32(len(password))) != 1 {
			log.Error("password length comparison failed")
			return ctx, http.StatusUnauthorized, fmt.Errorf("authentication failed")
		}

		if subtle.ConstantTimeCompare([]byte(StripeWebhookPassword), []byte(password)) != 1 {
			log.Error("password comparison failed")
			return ctx, http.StatusUnauthorized, fmt.Errorf("authentication failed")
		}

		event := &stripe.Event{}
		if err := json.NewDecoder(r.Body).Decode(event); err != nil {
			log.WithError(err).Error("failed to decode request")
			return ctx, http.StatusBadRequest, err
		}

		return context.WithValue(ctx, stripeEventKey, event), 0, nil
	}
}

func (s *service) stripeHookHandler() tp.HandleFunc {
	return func(ctx context.Context) (interface{}, int, error) {
		event, ok := ctx.Value(stripeEventKey).(*stripe.Event)
		if !ok {
			log.Error("can't decode stripe event from context")
			return nil, http.StatusBadRequest, fmt.Errorf("can't decode stripe event from context")
		}

		log.Infof("got stripe event: %#v", event)

		return map[string]bool{"ok": true}, http.StatusOK, nil
	}
}

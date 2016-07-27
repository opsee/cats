package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go"
)

func TestStripeHook(t *testing.T) {
	assert := assert.New(t)
	event := &stripe.Event{
		ID:     "111",
		Live:   true,
		Type:   "account.created",
		UserID: "111",
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "https://cats/hooks/stripe", bytes.NewBuffer(eventJSON))
	if err != nil {
		t.Fatal(err)
	}

	// testing auth
	StripeWebhookPassword = "kitties"
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("stripe:kitties"))))

	w := httptest.NewRecorder()
	s := &service{
		sluiceClient: &testSluiceClient{},
	}
	h := s.NewHandler()

	h.ServeHTTP(w, req)
	assert.Equal(200, w.Code)

	req, err = http.NewRequest("POST", "https://cats/hooks/stripe", bytes.NewBuffer(eventJSON))
	if err != nil {
		t.Fatal(err)
	}

	// testing baaaaad auth
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("stripe:BADPASS"))))
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(401, w.Code)
}

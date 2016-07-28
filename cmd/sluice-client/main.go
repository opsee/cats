package main

import (
	"crypto/tls"

	"github.com/opsee/gmunch/client"
	"github.com/stripe/stripe-go"
)

func main() {
	config := client.Config{
		TLSConfig: tls.Config{},
	}
	client, err := client.New("sluice.in.opsee.com:8443", config)
	if err != nil {
		panic(err)
	}

	err = client.Send("stripe_hook", &stripe.Event{
		Type: "hellow",
	})

	if err != nil {
		panic(err)
	}
}

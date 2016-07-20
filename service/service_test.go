package service

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
)

func TestMain(m *testing.M) {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	// initialize stripe with our account key
	stripe.Key = viper.GetString("stripe_key")

	os.Exit(m.Run())
}

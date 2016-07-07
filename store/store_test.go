package store

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	os.Exit(m.Run())
}

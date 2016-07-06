package main

import (
	"github.com/opsee/cats/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	svc, err := service.New(viper.GetString("postgres_conn"))
	if err != nil {
		log.WithError(err).Fatal("Unable to start service.")
	}

	log.WithError(svc.StartHTTP(viper.GetString("address"))).Fatal("Error in listener.")
}

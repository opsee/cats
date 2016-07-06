package main

import (
	"github.com/opsee/cats/service"
	log "github.com/opsee/logrus"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	svc, err := service.New(viper.GetString("postgres_conn"))
	if err != nil {
		log.WithError(err).Fatal("Unable to start service.")
	}

	log.WithError(svc.StartMux(
		viper.GetString("address"),
		viper.GetString("cert"),
		viper.GetString("cert_key"),
	)).Fatal("Error in listener.")
}

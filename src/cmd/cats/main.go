package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/opsee/cats/service"
	log "github.com/sirupsen/logrus"
)

type config struct {
	PostgresConn string `envconfig:"postgres_conn"`
	ListenAddr   string `envconfig:"listen_addr"`
}

func main() {
	var cfg config
	err := envconfig.Process("cats", &cfg)
	if err != nil {
		log.WithError(err).Fatal("Unable to process environment config.")
	}

	svc, err := service.NewService(cfg.PostgresConn)
	if err != nil {
		log.WithError(err).Fatal("Unable to start service.")
	}

	log.WithError(svc.StartHTTP(cfg.ListenAddr)).Fatal("Error in listener.")
}

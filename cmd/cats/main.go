package main

import (
	"io/ioutil"

	"github.com/opsee/cats/service"
	"github.com/opsee/cats/servicer"
	log "github.com/opsee/logrus"
	"github.com/opsee/vaper"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	// initialize vaper with our secret key
	keyPath := viper.GetString("vape_keyfile")
	if keyPath == "" {
		log.Fatal("Must set CATS_VAPE_KEYFILE environment variable.")
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Println("Unable to read keyfile:", keyPath)
		log.Fatal(err)
	}
	vaper.Init(key)

	// initialize the user servicer TODO(mark): deprecate this
	servicer.Init(servicer.Config{
		Host:        viper.GetString("opsee_host"),
		MandrillKey: viper.GetString("mandrill_key"),
		IntercomKey: viper.GetString("intercom_key"),
		CloseIOKey:  viper.GetString("closeio_key"),
		SlackUrl:    viper.GetString("slack_url"),
	})

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

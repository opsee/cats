package main

import (
	"os"
	"os/signal"
	"syscall"

	// "github.com/opsee/cats/jobs/slack"
	"github.com/keighl/mandrill"
	"github.com/opsee/cats/jobs/subscriptions"
	"github.com/opsee/cats/mailer"
	"github.com/opsee/gmunch"
	consumer "github.com/opsee/gmunch/consumer/kinesis"
	producer "github.com/opsee/gmunch/producer/kinesis"
	"github.com/opsee/gmunch/server"
	"github.com/opsee/gmunch/worker"
	log "github.com/opsee/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
)

func main() {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	mailer.Client = mandrill.ClientWithKey(viper.GetString("mandrill_key"))
	mailer.BaseURL = viper.GetString("opsee_host")

	stripe.Key = viper.GetString("stripe_key")

	server := server.New(server.Config{
		LogLevel: viper.GetString("log_level"),
		Producer: producer.New(producer.Config{
			Stream: viper.GetString("kinesis_stream"),
		}),
		Consumer: consumer.New(consumer.Config{
			Stream:        viper.GetString("kinesis_stream"),
			EtcdEndpoints: viper.GetStringSlice("etcd_address"),
			ShardPath:     viper.GetString("shard_path"),
		}),
		Dispatch: worker.Dispatch{
			"stripe_hook": func(evt *gmunch.Event) []worker.Task {
				return []worker.Task{subscriptions.New(evt)}
			},
			// "slack_notification": func(evt *gmunch.Event) []worker.Task {
			// 	return []worker.Task{slack.New(evt)}
			// },
		},
	})

	sigChan := make(chan os.Signal, 1)
	errChan := make(chan error)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		errChan <- server.Start(
			viper.GetString("sluice_address"),
			viper.GetString("cert"),
			viper.GetString("cert_key"),
		)
	}()

	var err error
	select {
	case err = <-errChan:
		log.Info("received error from grpc service")
	case <-sigChan:
		log.Info("received interrupt")
	}

	server.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

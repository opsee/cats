package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/keighl/mandrill"
	newrelic "github.com/newrelic/go-agent"
	"github.com/opsee/cats/checks/results"
	"github.com/opsee/cats/jobs/subscriptions"
	"github.com/opsee/cats/mailer"
	"github.com/opsee/cats/service"
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

	awsSession := session.New(&aws.Config{Region: aws.String("us-west-2")})
	s3Store := &results.S3Store{
		S3Client:   s3.New(awsSession),
		BucketName: viper.GetString("results_s3_bucket"),
	}

	agentConfig := newrelic.NewConfig("Cats", viper.GetString("newrelic_key"))
	agentConfig.BetaToken = viper.GetString("newrelic_beta_token")
	agent, err := newrelic.NewApplication(agentConfig)
	if err != nil {
		log.WithError(err).Fatal("Unable to start service.")
	}

	catsSvc, err := service.New(viper.GetString("postgres_conn"), s3Store, agent)
	if err != nil {
		log.WithError(err).Fatal("Can't create cats service")
	}

	server := server.New(server.Config{
		LogLevel: viper.GetString("log_level"),
		Producer: producer.New(producer.Config{
			Stream: viper.GetString("kinesis_stream"),
			Region: "us-west-2",
		}),
		Consumer: consumer.New(consumer.Config{
			Stream:        viper.GetString("kinesis_stream"),
			EtcdEndpoints: viper.GetStringSlice("etcd_address"),
			ShardPath:     viper.GetString("shard_path"),
			Region:        "us-west-2",
		}),
		Dispatch: worker.Dispatch{
			"stripe_hook": func(evt *gmunch.Event) []worker.Task {
				return []worker.Task{subscriptions.New(catsSvc, evt)}
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

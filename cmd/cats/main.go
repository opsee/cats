package main

import (
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/opsee/cats/checks/results"
	"github.com/opsee/cats/service"
	"github.com/opsee/cats/servicer"
	vapestore "github.com/opsee/cats/servicer/store"
	log "github.com/opsee/logrus"
	"github.com/opsee/vaper"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
)

func main() {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	// initialize stripe with our account key, and our service with our webhook password
	stripe.Key = viper.GetString("stripe_key")
	service.StripeWebhookPassword = viper.GetString("stripe_webhook_password")

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

	// initialize the user servicer and store TODO(mark): deprecate this
	vapestore.Init(viper.GetString("postgres_conn"))
	servicer.Init(servicer.Config{
		Host:        viper.GetString("opsee_host"),
		MandrillKey: viper.GetString("mandrill_key"),
		IntercomKey: viper.GetString("intercom_key"),
		CloseIOKey:  viper.GetString("closeio_key"),
		SlackUrl:    viper.GetString("slack_url"),
	})

	//resultStore := &results.DynamoStore{dynamodb.New(session.New(aws.NewConfig().WithRegion("us-west-2")))}
	resultStore := &results.S3Store{
		BucketName: viper.GetString("results_s3_bucket"),
		S3Client:   s3.New(session.New(aws.NewConfig().WithRegion("us-west-2"))),
	}

	svc, err := service.New(viper.GetString("postgres_conn"), resultStore)
	if err != nil {
		log.WithError(err).Fatal("Unable to start service.")
	}

	log.WithError(svc.StartMux(
		viper.GetString("address"),
		viper.GetString("cert"),
		viper.GetString("cert_key"),
	)).Fatal("Error in listener.")
}

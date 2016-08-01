package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	newrelic "github.com/newrelic/go-agent"
	"github.com/nsqio/go-nsq"
	"github.com/opsee/basic/schema"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/cats/checks"
	"github.com/opsee/cats/checks/results"
	"github.com/opsee/cats/checks/worker"
	"github.com/opsee/cats/service"
	"github.com/opsee/cats/store"
	log "github.com/opsee/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

var (
	checkResultsHandled = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "check_results_handled",
		Help: "Total number of check results processed.",
	})
)

func init() {
	prometheus.MustRegister(checkResultsHandled)
}

func main() {
	viper.SetEnvPrefix("cats")
	viper.AutomaticEnv()

	viper.SetDefault("log_level", "info")
	logLevelStr := viper.GetString("log_level")
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		log.WithError(err).Error("Could not parse log level, using default.")
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	go func() {
		hostname, err := os.Hostname()
		if err != nil {
			log.WithError(err).Error("Error getting hostname.")
			return
		}

		ticker := time.Tick(5 * time.Second)
		for {
			<-ticker
			err = prometheus.Push("pracovnik", hostname, "172.30.35.35:9091")
			if err != nil {
				log.WithError(err).Error("Error pushing to pushgateway.")
			}
		}
	}()

	nsqConfig := nsq.NewConfig()
	nsqConfig.MaxInFlight = 4

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	maxTasks := viper.GetInt("max_tasks")

	consumer, err := worker.NewConsumer(&worker.ConsumerConfig{
		Topic:            "_.results",
		Channel:          "dynamo-results-worker",
		LookupdAddresses: viper.GetStringSlice("nsqlookupd_addrs"),
		NSQConfig:        nsqConfig,
		HandlerCount:     maxTasks,
	})

	if err != nil {
		log.WithError(err).Fatal("Failed to create consumer.")
	}

	nsqdHost := viper.GetString("nsqd_host")
	producer, err := nsq.NewProducer(nsqdHost, nsqConfig)
	if err != nil {
		log.WithError(err).Fatal("Failed to create producer")
	}

	db, err := sqlx.Open("postgres", viper.GetString("postgres_conn"))
	if err != nil {
		log.WithError(err).Fatal("Cannot connect to database.")
	}

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

	consumer.AddHandler(func(msg *nsq.Message) error {
		result := &schema.CheckResult{}
		if err := proto.Unmarshal(msg.Body, result); err != nil {
			log.WithError(err).Error("Error unmarshalling message from NSQ.")
			return err
		}

		logger := log.WithFields(log.Fields{
			"customer_id": result.CustomerId,
			"check_id":    result.CheckId,
			"bastion_id":  result.BastionId,
		})

		// TODO(greg): CheckResult objects should probably have a validator.
		if result.CustomerId == "" || result.CheckId == "" {
			logger.Error("Received invalid check result.")
			return nil
		}

		task := worker.NewCheckWorker(db, s3Store, result)
		_, err = task.Execute()
		if err != nil {
			logger.WithError(err).Error("Error executing task.")
			return err
		}

		checkResultsHandled.Inc()
		return nil
	})

	checks.AddHook(func(newStateID checks.StateId, state *checks.State, result *schema.CheckResult) {
		logger := log.WithFields(log.Fields{
			"customer_id":           state.CustomerId,
			"check_id":              state.CheckId,
			"min_failing_count":     state.MinFailingCount,
			"min_failing_time":      state.MinFailingTime,
			"failing_count":         state.FailingCount,
			"failing_time_s":        state.TimeInState().Seconds(),
			"old_state":             state.Id.String(),
			"new_state":             newStateID.String(),
			"bastion_id":            result.BastionId,
			"result.response_count": len(result.Responses),
			"result.passing":        result.Passing,
			"result.failing_count":  result.FailingCount(),
			"result.timestamp":      result.Timestamp.String(),
		})

		checkStore := store.NewCheckStore(db)

		logEntry, err := checkStore.CreateStateTransitionLogEntry(state.CheckId, state.CustomerId, state.Id, newStateID)
		if err != nil {
			logger.WithError(err).Error("Error creating StateTransitionLogEntry")
		}

		logger.Infof("Created StateTransitionLogEntry: %d", logEntry.Id)

		resultsResp, err := catsSvc.GetCheckResults(context.Background(), &opsee.GetCheckResultsRequest{
			CustomerId: state.CustomerId,
			CheckId:    state.CheckId,
		})

		if err != nil {
			logger.WithError(err).Error("Error getting results for check")
			return
		}
		results := resultsResp.Results
		// What we just got from the check store has the old result object for this
		// bastion. Replace it with the result from the transition.
		for i, r := range results {
			if result.BastionId == r.BastionId {
				results[i] = result
			}
		}

		check, err := checkStore.GetCheck(&schema.User{CustomerId: state.CustomerId}, state.CheckId)
		if err != nil {
			logger.WithError(err).Error("Error getting check from db: ", state.CheckId)
			return
		}
		check.Results = results
		check.State = state.State
		check.FailingCount = state.FailingCount
		check.ResponseCount = state.ResponseCount

		err = s3Store.PutCheckSnapshot(logEntry.Id, check)
		if err != nil {
			logger.WithError(err).Error("Error putting transition snapshot to s3")
			return
		}

		if (state.Id == checks.StateFailWait && newStateID == checks.StateFail) ||
			(state.Id == checks.StatePassWait && newStateID == checks.StateOK) ||
			(state.Id == checks.StatePassWait && newStateID == checks.StateWarn) {
			logger.Info("Sending alert.")

			var alertResult *schema.CheckResult
			// Pick the first failing result or the first passing.
			if newStateID == checks.StateFail {
				for _, r := range results {
					if !r.Passing {
						alertResult = r
						break
					}
				}
			} else {
				for _, r := range results {
					if r.Passing {
						alertResult = r
						break
					}
				}
			}
			if alertResult == nil {
				logger.Error("Could not find an appropriate result to send to alert.")
				return
			}

			resultBytes, err := proto.Marshal(alertResult)
			if err != nil {
				logger.WithError(err).Error("Unable to marshal Check to protobuf")
			}
			if err := producer.Publish("alerts", resultBytes); err != nil {
				logger.WithError(err).Error("Error publishing alert to NSQ.")
			}
		}
	})

	if err := consumer.Start(); err != nil {
		log.WithError(err).Fatal("Failed to start consumer.")
	}

	<-sigChan

	consumer.Stop()
}

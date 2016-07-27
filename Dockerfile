FROM quay.io/opsee/vinz:latest

ENV CATS_POSTGRES_CONN=""
ENV CATS_ADDRESS=""
ENV CATS_CERT="cert.pem"
ENV CATS_CERT_KEY="key.pem"
ENV CATS_ETCD_ADDRESS=""
ENV CATS_NSQLOOKUPD_ADDRS=""
ENV CATS_MAX_TASKS=""
ENV CATS_ALERTS_SQS_URL=""
ENV CATS_NSQD_HOST=""
ENV CATS_VAPE_KEYFILE="/vape.test.key"
ENV CATS_MANDRILL_KEY=""
ENV CATS_OPSEE_HOST=""
ENV CATS_INTERCOM_KEY=""
ENV CATS_CLOSEIO_KEY=""
ENV CATS_SLACK_URL=""
ENV CATS_STRIPE_KEY=""
ENV CATS_STRIPE_WEBHOOK_PASSWORD=""
ENV CATS_RESULTS_S3_BUCKET=""
ENV CATS_NEWRELIC_KEY=""
ENV CATS_NEWRELIC_BETA_TOKEN=""
ENV CATS_LOG_LEVEL=""
ENV CATS_KINESIS_STREAM=""
ENV CATS_SHARD_PATH=""
ENV CATS_SLUICE_ADDRESS=""

RUN apk add --update bash ca-certificates curl
RUN curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate

COPY target/linux/amd64/bin/* /
COPY run.sh /run.sh
COPY migrations /migrations
COPY cert.pem /
COPY key.pem /
COPY vape.test.key /

EXPOSE 9101 9107

CMD ["/cats"]

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

RUN apk add --update bash ca-certificates curl
RUN curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate

COPY target/linux/amd64/bin/* /
COPY run.sh /run.sh
COPY migrations /migrations
COPY cert.pem /
COPY key.pem /

EXPOSE 9101

CMD ["/cats"]

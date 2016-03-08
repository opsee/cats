FROM quay.io/opsee/vinz:latest

ENV CATS_POSTGRES_CONN=""
ENV CATS_ADDRESS=""

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

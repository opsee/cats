FROM quay.io/opsee/vinz:latest

ENV CATS_POSTGRES_CONN=""
ENV CATS_LISTEN_ADDR=""

RUN apk add --update bash ca-certificates curl
RUN curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate
RUN curl -Lo /opt/bin/ec2-env https://s3-us-west-2.amazonaws.com/opsee-releases/go/ec2-env/ec2-env && \
    chmod 755 /opt/bin/ec2-env

COPY target/linux/amd64/bin/* /
COPY run.sh /run.sh
COPY migrations /migrations

EXPOSE 9096

ENTRYPOINT ["/run.sh"]

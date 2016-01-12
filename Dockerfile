FROM gliderlabs/alpine:3.2

ENV CATS_POSTGRES_CONN=""
ENV CATS_LISTEN_ADDR=""

ADD bin/cats-linux-amd64 /cats
ENTRYPOINT ["/cats"]

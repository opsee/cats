#!/bin/bash
set -e

CMD=${1:-cats}
APPENV=${APPENV:-catsenv}

/opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$APPENV > /$APPENV

source /$APPENV && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/vape.key > /vape.key && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$CATS_CERT > /$CATS_CERT && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$CATS_CERT_KEY > /$CATS_CERT_KEY && \
  chmod 600 /$CATS_CERT_KEY && \
	/opt/bin/migrate -url "$CATS_POSTGRES_CONN" -path /migrations up && \
	/${CMD}

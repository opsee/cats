#!/bin/bash
set -e

eval "$(/usr/bin/env go env)"

migrate -url $CATS_POSTGRES_CONN -path ./migrations up
echo "loading fixtures" | psql -f fixtures.sql $CATS_POSTGRES_CONN

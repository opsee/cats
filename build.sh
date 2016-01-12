#!/bin/bash
set -e

echo "loading schema for tests..."
echo "drop database cats_test" | psql -U postgres -h postgres
echo "create database cats_test" | psql -U postgres -h postgres
migrate -url $CATS_POSTGRES_CONN -path ./migrations up
echo "loading fixtures" | psql -U postgres -h postgres -f fixtures.sql cats_test

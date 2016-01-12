#!/bin/bash
set -e

eval "$(/usr/bin/env go env)"

echo "loading schema for tests..."
echo "drop database cats_test" | psql -U postgres -h postgres
echo "create database cats_test" | psql -U postgres -h postgres
migrate -url $CATS_POSTGRES_CONN -path ./migrations up
echo "loading fixtures" | psql -U postgres -h postgres -f fixtures.sql cats_test

PROTO_DIR=bastion_proto
CHECKER_PROTO_GO=checker.pb.go
PROTO_SRC=src/github.com/opsee/cats/checker
CHECKER_PROTO_TARGET=${PROTO_SRC}/${CHECKER_PROTO_GO}

protoc --go_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. ${PROTO_DIR}/checker.proto

cp ${PROTO_DIR}/checker.pb.go ${CHECKER_PROTO_TARGET}

PROTO_DIR=bastion_proto
CHECKER_PROTO_GO=checker.pb.go
PROTO_SRC=src/github.com/opsee/cats/checker
CHECKER_PROTO_TARGET=$(PROTO_SRC)/$(CHECKER_PROTO_GO)

build: fmt test
	gb build

fmt:
	@gofmt -w src/

test: $(CHECKER_PROTO_TARGET)
	gb test

$(CHECKER_PROTO_TARGET): proto $(PROTO_SRC)
	cp $(PROTO_DIR)/checker.pb.go $(CHECKER_PROTO_TARGET) 

$(PROTO_SRC):
	mkdir -p $(PROTO_SRC)

proto: $(PROTO_DIR)/checker.proto
	protoc --go_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. $(PROTO_DIR)/checker.proto

migrate:
	migrate -url $(POSTGRES_CONN) -path ./migrations up

clean:
	rm -rf pkg/

.PHONY: clean migrate

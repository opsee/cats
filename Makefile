APPENV := testenv
REV ?= latest

build: deps fmt $(APPENV)
	docker run \
		--link cats_postgres_1:postgres \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-v `pwd`:/build \
		quay.io/opsee/build-go:go15
	docker build -t quay.io/opsee/cats:$(REV) .

deps:
	docker-compose up -d

fmt:
	@gofmt -w src/

migrate:
	migrate -url $(CATS_POSTGRES_CONN) -path ./migrations up

clean:
	rm -rf pkg/

.PHONY: clean migrate

APPENV := testenv

build: fmt $(APPENV)
	docker run \
		--link postgres:postgres \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-v `pwd`:/build \
		quay.io/opsee/build-go:go15
	docker build -t quay.io/opsee/cats:latest .

fmt:
	@gofmt -w src/

migrate:
	migrate -url $(CATS_POSTGRES_CONN) -path ./migrations up

clean:
	rm -rf pkg/

.PHONY: clean migrate

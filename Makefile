ENVIRONMENT ?= test
PROJECT := cats
APPENV := $(ENVIRONMENT)env
GITCOMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
 	GITCOMMIT := $(GITCOMMIT)-dirty
	endif

IMAGE_VERSION ?= $(GITCOMMIT)

all: build

deps:
	docker-compose stop
	docker-compose rm -f
	docker-compose up -d
	docker run --link $(PROJECT)_postgresql:postgres aanand/wait

fmt:
	@govendor fmt +local

migrate:
	migrate -url $($(shell echo $(PROJECT) | tr a-z A-Z)_POSTGRES_CONN) -path ./migrations up

build: deps $(APPENV)
	docker run \
		--link $(PROJECT)_postgresql:postgresql \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-e GODEBUG=netdns=cgo \
		-e PROJECT=github.com/opsee/$(PROJECT) \
		-v `pwd`:/gopath/src/github.com/opsee/$(PROJECT) \
		quay.io/opsee/build-go:proto16
	docker build -t quay.io/opsee/$(PROJECT):$(GITCOMMIT) .

run: build $(APPENV)
	docker run \
		--link $(PROJECT)_postgresql:postgresql \
		--env-file ./$(APPENV) \
		-e GODEBUG=netdns=cgo \
		-e AWS_DEFAULT_REGION \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		--rm \
		quay.io/opsee/$(PROJECT):$(GITCOMMIT)

push:
	docker push quay.io/opsee/$(PROJECT):$(GITCOMMIT)

deploy-plan: terraform
	TF_VAR_image_version=$(IMAGE_VERSION) $(MAKE) -C terraform $(ENVIRONMENT)-plan

deploy: deploy-plan
	$(MAKE) -C terraform $(ENVIRONMENT)-apply

.PHONY: build run migrate all push

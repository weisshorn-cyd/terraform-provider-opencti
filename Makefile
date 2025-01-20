SHELL = /bin/sh

ENV ?= ./docker-compose.env

GOOS ?= linux
GOARCH ?= amd64
VERSION ?= 0.1.0

include $(ENV)

COMPOSE_FILE := ./docker-compose.yml
COMPOSE_ENV_FILE := ./docker-compose.env

all: fmt lint install generate

.PHONY: build
build:
	go build -o terraform-provider-opencti_$(VERSION)

.PHONY: prepare-examples
prepare-examples:
	mkdir -p examples/.terraform/plugins/terraform.local/weisshorn-cyd/opencti/$(VERSION)/$(GOOS)_$(GOARCH)
	mkdir -p examples/terraform.d/plugins/terraform.local/weisshorn-cyd/opencti/$(VERSION)/$(GOOS)_$(GOARCH)
	cp terraform-provider-opencti_* examples/.terraform/plugins/terraform.local/weisshorn-cyd/opencti/$(VERSION)/$(GOOS)_$(GOARCH)/
	cp terraform-provider-opencti_* examples/terraform.d/plugins/terraform.local/weisshorn-cyd/opencti/$(VERSION)/$(GOOS)_$(GOARCH)/

.PHONY: build-examples
build-examples: build prepare-examples

.PHONY: fmt
fmt:
	gofumpt -l -w .
	wsl --fix ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: start-opencti
start-opencti:
	sudo docker compose --file $(COMPOSE_FILE) --env-file $(COMPOSE_ENV_FILE) up -d

.PHONY: stop-opencti
stop-opencti:
	sudo docker compose --file $(COMPOSE_FILE) --env-file $(COMPOSE_ENV_FILE) down

.PHONY: generate
generate:
	cd tools; go generate ./...

.PHONY: test
test:
	go test -v -cover -timeout=120s -parallel=10 ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: clean
clean:
	go clean -cache -testcache -modcache
	rm -f ./terraform-provider-opencti
	rm -rf ./examples/.terraform
	rm -rf ./examples/terraform.d

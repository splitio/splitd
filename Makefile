.PHONY: clean build test sidecar_image unit-tests entrypoint-test

# Setup defaults
GO ?=go
DOCKER ?= docker
SHELL = /usr/bin/env bash -o pipefail

# setup platform specific docker image builds
PLATFORM ?=
PLATFORM_STR := $(if $(PLATFORM),--platform=$(PLATFORM),)

VERSION	:= $(shell cat splitio/version.go | grep 'const Version' | sed 's/const Version = //' | tr -d '"')
GO_FILES := $(shell find . -name "*.go") go.sum

default: help

## generate go.sum from go.mod
go.sum: go.mod
	go mod tidy

## delete built assets
clean:
	rm -Rf splitcli \
		splitd \
		splitd-linux-amd64-$(VERSION).bin \
		splitd-darwin-amd64-$(VERSION).bin \
		splitd-linux-arm-$(VERSION).bin \
		splitd-darwin-arm-$(VERSION).bin

## build binaries for this platform
build: splitd splitcli

## run all tests
test: unit-tests entrypoint-test

## run go unit tests
unit-tests:
	$(GO) test ./... -count=1 -race -coverprofile=coverage.out

## run bash entrypoint tests
entrypoint-test: splitd # requires splitd binary to generate a config and validate env var forwarding
	bash infra/test/test_entrypoint.sh

## build splitd for local machine
splitd: $(GO_FILES)
	go build -o splitd cmd/splitd/main.go

## build splitcli for local machine
splitcli: $(GO_FILES)
	go build -o splitcli cmd/splitcli/main.go

## build docker images for sidecar
images_release: # entrypoints
	$(DOCKER) build $(PLATFORM_STR) -t splitsoftware/splitd-sidecar:latest -t splitsoftware/splitd-sidecar:$(VERSION) -f infra/sidecar.Dockerfile .
	@echo "Image created. Make sure everything works ok, and then run the following commands to push them."
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar:latest"
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar:$(VERSION)"

## build release for binaires
binaries_release: splitd-linux-amd64-$(VERSION).bin splitd-darwin-amd64-$(VERSION).bin splitd-linux-arm-$(VERSION).bin splitd-darwin-arm-$(VERSION).bin

splitd-linux-amd64-$(VERSION).bin: $(GO_FILES)
	GOARCH=amd64 GOOS=linux $(GO) build -o $@ cmd/splitd/main.go

splitd-darwin-amd64-$(VERSION).bin: $(GO_FILES)
	GOARCH=amd64 GOOS=darwin $(GO) build -o $@ cmd/splitd/main.go

splitd-linux-arm-$(VERSION).bin: $(GO_FILES)
	GOARCH=arm64 GOOS=linux $(GO) build -o $@ cmd/splitd/main.go

splitd-darwin-arm-$(VERSION).bin: $(GO_FILES)
	GOARCH=arm64 GOOS=darwin $(GO) build -o $@ cmd/splitd/main.go

# Help target borrowed from: https://docs.cloudposse.com/reference/best-practices/make-best-practices/
## This help screen
help:
	@printf "Available targets:\n\n"
	@awk '/^[a-zA-Z\-\_0-9%:\\]+/ { \
	    helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
		    helpCommand = $$1; \
		    helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
		    gsub("\\\\", "", helpCommand); \
		    gsub(":+$$", "", helpCommand); \
		    printf "  \x1b[32;01m%-35s\x1b[0m %s\n", helpCommand, helpMessage; \
		} \
	    } \
	    { lastLine = $$0 }' $(MAKEFILE_LIST) | sort -u
	@printf "\n"

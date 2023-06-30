.PHONY: clean build test sidecar_image unit-tests entrypoint-test

# Setup defaults
GO ?=go
DOCKER ?= docker
SHELL = /usr/bin/env bash -o pipefail
PLATFORM ?=
VERSION	:= $(shell cat splitio/version.go | grep 'const Version' | sed 's/const Version = //' | tr -d '"')

GO_FILES := $(shell find . -name "*.go") go.sum

go.sum: go.mod
	go mod tidy

clean:
	rm -Rf splitcli splitd

build: splitd split-cli

test: unit-tests entrypoint-test

unit-tests:
	$(GO) test ./... -count=1 -race

entrypoint-test:
	bash infra/test/test_entrypoint.sh

splitd: $(GO_FILES)
	go build -o splitd cmd/splitd/main.go

splitcli: $(GO_FILES)
	go build -o splitcli cmd/splitcli/main.go

images_release: # entrypoints
	$(DOCKER) build $(platform_str) -t splitsoftware/splitd-sidecar:latest -t splitsoftware/splitd-sidecar:$(VERSION) -f infra/sidecar.Dockerfile .
	@echo "Image created. Make sure everything works ok, and then run the following commands to push them."
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar:latest"
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar:$(VERSION)"


# helper macros
platform_str = $(if $(PLATFORM),--platform=$(PLATFORM),)

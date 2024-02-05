.PHONY: clean build test sidecar_image unit-tests entrypoint-test splitio/commitsha.go

# Setup defaults
GO ?=go
DOCKER ?= docker
SHELL = /usr/bin/env bash -o pipefail

# setup platform specific docker image builds
PLATFORM ?=
PLATFORM_STR := $(if $(PLATFORM),--platform=$(PLATFORM),)

VERSION	:= $(shell cat splitio/version.go | grep 'const Version' | sed 's/const Version = //' | tr -d '"')
COMMIT_SHA := $(shell bash -c '[[ ! -z $${GITHUB_SHA} ]] && echo $${GITHUB_SHA:0:7} || git rev-parse --short=7 HEAD')
COMMIT_SHA_FILE := splitio/commitsha.go

GO_FILES := $(shell find . -name "*.go" -not -name "$(COMMIT_SHA_FILE)") go.sum
ENFORCE_FIPS := -tags enforce_fips

CONFIG_TEMPLATE ?= splitd.yaml.tpl
COVERAGE_FILE ?= coverage.out


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
build: splitd splitcli sdhelper

## print current commit SHA (from repo metadata if local, from env-var if GHA)
printsha:
	@echo $(COMMIT_SHA)


## run all tests
test: unit-tests entrypoint-test

## run go unit tests
unit-tests:
	$(GO) test ./... -count=1 -race -coverprofile=$(COVERAGE_FILE)

## display unit test coverage derived from last test run (use `make test display-coverage` for up-to-date results)
display-coverage: coverage.out
	go tool cover -html=coverage.out

## run bash entrypoint tests
entrypoint-test: splitd # requires splitd binary to generate a config and validate env var forwarding
	bash infra/test/test_entrypoint.sh

## build splitd for local machine
splitd: $(GO_FILES) $(COMMIT_SHA_FILE)
	go build -o splitd cmd/splitd/main.go

## build splitd for local machine
splitd-fips: $(GO_FILES) $(COMMIT_SHA_FILE)
	GOEXPERIMENT=boringcrypto go build -o $@ $(ENFORCE_FIPS) cmd/splitd/main.go

## build splitcli for local machine
splitcli: $(GO_FILES)
	go build -o splitcli cmd/splitcli/main.go

## regenerate config file template with defaults
$(CONFIG_TEMPLATE): $(SOURCES) sdhelper
	./sdhelper -command="gen-config-template" > $(CONFIG_TEMPLATE)
## build splitd helper (for code/doc generation purposes only)
sdhelper: $(GO_FILES)
	go build -o sdhelper cmd/sdhelper/main.go

## build docker images for sidecar
images_release: # entrypoints
	$(DOCKER) build $(PLATFORM_STR) \
		-t splitsoftware/splitd-sidecar:latest -t splitsoftware/splitd-sidecar:$(VERSION) \
		--build-arg COMMIT_SHA=$(COMMIT_SHA) \
		-f infra/sidecar.Dockerfile .
	$(DOCKER) build $(PLATFORM_STR) -t splitsoftware/splitd-sidecar-fips:latest -t splitsoftware/splitd-sidecar-fips:$(VERSION) \
		--build-arg FIPS_MODE=1 --build-arg COMMIT_SHA=$(COMMIT_SHA) \
		-f infra/sidecar.Dockerfile .
	@echo "Image created. Make sure everything works ok, and then run the following commands to push them."
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar:latest"
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar:$(VERSION)"
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar-fips:latest"
	@echo "$(DOCKER) push splitsoftware/splitd-sidecar-fips:$(VERSION)"

## build release for binaires
binaries_release: splitd-linux-amd64-$(VERSION).bin splitd-darwin-amd64-$(VERSION).bin splitd-linux-arm-$(VERSION).bin splitd-darwin-arm-$(VERSION).bin

$(COVERAGE_FILE): unit-tests

$(COMMIT_SHA_FILE):
	@echo "package splitio" > $(COMMIT_SHA_FILE)
	@echo "" >> $(COMMIT_SHA_FILE)
	@echo "const CommitSHA = \"$(COMMIT_SHA)\"" >> $(COMMIT_SHA_FILE)

splitd-linux-amd64-$(VERSION).bin: $(GO_FILES)
	GOARCH=amd64 GOOS=linux $(GO) build -o $@ cmd/splitd/main.go

splitd-linux-amd64-fips-$(VERSION).bin: $(GO_FILES)
	GOEXPERIMENT=boringcrypto GOARCH=amd64 GOOS=linux $(GO) build -o $@ $(ENFORCE_FIPS) cmd/splitd/main.go

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

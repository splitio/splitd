.PHONY: clean build test sidecar_image

GO ?=go
GO_FILES := $(shell find . -name "*.go") go.sum

go.sum: go.mod
	go mod tidy

clean:
	rm -Rf splitcli splitd

build: splitd split-cli

test:
	$(GO) test ./... -count=1 -race

splitd: $(GO_FILES)
	go build -o splitd cmd/splitd/main.go

splitcli: $(GO_FILES)
	go build -o splitcli cmd/splitcli/main.go

sidecar_image:
	docker build -f infra/sidecar.Dockerfile -t splitd_sidecar .

.PHONY: clean

GO_FILES := $(shell find . -name "*.go") go.sum

go.sum: go.mod
	go mod tidy

clean:
	rm -Rf split-cli splitd

splitd: $(GO_FILES)
	go build -o splitd cmd/splitd/main.go

split-cli: $(GO_FILES)
	go build -o split-cli cmd/split-cli/main.go


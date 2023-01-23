.PHONY: clean

GO_FILES := $(shell find . -name "*.go") go.sum

go.sum: go.mod
	go mod tidy

clean:
	rm -Rf splitcli splitd

splitd: $(GO_FILES)
	go build -o splitd cmd/splitd/main.go

splitcli: $(GO_FILES)
	go build -o splitcli cmd/splitcli/main.go


.PHONY: clean

clean:
	rm -Rf server splitd

splitd:
	go build -o splitd cmd/splitd/main.go

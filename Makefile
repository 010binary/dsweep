BINARY_NAME=dsweep
MAIN_PACKAGE_PATH=./cmd/dsweep/main.go

.DEFAULT_GOAL := run

.PHONY: run build test fmt clean

run:
	go run $(MAIN_PACKAGE_PATH)

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PACKAGE_PATH)

test:
	go test -v ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin/

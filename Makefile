.PHONY: build test run lint clean

BINARY_NAME=scheduler

build:
	go build -o bin/$(BINARY_NAME) ./cmd/scheduler

test:
	go test -v -race ./...

run:
	go run ./cmd/scheduler

lint:
	golangci-lint run

clean:
	rm -rf bin/
	go clean

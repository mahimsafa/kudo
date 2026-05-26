.PHONY: build test clean

BINARY=kudo
VERSION?=0.1.0

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY) ./cmd/kudo

test:
	go test ./... -v -race

clean:
	rm -rf bin/

lint:
	golangci-lint run ./...

proto:
	protoc --go_out=. --go-grpc_out=. internal/api/proto/*.proto

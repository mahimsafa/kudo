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
	protoc --go_out=internal/api/proto --go_opt=paths=source_relative \
		--go-grpc_out=internal/api/proto --go-grpc_opt=paths=source_relative \
		internal/api/proto/kudo.proto

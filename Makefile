.PHONY: run test build

run:
	go run ./cmd/server

test:
	go test ./...

build:
	go build -o dist/kube-tenant-console ./cmd/server

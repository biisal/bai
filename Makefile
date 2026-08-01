GREETING := Hello from bAI!
SHELL := /bin/bash
BINARY_PATH := ./bin/bai

.PHONY: default build run test release lint clean install format
.ONESHELL:

default:
	@echo "$(GREETING)"

build: 
	go build -o ${BINARY_PATH} ./cmd/bai/...
	echo "build was successful"

run: build
	./${BINARY_PATH}

test:
	go test ./... -failfast

release:
	goreleaser release --clean --snapshot


release-full:
	goreleaser release

lint:
	golangci-lint run


clean:
	rm -rf bin/
	rm -rf tmp/

install:
	go mod download

format:
	gofmt -w .

format-check:
	gofmt -l .


db-generate:
	@sqlc generate
	@gofmt -w -r 'interface{} -> any' ./internal/db

GREETING := Hello from bAI!
SHELL := /bin/bash
BINARY_PATH := ./bin/bai

.PHONY: default build run dev test release lint clean install format
.ONESHELL:

default:
	@echo "$(GREETING)"

build: 
	go build -o ${BINARY_PATH} ./cmd/bai/...
	echo "build was successful"

run: build
	./${BINARY_PATH}
	
dev: build
	./${BINARY_PATH} --dev
	
test:
	go test ./... -failfast

release:
	goreleaser release --clean --snapshot


release-full:
	goreleaser release

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix


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

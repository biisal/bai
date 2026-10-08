GREETING := Hello from bAI!
SHELL := /bin/bash
BINARY_PATH := ./bin/bai
INSTALL_PATH := ~/.local/bin

.PHONY: default build build-linux run dev watch test release lint lint-fix clean install format format-check db-generate frontend-build frontend-dev
.ONESHELL:

default:
	@echo "$(GREETING)"

build:
	go build -o ${BINARY_PATH} ./cmd/bai/...
	echo "build was successful"

install: build
	@mv ${BINARY_PATH} ${INSTALL_PATH} 
	echo "installed to ${INSTALL_PATH}"

run: build
	./${BINARY_PATH}

dev: build
	./${BINARY_PATH} --dev

watch:
	@command -v watchexec >/dev/null 2>&1 || (echo "watchexec is required: install watchexec for hot reloading" && exit 1)
	watchexec -r -e go -- 'go run ./cmd/bai/... --dev'

test:
	go test -race -failfast ./...

release:
	goreleaser release --clean --snapshot

release-full:
	goreleaser release

lint:
	golangci-lint run
	cd frontend && pnpm lint

lint-fix:
	golangci-lint run --fix
	cd frontend && pnpm lint:fix

clean:
	rm -rf bin/
	rm -rf tmp/

format:
	gofmt -w .
	cd frontend && pnpm format

format-check:
	gofmt -l .
	cd frontend && pnpm format:check

db-generate:
	@sqlc generate
	@gofmt -w -r 'interface{} -> any' ./internal/db

frontend-build:
	cd frontend && pnpm build

frontend-dev:
	cd frontend && pnpm dev

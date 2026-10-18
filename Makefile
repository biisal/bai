GREETING := Hello from bAI!
SHELL := /bin/bash
BINARY_PATH := ./bin/bai
INSTALL_PATH := ~/.local/bin

FRONTEND_CMD := pnpm
FRONTEND_PATH := ./frontend

.PHONY: default build build-linux run dev watch test release lint lint-fix clean install format format-check db-generate frontend-build frontend-dev
.ONESHELL:

default:
	@echo "$(GREETING)"

deps:
	go mod download
	cd ${FRONTEND_PATH} && ${FRONTEND_CMD} install

build-frontend:
	cd ${FRONTEND_PATH} && ${FRONTEND_CMD} build


build-bai: 
	go build -ldflags "-s -w -X github.com/biisal/bai/internal/consts.Version=local" -o ${BINARY_PATH} ./cmd/bai/... 
	echo "build was successful"

build: deps build-frontend build-bai

install: build
	@mv ${BINARY_PATH} ${INSTALL_PATH} 
	echo "installed to ${INSTALL_PATH}"

run: build-bai
	./${BINARY_PATH}

dev: build-bai
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
	cd ${FRONTEND_PATH} && ${FRONTEND_CMD} lint

lint-fix:
	golangci-lint run --fix
	cd ${FRONTEND_PATH} && ${FRONTEND_CMD} lint:fix

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
	cd ${FRONTEND_PATH} && ${FRONTEND_CMD} build

frontend-dev:
	cd ${FRONTEND_PATH} && ${FRONTEND_CMD} dev

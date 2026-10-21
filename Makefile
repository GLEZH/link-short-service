include .env

COVERAGE_FILE = coverage.out
BUILD_VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = -X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)

export PROJECT_NAME
export DATABASE_DSN

develop:
	go mod download

build-dev:
	mkdir -p bin
	go build -o ./bin/shortener ./cmd/shortener

build:
	mkdir -p bin
	go build -ldflags "$(BUILD_LDFLAGS)" -o ./bin/shortener ./cmd/shortener

local:
	docker compose --env-file .env.dev up --force-recreate --build

local-down:
	docker compose --env-file .env.dev down

migrate:
	go run github.com/pressly/goose/v3/cmd/goose -dir internal/database/migrations postgres "$(DATABASE_DSN)" up

test:
	go run gotest.tools/gotestsum@v1.13.0 --format pkgname -- -coverpkg=./... -coverprofile=$(COVERAGE_FILE) -covermode=atomic -count=1 ./...
	go tool cover -func=$(COVERAGE_FILE) | tail -n 1

staticlint:
	go run ./cmd/staticlint ./...

run:
	DATABASE_DSN="$(DATABASE_DSN)" go run ./cmd/shortener

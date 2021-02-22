SHELL=bash
MAIN=dp-dimension-search-api

BUILD=build
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

export ENABLE_PRIVATE_ENDPOINTS?=true

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build:
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD)/$(BIN_DIR)/$(MAIN) main.go

.PHONY: debug
debug: build
	HUMAN_LOG=1 go run -race $(LDFLAGS) main.go

.PHONY: acceptance-publishing
acceptance-publishing: build
	ENABLE_PRIVATE_ENDPOINTS=true MONGODB_DATABASE=test HUMAN_LOG=1 go run $(LDFLAGS) main.go

.PHONY: acceptance-web
acceptance-web: build
	ENABLE_PRIVATE_ENDPOINTS=false MONGODB_DATABASE=test HUMAN_LOG=1 go run $(LDFLAGS) main.go

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: build debug test

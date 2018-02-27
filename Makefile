SHELL=bash
MAIN=dp-search-api

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

export ENABLE_PRIVATE_ENDPOINTS=true

build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN)
debug: build
	HUMAN_LOG=1 go run -race main.go
test:
	go test -cover $(shell go list ./... | grep -v /vendor/)
.PHONY: build debug test

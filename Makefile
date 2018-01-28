NAME := rtmp
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.revision=$(REVISION)'
DIR_BIN := ./bin

.DEFAULT_GOAL := help

.PHONY: setup
setup:  ## Setup for required tools.
	go get golang.org/x/tools/cmd/goimports

.PHONY: fmt
fmt: ## Formatting source codes.
	@goimports -w .

.PHONY: test
test: ## Run tests.
	go test .

.PHONY: lint
lint: ## Run go vet.
	@go vet ./...

.PHONY: build
build:  ## Build a binary.
	mkdir -p $(DIR_BIN)
	go build -o $(DIR_BIN)/server -ldflags "$(LDFLAGS)" ./_example/server/main.go
	go build -o $(DIR_BIN)/proxy -ldflags "$(LDFLAGS)" ./_example/proxy/main.go

.PHONY: help
help: ## Show help text
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-20s\033[0m %s\n", $$1, $$2}'
.PHONY: $(filter-out help, $(MAKECMDGOALS))
.DEFAULT_GOAL := help

USER_ID=$(shell id -u ${USER})
GROUP_UD=$(shell id -g ${USER})

DIR=${PWD}

GOOS ?= linux
GOARCH ?= amd64

help:
	@echo "\033[33mUsage:\033[0m\n  make [target] [arg=\"val\"...]\n\n\033[33mTargets:\033[0m"
	@grep -E '^[\.a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[32m%-10s\033[0m %s\n", $$1, $$2}'

mod-tidy: ## update go.sum via Docker
	docker run --rm \
		-v $(DIR):/app \
		-w /app \
		-e GOFLAGS=-mod=mod \
		golang:1.22-alpine ash -c "go mod tidy"

build2: ## build the server binary inside Docker
	docker run --rm \
		-v $(DIR):/app \
		-w /app \
		-e CGO_ENABLED=0 -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) \
		golang:1.22-alpine ash -c "go build -o /app/bin/server_$(GOOS)_$(GOARCH) /app/cmd/server"

test: ## run unit tests via Docker
	docker run --rm \
		-v $(DIR):/app \
		-w /app \
		-e CGO_ENABLED=0 \
		golang:1.22-alpine ash -c "go test ./domain/... ./adapter/web/... -v"

cs-fix: ## format Go source code
	docker run --rm -v $(DIR):/app -w /app golang:1.22-alpine ash -c "gofmt -l -w /app/domain /app/port /app/adapter /app/internal /app/cmd"

up: ## build and start Docker containers
	docker-compose up --force-recreate --build --remove-orphans -d

down: ## stop Docker containers
	docker-compose down

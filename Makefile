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

build: ## to build binary
	docker run -v $(DIR):/app -w /app -e CGO_ENABLED=0 -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) golang:1.17.5-alpine ash -c "go build -o /app/bin/http$(FILENAME_SUBFIX) /app/cmd/http \
		go build -o /app/bin/logger$(FILENAME_SUBFIX) /app/cmd/logger \
		go build -o /app/bin/commander$(FILENAME_SUBFIX) /app/cmd/commander"

build2: ## to build binaries
	docker run \
		-v $(DIR):/app \
		-w /app \
		-e CGO_ENABLED=0 -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) \
		golang:1.17.5-alpine ash -c "sh ./scripts/build-binaries.sh"

cs-fix: ## to fix the coding style issues
	docker run -v $(DIR):/app -w /app golang:1.17.5-alpine ash -c "gofmt -l -w /app/internal /app/cmd"

up: ## to build and start Docker containers
	docker-compose up --force-recreate --build --remove-orphans -d

down: ## to stop Docker containers
	docker-compose down

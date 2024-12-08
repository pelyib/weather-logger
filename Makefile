.PHONY: $(filter-out help, $(MAKECMDGOALS))
.DEFAULT_GOAL:=help


USER_ID=$(shell id -u ${USER})
GROUP_UD=$(shell id -g ${USER})

DIR=${PWD}

GOOS ?= linux
GOARCH ?= amd64

help:
	@echo "\033[33mUsage:\033[0m\n  make [target] [arg=\"val\"...]\n\n\033[33mTargets:\033[0m"
	@grep -E '^[\.a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[32m%-10s\033[0m %s\n", $$1, $$2}'

build: ## to build binary
	@docker run -v $(DIR):/app -w /app -e CGO_ENABLED=0 -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) golang:1.17.5-alpine ash -c "go build -o /app/bin/http$(FILENAME_SUBFIX) /app/cmd/http \
		go build -o /app/bin/logger$(FILENAME_SUBFIX) /app/cmd/logger \
		go build -o /app/bin/commander$(FILENAME_SUBFIX) /app/cmd/commander"

build2: ## to build binaries
	@docker run \
		-v $(DIR):/app \
		-w /app \
		-e CGO_ENABLED=0 -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) \
		golang:1.17.5-alpine ash -c "sh ./scripts/build-binaries.sh"

cs-fix: ## to fix the coding style issues
	@docker run -v $(DIR):/app -w /app golang:1.17.5-alpine ash -c "gofmt -l -w /app/internal /app/cmd"

wait-rabbitmq:
	@docker compose up mq -d
	@printf "waiting for RabbitMQ to be ready..."
	@for i in {1..15}; do \
		sleep 1; \
		printf "."; \
	done; \
	echo ""

wait-couchdb:
	@if [[ ! -f .couchdb.env ]]; then echo ".couchdb.env file not found!"; exit 1; fi
	@docker compose up couchdb -d
	@printf "waiting for CouchDB to be ready..."
	@for i in {1..3}; do \
		sleep 1; \
		printf "."; \
	done; \
	echo ""
	@source .couchdb.env && ./scripts/couchdb.sh ${COUCHDB_ADMIN_NAME} ${COUCHDB_ADMIN_PW} "./.env"

up: wait-rabbitmq wait-couchdb ## to build and start Docker containers
	@docker compose up --remove-orphans -d

down: ## to stop Docker containers
	@docker compose down --remove-orphans

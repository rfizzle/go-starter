APP_NAME="gostarter"
APP_PACKAGE="github.com/rfizzle/go-starter"
GOCMD=go
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOGENERATE=$(GOCMD) generate
GOVULNCHECK=govulncheck
GOFMT=gofmt
GOIMPORTS=goimports
GOLANGCI_LINT=golangci-lint
SQLC=sqlc
GORELEASER=goreleaser
PROJECT_GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./internal/api/*")
PROJECT_PACKAGES=./...
ROOT=$(CURDIR)
YAMLFMT=yamlfmt -conf "${ROOT}/.yamlfmt"
COVERAGE_PROFILE=coverage.out
TOOL_BIN?=

ifneq (,$(findstring 256color, ${TERM}))
	RED     := $(shell tput -Txterm setaf 1)
	GREEN   := $(shell tput -Txterm setaf 2)
	YELLOW  := $(shell tput -Txterm setaf 3)
	BLUE    := $(shell tput -Txterm setaf 4)
	MAGENTA := $(shell tput -Txterm setaf 5)
	CYAN    := $(shell tput -Txterm setaf 6)
	WHITE   := $(shell tput -Txterm setaf 7)
	RESET   := $(shell tput -Txterm sgr0)
else
	BLACK        := ""
	RED          := ""
	GREEN        := ""
	YELLOW       := ""
	LIGHTPURPLE  := ""
	PURPLE       := ""
	BLUE         := ""
	WHITE        := ""
	RESET        := ""
endif

.PHONY: all build linux osx windows clean clean_vendor dev-up dev-down dev-run dev-reset tools fmt fmt-check lint check test race cover vuln ci env generate generate-check sqlc migration sync-version release-snapshot docker check-goreleaser

all: help

## Build:
check-goreleaser:
	@command -v $(GORELEASER) >/dev/null 2>&1 || { echo "${RED}goreleaser not found. Run: make tools${RESET}"; exit 1; }

build: ## Build a local binary in bin/
	@echo "${MAGENTA}Building local binary...${RESET}"
	@mkdir -p bin
	@$(GOCMD) build -o bin/$(APP_NAME) .

# linux/osx/windows below are amd64-only local-dev conveniences (e.g. for the
# Dockerfile's `make linux`). Use release-snapshot for the full release matrix.
linux: check-goreleaser ## Build your project and put the output binary in bin/linux/
	@echo "${MAGENTA}Building package for Linux...${RESET}"
	@mkdir -p bin/linux
	@GOOS=linux GOARCH=amd64 $(GORELEASER) build --single-target --auto-snapshot --clean -o bin/linux/starter

osx: check-goreleaser ## Build your project and put the output binary in bin/osx/
	@echo "${MAGENTA}Building package for OSX...${RESET}"
	@mkdir -p bin/osx
	@GOOS=darwin GOARCH=amd64 $(GORELEASER) build --single-target --auto-snapshot --clean -o bin/osx/starter

windows: check-goreleaser ## Build your project and put the output binary in bin/windows/
	@echo "${MAGENTA}Building package for Windows...${RESET}"
	@mkdir -p bin/windows
	@GOOS=windows GOARCH=amd64 $(GORELEASER) build --single-target --auto-snapshot --clean -o bin/windows/starter.exe

clean: ## Remove build outputs (bin/ and dist/)
	@echo "${MAGENTA}Cleaning binaries...${RESET}"
	@rm -fr ./bin ./dist

clean_vendor: ## Remove built binaries
	@echo "${MAGENTA}Cleaning vendor files...${RESET}"
	@rm -fr ./vendor/*

## Development:
tools: ## Install the exact development tool versions (set TOOL_BIN to isolate)
	@TOOL_BIN="$(TOOL_BIN)" ./scripts/bootstrap-tools.bash

dev-up: ## Start local PostgreSQL and wait until it is healthy
	@docker compose up --detach --wait postgres

dev-down: ## Stop local services while preserving PostgreSQL data
	@docker compose down --remove-orphans

dev-run: dev-up ## Migrate, seed, and run the service using .env
	@test -f .env || { echo "${RED}.env not found. Run: cp .env.example .env${RESET}"; exit 1; }
	@set -a; . ./.env; set +a; $(GOCMD) run . seed
	@set -a; . ./.env; set +a; exec $(GOCMD) run . serve

dev-reset: ## Delete local PostgreSQL data and start a fresh database
	@echo "${YELLOW}Deleting the local PostgreSQL volume and all of its data...${RESET}"
	@docker compose down --volumes --remove-orphans
	@docker compose up --detach --wait postgres

## Format:
fmt: ## Run gofmt on all source files
	@echo "${MAGENTA}Running gofmt...${RESET}"
	@$(GOFMT) -e -s -w $(PROJECT_GOFILES)
	@echo "${MAGENTA}Running goimports...${RESET}"
	@$(GOIMPORTS) -e -format-only -w -d $(PROJECT_GOFILES)
	@echo "${MAGENTA}Running yamlfmt...${RESET}"
	@# Pass the api/ directory (not a *.yaml glob): yamlfmt walks it and
	@# matches every .yaml/.yml file by extension, so the api/openapi.yml
	@# spec is covered. A *.yaml glob silently skipped the .yml spec.
	@$(YAMLFMT) api compose.yaml

fmt-check: ## Verify Go and YAML files are formatted
	@echo "${MAGENTA}Checking gofmt...${RESET}"
	@test -z "$$($(GOFMT) -e -s -l $(PROJECT_GOFILES))"
	@echo "${MAGENTA}Checking goimports...${RESET}"
	@test -z "$$($(GOIMPORTS) -e -format-only -l $(PROJECT_GOFILES))"
	@echo "${MAGENTA}Checking YAML formatting...${RESET}"
	@$(YAMLFMT) -lint api compose.yaml

lint: ## Run go vet and golangci-lint
	@echo "${MAGENTA}Running go vet...${RESET}"
	@$(GOVET) $(PROJECT_PACKAGES)
	@echo "${MAGENTA}Running golangci-lint...${RESET}"
	@$(GOLANGCI_LINT) run

check: generate-check fmt-check lint test ## Run all fast validation checks

## Dependencies:
tidy: ## Rebuild go.mod with only required dependencies
	@echo "${MAGENTA}Rebuilding go.mod file...${RESET}"
	@$(GOMOD) tidy

vendor: tidy ## Sync dependencies to vendor directory
	@echo "${MAGENTA}Syncing dependencies to vendor...${RESET}"
	@$(GOMOD) vendor

env: ## Check env for correct setup
	@echo "${MAGENTA}Checking env...${RESET}"
	@./scripts/check-env.bash

## Test:
test: ## Run the tests of the project
	@echo "${MAGENTA}Running golang tests...${RESET}"
	@$(GOTEST) -v $(PROJECT_PACKAGES)

race: ## Run the tests of the project with race conditions
	@echo "${MAGENTA}Running golang tests...${RESET}"
	@$(GOTEST) -v -race $(PROJECT_PACKAGES)

cover: ## Run tests and write a coverage profile
	@echo "${MAGENTA}Running golang tests with coverage...${RESET}"
	@$(GOTEST) -covermode=atomic -coverprofile=$(COVERAGE_PROFILE) $(PROJECT_PACKAGES)
	@$(GOCMD) tool cover -func=$(COVERAGE_PROFILE)

vuln: ## Scan dependencies for known vulnerabilities
	@command -v $(GOVULNCHECK) >/dev/null 2>&1 || { echo "${RED}govulncheck not found. Run: make tools${RESET}"; exit 1; }
	@$(GOVULNCHECK) $(PROJECT_PACKAGES)

ci: check ## Run the read-only validation workflow used by CI

generate: sqlc ## Run go generate (ogen) and sqlc
	@echo "${MAGENTA}Running golang generations...${RESET}"
	@$(GOGENERATE)

generate-check: generate ## Verify generated code is current
	@echo "${MAGENTA}Checking generated code...${RESET}"
	@git diff --exit-code -- internal/api/v1 internal/storage/postgres/db
	@test -z "$$(git ls-files --others --exclude-standard -- internal/api/v1 internal/storage/postgres/db)" || { \
		echo "${RED}Untracked generated files found:${RESET}"; \
		git ls-files --others --exclude-standard -- internal/api/v1 internal/storage/postgres/db; \
		exit 1; \
	}

sqlc: ## Generate typed SQL access code from sqlc.yaml
	@echo "${MAGENTA}Running sqlc generate...${RESET}"
	@command -v $(SQLC) >/dev/null 2>&1 || { echo "${RED}sqlc not found. Run: make tools${RESET}"; exit 1; }
	@$(SQLC) generate

migration: ## Scaffold a sequential goose SQL migration (usage: make migration name=add_widgets)
	@test -n "$(name)" || { echo "${RED}name is required. Usage: make migration name=add_widgets${RESET}"; exit 1; }
	@command -v goose >/dev/null 2>&1 || { echo "${RED}goose not found. Run: make tools${RESET}"; exit 1; }
	@echo "${MAGENTA}Creating migration $(name)...${RESET}"
	@# -s keeps the 00001/00002 sequential scheme. Always use this target rather
	@# than bare `goose create`: a timestamped file would sort after the padded
	@# sequential ones and break the ordering.
	@goose -s -dir internal/storage/postgres/migrations create $(name) sql

## Release:
release-snapshot: check-goreleaser ## Build snapshot release archives in dist/
	@echo "${MAGENTA}Building snapshot release archives...${RESET}"
	@$(GORELEASER) release --clean --skip=publish,announce --auto-snapshot

sync-version: ## Propagate the git tag version (single source of truth) into .version and api/openapi.yml
	@echo "${MAGENTA}Syncing version from the latest git tag...${RESET}"
	@./scripts/sync-version.bash

## Docker:
docker: ## Use the dockerfile to build the container
	@echo "${MAGENTA}Building docker image...${RESET}"
	@./scripts/docker-build.bash

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

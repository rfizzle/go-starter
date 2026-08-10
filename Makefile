APP_NAME="gostarter"
APP_PACKAGE="github.com/rfizzle/go-starter"
GOCMD=go
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOGENERATE=$(GOCMD) generate
GOVULNCHECK=govulncheck
GITLEAKS=gitleaks
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
COVERAGE_MAINTAINED_PROFILE=coverage.maintained.out
TOOL_BIN?=

# Both branches define the same variables. The fallback assigns a truly empty
# value (no quotes): `:= ""` would expand to two literal characters and print
# them, which is what plain terminals and CI logs get.
ifneq (,$(findstring 256color, ${TERM}))
	RED     := $(shell tput -Txterm setaf 1)
	GREEN   := $(shell tput -Txterm setaf 2)
	YELLOW  := $(shell tput -Txterm setaf 3)
	MAGENTA := $(shell tput -Txterm setaf 5)
	CYAN    := $(shell tput -Txterm setaf 6)
	RESET   := $(shell tput -Txterm sgr0)
else
	RED     :=
	GREEN   :=
	YELLOW  :=
	MAGENTA :=
	CYAN    :=
	RESET   :=
endif

.PHONY: all help build linux osx windows clean clean_vendor dev-up dev-down dev-run dev-reset init tools agents-setup fmt fmt-check lint architecture-check docs-check skills-check check tidy vendor test test-unit test-integration test-all race cover vuln secrets-check ci env generate generate-check sqlc migration release-snapshot container container-check check-goreleaser

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
	@GOOS=linux GOARCH=amd64 $(GORELEASER) build --single-target --snapshot --clean -o bin/linux/starter

osx: check-goreleaser ## Build your project and put the output binary in bin/osx/
	@echo "${MAGENTA}Building package for OSX...${RESET}"
	@mkdir -p bin/osx
	@GOOS=darwin GOARCH=amd64 $(GORELEASER) build --single-target --snapshot --clean -o bin/osx/starter

windows: check-goreleaser ## Build your project and put the output binary in bin/windows/
	@echo "${MAGENTA}Building package for Windows...${RESET}"
	@mkdir -p bin/windows
	@GOOS=windows GOARCH=amd64 $(GORELEASER) build --single-target --snapshot --clean -o bin/windows/starter.exe

clean: ## Remove build outputs (bin/ and dist/)
	@echo "${MAGENTA}Cleaning binaries...${RESET}"
	@rm -fr ./bin ./dist

clean_vendor: ## Remove built binaries
	@echo "${MAGENTA}Cleaning vendor files...${RESET}"
	@rm -fr ./vendor/*

## Development:
# Variable names mirror the initializer flags. They are lowercase (like
# `make migration name=`) so that `app-name` cannot collide with the APP_NAME
# build variable above.
INIT_USAGE=make init module=example.com/acme/orders app-name='Acme Orders' binary=orders-api env-prefix=ORDERS image=registry.example.com/acme/orders-api [description='Order processing API']

init: ## Initialize this clone once (required: module app-name binary env-prefix image; optional: description)
	@test -n "$(module)" || { echo "${RED}module is required. Usage: $(INIT_USAGE)${RESET}"; exit 1; }
	@test -n "$(app-name)" || { echo "${RED}app-name is required. Usage: $(INIT_USAGE)${RESET}"; exit 1; }
	@test -n "$(binary)" || { echo "${RED}binary is required. Usage: $(INIT_USAGE)${RESET}"; exit 1; }
	@test -n "$(env-prefix)" || { echo "${RED}env-prefix is required. Usage: $(INIT_USAGE)${RESET}"; exit 1; }
	@test -n "$(image)" || { echo "${RED}image is required. Usage: $(INIT_USAGE)${RESET}"; exit 1; }
	@echo "${MAGENTA}Initializing template identity...${RESET}"
	@$(GOCMD) run ./tools/initialize \
		--module "$(module)" \
		--app-name "$(app-name)" \
		--binary "$(binary)" \
		--env-prefix "$(env-prefix)" \
		--image "$(image)" \
		$(if $(description),--description "$(description)")

tools: ## Install the exact development tool versions (set TOOL_BIN to isolate)
	@TOOL_BIN="$(TOOL_BIN)" ./scripts/bootstrap-tools.bash

agents-setup: ## Regenerate the relative .claude skill and agent discovery links
	@echo "${MAGENTA}Regenerating agent discovery links...${RESET}"
	@./scripts/setup-agents.bash

dev-up: ## Start local PostgreSQL and wait until it is healthy
	@podman compose up --detach --wait postgres

dev-down: ## Stop local services while preserving PostgreSQL data
	@podman compose down --remove-orphans

dev-run: dev-up ## Migrate, seed, and run the service using .env
	@test -f .env || { echo "${RED}.env not found. Run: cp .env.example .env${RESET}"; exit 1; }
	@set -a; . ./.env; set +a; $(GOCMD) run . migrate up
	@set -a; . ./.env; set +a; $(GOCMD) run . seed
	@set -a; . ./.env; set +a; exec $(GOCMD) run . serve

dev-reset: ## Delete local PostgreSQL data and start a fresh database
	@echo "${YELLOW}Deleting the local PostgreSQL volume and all of its data...${RESET}"
	@podman compose down --volumes --remove-orphans
	@podman compose up --detach --wait postgres

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

architecture-check: ## Verify module, domain, platform, and generated-transport import boundaries
	@./scripts/check-architecture.bash

docs-check: ## Verify documentation links, commands, paths, and Go version
	@./scripts/check-docs.bash

skills-check: ## Validate the portable skill library and its discovery links
	@command -v node >/dev/null 2>&1 || { echo "${RED}node not found. Install Node.js to validate the skill library.${RESET}"; exit 1; }
	@echo "${MAGENTA}Validating skill library...${RESET}"
	@node .agents/skills/scripts/validate-skills.mjs
	@./scripts/setup-agents.bash --check

check: architecture-check docs-check skills-check generate-check fmt-check lint test ## Run all fast validation checks

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
test: test-unit ## Run the unit tests of the project

test-unit: ## Run tests that do not require external services
	@echo "${MAGENTA}Running golang tests...${RESET}"
	@$(GOTEST) -v $(PROJECT_PACKAGES)

test-integration: ## Run tests backed by disposable external services
	@echo "${MAGENTA}Running integration tests...${RESET}"
	@$(GOTEST) -v -count=1 -tags=integration $(PROJECT_PACKAGES)

test-all: test-unit test-integration ## Run unit and integration tests

race: ## Run the tests of the project with race conditions
	@echo "${MAGENTA}Running golang tests...${RESET}"
	@$(GOTEST) -v -race $(PROJECT_PACKAGES)

cover: ## Run unit and integration tests and report/gate maintained-code coverage
	@echo "${MAGENTA}Running golang tests with coverage...${RESET}"
	@$(GOTEST) -covermode=atomic -coverprofile=$(COVERAGE_PROFILE) -tags=integration -count=1 $(PROJECT_PACKAGES)
	@./scripts/check-coverage.bash $(COVERAGE_PROFILE) $(COVERAGE_MAINTAINED_PROFILE)

vuln: ## Scan dependencies for known vulnerabilities
	@command -v $(GOVULNCHECK) >/dev/null 2>&1 || { echo "${RED}govulncheck not found. Run: make tools${RESET}"; exit 1; }
	@$(GOVULNCHECK) $(PROJECT_PACKAGES)

secrets-check: ## Scan the full git history for committed secrets
	@command -v $(GITLEAKS) >/dev/null 2>&1 || { echo "${RED}gitleaks not found. Run: make tools${RESET}"; exit 1; }
	@echo "${MAGENTA}Scanning git history for secrets...${RESET}"
	@# --redact keeps a detected value out of the (public) CI log; the report
	@# still names the commit, file, and line needed to fix it.
	@$(GITLEAKS) git --config .gitleaks.toml --redact --no-banner .

ci: check secrets-check vuln ## Run the read-only validation workflow used by CI (check + secrets-check + vuln)

generate: sqlc ## Run go generate (ogen) and sqlc
	@echo "${MAGENTA}Running golang generations...${RESET}"
	@$(GOGENERATE)

generate-check: generate ## Verify generated code is current
	@echo "${MAGENTA}Checking generated code...${RESET}"
	@git diff --exit-code -- internal/api/v1 internal/platform/database/postgres/db
	@test -z "$$(git ls-files --others --exclude-standard -- internal/api/v1 internal/platform/database/postgres/db)" || { \
		echo "${RED}Untracked generated files found:${RESET}"; \
		git ls-files --others --exclude-standard -- internal/api/v1 internal/platform/database/postgres/db; \
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
	@goose -s -dir internal/platform/database/postgres/migrations create $(name) sql

## Release:
release-snapshot: check-goreleaser ## Build a nonpublishing release snapshot in dist/
	@echo "${MAGENTA}Building snapshot release archives...${RESET}"
	@$(GORELEASER) release --clean --skip=publish,announce,docker --snapshot

## Container:
container: ## Build the container image with Podman
	@echo "${MAGENTA}Building container image with Podman...${RESET}"
	@./scripts/container-build.bash

container-check: check-goreleaser ## Validate release and container packaging configuration
	@echo "${MAGENTA}Checking GoReleaser configuration...${RESET}"
	@$(GORELEASER) check

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

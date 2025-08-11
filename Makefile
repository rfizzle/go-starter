APP_NAME="gostarter"
APP_PACKAGE="github.com/rfizzle/go-starter"
GOCMD=go
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOGENERATE=$(GOCMD) generate
GOFMT=gofmt
GOIMPORTS=goimports
GOLANGCI_LINT=golangci-lint
PROJECT_GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
PROJECT_PACKAGES=$(shell go list ./...)
ROOT=$(shell pwd)
YAMLFMT=yamlfmt -conf "${ROOT}/.yamlfmt"

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

.PHONY: all build linux osx windows clean clean_vendor test env

all: help

## Build:
build: generate fmt clean tidy ## Build all binaries
	@echo "${MAGENTA}Building all distro packages...${RESET}"
	@./scripts/build.bash --osx --linux --windows

linux: ## Build your project and put the output binary in bin/linux/
	@echo "${MAGENTA}Building package for Linux...${RESET}"
	@./scripts/build.bash --linux

osx: ## Build your project and put the output binary in bin/osx/
	@echo "${MAGENTA}Building package for OSX...${RESET}"
	@./scripts/build.bash --osx

windows: ## Build your project and put the output binary in bin/windows/
	@echo "${MAGENTA}Building package for Windows...${RESET}"
	@./scripts/build.bash --windows

clean: ## Remove build related file (bin/ and vendor/)
	@echo "${MAGENTA}Cleaning binaries...${RESET}"
	@rm -fr ./bin/osx ./bin/linux ./bin/windows

clean_vendor: ## Remove built binaries
	@echo "${MAGENTA}Cleaning vendor files...${RESET}"
	@rm -fr ./vendor/*

## Format:
fmt: ## Run gofmt on all source files
	@echo "${MAGENTA}Running gofmt...${RESET}"
	@$(GOFMT) -e -s -w $(PROJECT_GOFILES)
	@echo "${MAGENTA}Running goimports...${RESET}"
	@$(GOIMPORTS) -e -format-only -w -d $(PROJECT_GOFILES)
	@echo "${MAGENTA}Running yamlfmt...${RESET}"
	@$(YAMLFMT) api/*.yaml
	@$(YAMLFMT) api/**/*.yaml

lint: ## Run go vet and golangci-lint
	@echo "${MAGENTA}Running go vet...${RESET}"
	@$(GOVET) $(PROJECT_PACKAGES)
	@echo "${MAGENTA}Running golangci-lint...${RESET}"
	@$(GOLANGCI_LINT) run

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

ci: ## Run tests and lint in CI
	@echo "${MAGENTA}Running golang tests...${RESET}"
	@$(GOTEST) -v -failfast $(PROJECT_PACKAGES)
	@echo "${MAGENTA}Running gofmt...${RESET}"
	@test -z $($(GOFMT) -e -s -l $(PROJECT_GOFILES))
	@echo "${MAGENTA}Running golangci-lint...${RESET}"
	@$(GOLANGCI_LINT) run

generate: ## Run go generate
	@echo "${MAGENTA}Running golang generations...${RESET}"
	@$(GOGENERATE)

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
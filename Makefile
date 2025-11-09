# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Project variables
DOCKER_IMAGE = gloo-foo/vsl
BUILD_DIR ?= bin
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git symbolic-ref -q --short HEAD)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE ?= $(shell date +%FT%T%z)
LDFLAGS = -X main.appVersion=${VERSION} -X main.commitHash=${COMMIT_HASH} -X main.buildDate=${BUILD_DATE}

export CGO_ENABLED ?= 0

## Build

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
DIST_BINARY = dist/vsl-${GOOS}-${GOARCH}$(if $(filter windows,${GOOS}),.exe,)
BIN_TARGET = ${BUILD_DIR}/vsl-${GOOS}-${GOARCH}$(if $(filter windows,${GOOS}),.exe,)

${BUILD_DIR}:
	@mkdir -p $@

${DIST_BINARY}:
	GOFLAGS="-buildvcs=false" go tool goreleaser build --single-target --snapshot --clean

${BIN_TARGET}: ${BUILD_DIR} ${DIST_BINARY}
	cp ${DIST_BINARY} $@

.PHONY: build
build: ${BIN_TARGET} ## Build binary for current platform only

.PHONY: build-all
build-all: ## Build binaries for all platforms
	GOFLAGS="-buildvcs=false" go tool goreleaser build --snapshot --clean

.PHONY: release
release: ## Create a release with goreleaser
	go tool goreleaser release --clean

.PHONY: release-snapshot
release-snapshot: ## Create a snapshot release (no git tag required)
	GOFLAGS="-buildvcs=false" go tool goreleaser release --snapshot --clean

.PHONY: clean
clean: ## Clean builds
	rm -rf ${BUILD_DIR}/vsl* dist/

## Test

.PHONY: test
test: ## Run tests
	GOFLAGS="-buildvcs=false" go tool gotestsum --format short -- ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	GOFLAGS="-buildvcs=false" go tool gotestsum --format short-verbose -- -v ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	GOFLAGS="-buildvcs=false" go test -run ^TestIntegration ./...

.PHONY: test-functional
test-functional: ## Run functional tests
	GOFLAGS="-buildvcs=false" go test -run ^TestFunctional ./...

## Code Quality

.PHONY: lint
lint: ## Run linter
	GOFLAGS="-buildvcs=false" go tool golangci-lint run

.PHONY: check
check: lint test ## Run tests and linters

.PHONY: gorelease
gorelease: ## Check for API compatibility issues
	go tool gorelease

## Code Generation

.PHONY: generate
generate: ## Generate code
	go generate ./...

.PHONY: graphql
graphql: ## Generate GraphQL code
	go tool gqlgen

.PHONY: proto
proto: ## Generate protobuf code
	go tool buf generate

## Docker

.PHONY: docker
docker: ## Build a Docker image
	docker build -t ${DOCKER_IMAGE}:${VERSION} .

.PHONY: docker-debug
docker-debug: ## Build a Docker image with remote debugging capabilities
	docker build -t ${DOCKER_IMAGE}:${VERSION}-debug --build-arg BUILD_TARGET=debug .

## Utilities

.PHONY: tidy
tidy: ## Tidy and verify dependencies
	go mod tidy
	go mod verify

.PHONY: fmt
fmt: ## Format code
	go tool gofumpt -l -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# Variable outputting/exporting rules
var-%: ; @echo $($*)
varexport-%: ; @echo $*=$($*)

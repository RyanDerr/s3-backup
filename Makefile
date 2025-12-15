# Makefile for s3-backup
# Provides convenient commands for building, testing, and releasing

.PHONY: help build test test-coverage test-race test-fuzz lint clean install docker-build docker-run deps fmt vet

# Variables
BINARY_NAME=s3-backup
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.Version=${VERSION}"
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
FUZZ_TIME?=30s

# Docker
DOCKER_IMAGE?=s3-backup
DOCKER_TAG?=${VERSION}

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

##@ General

help: ## Display this help
	@echo "${COLOR_BOLD}s3-backup Makefile${COLOR_RESET}"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make ${COLOR_BLUE}<target>${COLOR_RESET}\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  ${COLOR_BLUE}%-15s${COLOR_RESET} %s\n", $$1, $$2 } /^##@/ { printf "\n${COLOR_BOLD}%s${COLOR_RESET}\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

deps: ## Download dependencies
	@echo "${COLOR_GREEN}Downloading dependencies...${COLOR_RESET}"
	go mod download
	go mod verify
	go mod tidy

fmt: ## Format code
	@echo "${COLOR_GREEN}Formatting code...${COLOR_RESET}"
	gofmt -s -w .
	go mod tidy

vet: ## Run go vet
	@echo "${COLOR_GREEN}Running go vet...${COLOR_RESET}"
	go vet ./...

lint: ## Run golangci-lint
	@echo "${COLOR_GREEN}Running linter...${COLOR_RESET}"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "${COLOR_YELLOW}golangci-lint not installed. Install with:${COLOR_RESET}"; \
		echo "  brew install golangci-lint"; \
		echo "  or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

##@ Testing

test: ## Run tests
	@echo "${COLOR_GREEN}Running tests...${COLOR_RESET}"
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "${COLOR_GREEN}Running tests with coverage...${COLOR_RESET}"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "${COLOR_GREEN}Coverage report generated: coverage.html${COLOR_RESET}"
	go tool cover -func=coverage.out | grep total:

test-race: ## Run tests with race detector
	@echo "${COLOR_GREEN}Running tests with race detector...${COLOR_RESET}"
	go test -v -race ./...

test-fuzz: ## Run fuzz tests (use FUZZ_TIME to set duration, e.g., make test-fuzz FUZZ_TIME=60s)
	@echo "${COLOR_GREEN}Running fuzz tests for ${FUZZ_TIME}...${COLOR_RESET}"
	@echo "${COLOR_YELLOW}Testing config package...${COLOR_RESET}"
	go test -fuzz=FuzzLoadFromFile -fuzztime=${FUZZ_TIME} ./internal/config
	go test -fuzz=FuzzLoadFromEnv -fuzztime=${FUZZ_TIME} ./internal/config
	go test -fuzz=FuzzValidateAWSRegion -fuzztime=${FUZZ_TIME} ./internal/config
	go test -fuzz=FuzzValidateS3Bucket -fuzztime=${FUZZ_TIME} ./internal/config
	@echo "${COLOR_YELLOW}Testing s3 package...${COLOR_RESET}"
	go test -fuzz=FuzzBuildObjectKey -fuzztime=${FUZZ_TIME} ./internal/s3
	go test -fuzz=FuzzFileCollectorWalk -fuzztime=${FUZZ_TIME} ./internal/s3
	go test -fuzz=FuzzValidateDirectories -fuzztime=${FUZZ_TIME} ./internal/s3
	@echo "${COLOR_GREEN}Fuzz tests completed!${COLOR_RESET}"

test-all: deps vet test test-race ## Run all tests (standard, race, coverage)
	@echo "${COLOR_GREEN}All tests passed!${COLOR_RESET}"

##@ Building

build: ## Build binary for current platform
	@echo "${COLOR_GREEN}Building ${BINARY_NAME} ${VERSION}...${COLOR_RESET}"
	go build ${LDFLAGS} -o ${BINARY_NAME} .
	@echo "${COLOR_GREEN}Build complete: ${BINARY_NAME}${COLOR_RESET}"

build-all: ## Build binaries for all platforms
	@echo "${COLOR_GREEN}Building for all platforms...${COLOR_RESET}"
	@mkdir -p dist
	@for platform in ${PLATFORMS}; do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		OUTPUT="dist/${BINARY_NAME}-${VERSION}-$${platform%/*}-$${platform#*/}"; \
		if [ "$${platform%/*}" = "windows" ]; then OUTPUT="$$OUTPUT.exe"; fi; \
		echo "Building $$OUTPUT..."; \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} go build ${LDFLAGS} -o $$OUTPUT .; \
		if [ "$$?" -eq 0 ]; then \
			(cd dist && sha256sum $$(basename $$OUTPUT) > $$(basename $$OUTPUT).sha256); \
		fi; \
	done
	@echo "${COLOR_GREEN}All platform builds complete! Check dist/ directory${COLOR_RESET}"

install: ## Install binary to $GOPATH/bin
	@echo "${COLOR_GREEN}Installing ${BINARY_NAME}...${COLOR_RESET}"
	go install ${LDFLAGS} .
	@echo "${COLOR_GREEN}Installed to $$(go env GOPATH)/bin/${BINARY_NAME}${COLOR_RESET}"

clean: ## Clean build artifacts
	@echo "${COLOR_GREEN}Cleaning...${COLOR_RESET}"
	rm -f ${BINARY_NAME}
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -rf testdata/fuzz/
	go clean -cache -testcache -modcache -fuzzcache
	@echo "${COLOR_GREEN}Clean complete${COLOR_RESET}"

##@ Docker

docker-build: ## Build Docker image
	@echo "${COLOR_GREEN}Building Docker image ${DOCKER_IMAGE}:${DOCKER_TAG}...${COLOR_RESET}"
	docker build --build-arg VERSION=${VERSION} -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	@echo "${COLOR_GREEN}Docker image built: ${DOCKER_IMAGE}:${DOCKER_TAG}${COLOR_RESET}"

docker-build-multi: ## Build multi-platform Docker image
	@echo "${COLOR_GREEN}Building multi-platform Docker image...${COLOR_RESET}"
	docker buildx build --platform linux/amd64,linux/arm64 \
		--build-arg VERSION=${VERSION} \
		-t ${DOCKER_IMAGE}:${DOCKER_TAG} \
		-t ${DOCKER_IMAGE}:latest \
		.
	@echo "${COLOR_GREEN}Multi-platform image built${COLOR_RESET}"

docker-run: ## Run Docker container (requires env vars: BACKUP_DIRS, AWS_REGION, S3_BUCKET)
	@echo "${COLOR_GREEN}Running Docker container...${COLOR_RESET}"
	@if [ -z "${BACKUP_DIRS}" ]; then echo "${COLOR_YELLOW}BACKUP_DIRS not set${COLOR_RESET}"; exit 1; fi
	@if [ -z "${AWS_REGION}" ]; then echo "${COLOR_YELLOW}AWS_REGION not set${COLOR_RESET}"; exit 1; fi
	@if [ -z "${S3_BUCKET}" ]; then echo "${COLOR_YELLOW}S3_BUCKET not set${COLOR_RESET}"; exit 1; fi
	docker run --rm \
		-e BACKUP_DIRS=${BACKUP_DIRS} \
		-e AWS_REGION=${AWS_REGION} \
		-e S3_BUCKET=${S3_BUCKET} \
		-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
		-e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
		${DOCKER_IMAGE}:${DOCKER_TAG}

docker-shell: ## Run shell in Docker container
	@echo "${COLOR_GREEN}Opening shell in container...${COLOR_RESET}"
	docker run -it --rm --entrypoint sh ${DOCKER_IMAGE}:${DOCKER_TAG}

##@ Release

release-check: ## Check if ready for release
	@echo "${COLOR_GREEN}Checking release readiness...${COLOR_RESET}"
	@if [ -z "$$(git status --porcelain)" ]; then \
		echo "${COLOR_GREEN}✓ Working directory clean${COLOR_RESET}"; \
	else \
		echo "${COLOR_YELLOW}✗ Working directory has uncommitted changes${COLOR_RESET}"; \
		exit 1; \
	fi
	@if git describe --exact-match --tags HEAD >/dev/null 2>&1; then \
		echo "${COLOR_GREEN}✓ On tagged commit: $$(git describe --tags)${COLOR_RESET}"; \
	else \
		echo "${COLOR_YELLOW}✗ Not on a tagged commit${COLOR_RESET}"; \
		exit 1; \
	fi
	@make test-all
	@echo "${COLOR_GREEN}✓ Ready for release!${COLOR_RESET}"

tag: ## Create a new tag (use: make tag VERSION=v1.0.0)
	@if [ -z "${VERSION}" ]; then \
		echo "${COLOR_YELLOW}VERSION not set. Usage: make tag VERSION=v1.0.0${COLOR_RESET}"; \
		exit 1; \
	fi
	@echo "${COLOR_GREEN}Creating tag ${VERSION}...${COLOR_RESET}"
	git tag -a ${VERSION} -m "Release ${VERSION}"
	@echo "${COLOR_GREEN}Tag created. Push with: git push origin ${VERSION}${COLOR_RESET}"

##@ Utilities

run: build ## Build and run
	@echo "${COLOR_GREEN}Running ${BINARY_NAME}...${COLOR_RESET}"
	./${BINARY_NAME}

version: ## Show version information
	@echo "${COLOR_BOLD}Version Info:${COLOR_RESET}"
	@echo "  Version: ${VERSION}"
	@echo "  Commit:  ${COMMIT}"
	@echo "  Built:   ${BUILD_TIME}"

info: ## Show build information
	@echo "${COLOR_BOLD}Build Info:${COLOR_RESET}"
	@echo "  Go Version:    $$(go version)"
	@echo "  Binary Name:   ${BINARY_NAME}"
	@echo "  Version:       ${VERSION}"
	@echo "  LDFLAGS:       ${LDFLAGS}"
	@echo "  Platforms:     ${PLATFORMS}"
	@echo "  Docker Image:  ${DOCKER_IMAGE}:${DOCKER_TAG}"

vendor: ## Vendor dependencies
	@echo "${COLOR_GREEN}Vendoring dependencies...${COLOR_RESET}"
	go mod vendor
	@echo "${COLOR_GREEN}Dependencies vendored to vendor/${COLOR_RESET}"

update-deps: ## Update dependencies to latest versions
	@echo "${COLOR_GREEN}Updating dependencies...${COLOR_RESET}"
	go get -u ./...
	go mod tidy
	@echo "${COLOR_GREEN}Dependencies updated${COLOR_RESET}"

.DEFAULT_GOAL := help

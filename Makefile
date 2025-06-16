# Federation-Go Makefile
# Comprehensive build and management system for GraphQL Federation architecture

# Variables
PROJECT_NAME := federation-go
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d ' ' -f 3)

# Build directories
BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/bin
DEPLOYMENTS_DIR := deployments
DOCKER_DIR := $(DEPLOYMENTS_DIR)/docker
KUBERNETES_DIR := $(DEPLOYMENTS_DIR)/kubernetes

# Service names
SERVICES := account catalog
GATEWAYS := gateway aggregator

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.goVersion=$(GO_VERSION)"
BUILD_FLAGS := -v $(LDFLAGS)

# Docker settings
DOCKER_REGISTRY := localhost:5000
DOCKER_TAG := $(VERSION)

.PHONY: help setup clean build build-all test test-coverage lint fmt vet mod-tidy
.PHONY: run-gateway run-aggregator run-account run-catalog run-all stop-all
.PHONY: docker-build docker-push docker-run infra-up infra-down
.PHONY: proto-gen graphql-gen deps-update security-scan

# Default target
all: clean build test

# Help target
help: ## Show this help message
	@echo "$(BLUE)Federation-Go Build System$(NC)"
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Setup development environment
setup: ## Setup development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/99designs/gqlgen@latest
	@go install github.com/air-verse/air@latest
	@npm install meta concurrently
	@curl -sSf https://atlasgo.sh | sh
	@mkdir -p $(BIN_DIR) $(DOCKER_DIR) $(KUBERNETES_DIR)
	@echo "$(GREEN)Development environment setup complete!$(NC)"

# Clean build artifacts
clean: ## Clean generated folder in each service
	@echo "$(BLUE)Cleaning generated folder...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Cleaning $$service generated folder...$(NC)"; \
		cd services/$$service && rm -rf generated && cd ../..; \
	done
	@echo "$(GREEN)Generated folder cleaned!$(NC)"

# Build all services
build: ## Build all services
	@echo "$(BLUE)Building all services...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Building $$service...$(NC)"; \
		cd services/$$service && go build $(BUILD_FLAGS) -o ../../$(BIN_DIR)/$$service ./cmd/app && cd ../..; \
	done
	@for gateway in $(GATEWAYS); do \
		echo "$(YELLOW)Building $$gateway...$(NC)"; \
		cd services/$$service && go build $(BUILD_FLAGS) -o ../../$(BIN_DIR)/$$service ./cmd/app && cd ../..; \
	done
	@echo "$(GREEN)All services built successfully!$(NC)"

# Testing
test: ## Run all tests in each service
	@echo "$(BLUE)Running tests...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Running tests for $$service...$(NC)"; \
		cd services/$$service && go test -v ./... && cd ../..; \
	done
	@echo "$(GREEN)All tests passed!$(NC)"

test-coverage: ## Run tests with coverage in each service
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Running tests with coverage for $$service...$(NC)"; \
		cd services/$$service && go test -v -coverprofile=coverage.out ./... && cd ../..; \
	done
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-integration: ## Run integration tests in each service
	@echo "$(BLUE)Running integration tests...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Running integration tests for $$service...$(NC)"; \
		cd services/$$service && go test -v -tags=integration ./... && cd ../..; \
	done
	@echo "$(GREEN)Integration tests completed!$(NC)"

# Code quality
lint: ## Run linter in each service
	@echo "$(BLUE)Running linter...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Running linter for $$service...$(NC)"; \
		cd services/$$service && golangci-lint run ./... && cd ../..; \
	done
	@echo "$(GREEN)Linting completed!$(NC)"

fmt: ## Format code in each service
	@echo "$(BLUE)Formatting code...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Formatting $$service...$(NC)"; \
		cd services/$$service && go fmt ./... && cd ../..; \
	done
	@echo "$(GREEN)Code formatted!$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Running go vet for $$service...$(NC)"; \
		cd services/$$service && go vet ./... && cd ../..; \
	done
	@echo "$(GREEN)Vet completed!$(NC)"

# Dependency management
mod-tidy: ## Tidy go modules for all services
	@echo "$(BLUE)Tidying go modules...$(NC)"
	@cd package && go mod tidy
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Tidying $$service modules...$(NC)"; \
		cd services/$$service && go mod tidy && cd ../..; \
	done
	@echo "$(GREEN)All modules tidied!$(NC)"

# Running services locally through air 
run-service: 
	@echo "$(BLUE)Starting all services...$(NC)"
	@concurrently -c auto --names "$(shell echo $(SERVICES) | tr ' ' ',')" \
		$(foreach service,$(SERVICES),"cd services/$(service) && air")
	@echo "$(GREEN)All services started!$(NC)"

run-gateway:
	@echo "$(BLUE)Starting all gateways...$(NC)"
	@for gateway in $(GATEWAYS); do \
		echo "$(YELLOW)Building $$gateway...$(NC)"; \
		cd services/$$gateway && air && cd ../..; \
	done
	@echo "$(GREEN)All gateways started!$(NC)"

# Docker operations
docker-build: ## Build Docker images for all services
	@echo "$(BLUE)Building Docker images...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Building $$service image...$(NC)"; \
		docker build -f deployments/docker/Dockerfile.$$service -t $(DOCKER_REGISTRY)/$(PROJECT_NAME)-$$service:$(DOCKER_TAG) .; \
	done
	@echo "$(GREEN)All Docker images built!$(NC)"

docker-push: docker-build ## Push Docker images to registry
	@echo "$(BLUE)Pushing Docker images...$(NC)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Pushing $$service image...$(NC)"; \
		docker push $(DOCKER_REGISTRY)/$(PROJECT_NAME)-$$service:$(DOCKER_TAG); \
	done
	@echo "$(GREEN)All Docker images pushed!$(NC)"

docker-run: ## Run services using Docker Compose
	@echo "$(BLUE)Starting services with Docker Compose...$(NC)"
	@docker-compose -f deployments/docker/docker-compose.yml up -d
	@echo "$(GREEN)Services started with Docker Compose!$(NC)"

# Infrastructure management
infra-up: ## Start infrastructure services (NATS, etcd, PostgreSQL, Redis)
	@echo "$(BLUE)Starting infrastructure services...$(NC)"
	@docker-compose -f deployments/docker/docker-compose.infra.yml up -d
	@echo "$(GREEN)Infrastructure services started!$(NC)"

infra-down: ## Stop infrastructure services
	@echo "$(YELLOW)Stopping infrastructure services...$(NC)"
	@docker-compose -f deployments/docker/docker-compose.infra.yml down
	@echo "$(GREEN)Infrastructure services stopped!$(NC)"

# Code generation
graphql-gen: ## Generate GraphQL code
	@echo "$(BLUE)Generating GraphQL code...$(NC)"
	@for service in $(SERVICES); do \
		if [ -f "services/$$service/gqlgen.yml" ]; then \
			echo "$(YELLOW)Generating GraphQL code for $$service...$(NC)"; \
			cd services/$$service &&  go generate ./... && cd ../..; \
		fi; \
	done
	@echo "$(GREEN)GraphQL code generated!$(NC)"

# Security and analysis
security-scan: ## Run security scan in each service
	@echo "$(BLUE)Running security scan...$(NC)"
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Running security scan for $$service...$(NC)"; \
		cd services/$$service && gosec ./... && cd ../..; \
	done
	@echo "$(GREEN)Security scan completed!$(NC)"

# Quick development cycle
dev: clean build test ## Quick development cycle: clean, build, test
	@echo "$(GREEN)Development cycle completed!$(NC)" 
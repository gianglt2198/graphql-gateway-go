.PHONY: structure

structure:
	@mkdir -p gateway
	@mkdir -p services/
	@mkdir -p proto
	@mkdir -p pkg/auth
	@mkdir -p pkg/tracing
	@mkdir -p pkg/logging
	@mkdir -p deployment

dependencies:
	@go install github.com/99designs/gqlgen

generate:
	@cd gateway && go run github.com/99designs/gqlgen generate
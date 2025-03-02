.PHONY: structure

structure:
	@mkdir -p cmd/main
	@mkdir -p internal/db
	@mkdir -p internal/models
	@mkdir -p internal/registry

dependencies:
	@go install github.com/99designs/gqlgen

generate:
	@go get github.com/99designs/gqlgen && go run github.com/99designs/gqlgen generate
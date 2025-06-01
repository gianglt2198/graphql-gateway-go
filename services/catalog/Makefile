.PHONY: int generate run clean migrate-new

init:
	@go install github.com/air-verse/air@latest
	@curl -sSf https://atlasgo.sh | sh

structure:
	@mkdir -p config
	@mkdir -p cmd/app
	@mkdir -p cmd/generate
	@mkdir -p migrations
	@mkdir -p internal/services
	@mkdir -p internal/repositories
	@mkdir -p ent/schema
	@mkdir -p graph/schema

# generate code for database and graphql schema
generate:
	@go generate ./...
	@go mod tidy
	@echo "🎉 graphql and database code generated successfully!"

# live-reloading command line utility for development
dev:
	@air

run:
	@go run cmd/app/main.go

test:
	@go test ./...

lint:
	@echo "Running golangci-lint ..."
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run --fix ./...
	@echo "golangci-lint, done!"

clean: 
	@cd ent && find . -not -path "*schema/*" -delete
	@cd graph && rm -rf generated

# generate a blank migration
migrate-new:
	@read -p "Enter migration name (e.g., create-todo-table): " name; \
		go run cmd/migrate/main.go add $$name
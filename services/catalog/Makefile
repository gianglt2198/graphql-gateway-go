.PHONY: dev generate migrate-gen migrate-up migrate-down migrate-new migrate-hash migrate-status clean help

# live-reloading command line utility for development
dev:
	@air

generate:
	@go generate ./...

# Migration commands
migrate-gen:
	@echo "Generating migration from Ent schema..."
	@go run cmd/migrate/migrate.go migrate-gen $(if $(name),-n $(name))

migrate-up:
	@echo "Running migrations..."
	@go run cmd/migrate/migrate.go migrate-up

migrate-down:
	@echo "Rolling back to version: $(version)"
	@go run cmd/migrate/migrate.go migrate-down $(if $(version),-v $(version))

migrate-new:
	@echo "Creating new empty migration: $(name)"
	@go run cmd/migrate/migrate.go migrate-new -n $(name)

migrate-hash:
	@echo "Regenerating migration hashes..."
	@go run cmd/migrate/migrate.go migrate-hash

migrate-status:
	@echo "Checking migration status..."
	@go run cmd/migrate/migrate.go migrate-status

clean:
	@rm -rf generated/

# Help target
help:
	@echo "Available targets:"
	@echo "  dev            - Start development server with live reload"
	@echo "  generate       - Generate Ent and GraphQL code"
	@echo "  migrate-gen    - Generate migration from Ent schema (required: name=migration_name)"
	@echo "  migrate-up     - Run pending migrations"
	@echo "  migrate-down   - Roll back migrations (required: version=target_version)"
	@echo "  migrate-new    - Create new empty migration (required: name=migration_name)"
	@echo "  migrate-hash   - Regenerate migration hashes"
	@echo "  migrate-status - Show migration status"
	@echo "  clean          - Clean generated files"
	@echo "  help           - Show this help message"


# go install -v github.com/incu6us/goimports-reviser/v3@latest
fmt:
	@echo "Running goimports-reviser ..."
	find . -type f -name "*.go" \
	  ! -path "./cmd/*" \
	  ! -path "./tools/*" \
	  ! -path "./generated/*" \
	  ! -path "./vendor/*" \
	  -exec goimports-reviser \
	   -company-prefixes github.com/gianglt2198/federation-go/package \
	   -rm-unused -set-alias -format {} \;
	@echo "goimports-reviser done!"
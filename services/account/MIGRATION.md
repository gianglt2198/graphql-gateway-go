# Database Migration Guide

This document explains how to use the database migration system for the account service, which uses **Atlas CLI** and **Ent** for comprehensive schema management.

## Prerequisites

- Atlas CLI (required for migration operations)
- PostgreSQL database
- Go 1.24+

**Installing Atlas CLI:**
```bash
curl -sSf https://atlasgo.sh | sh
```

## Migration Commands

### 1. `migrate-gen` - Generate Migration from Ent Schema

Creates a new migration file based on your current Ent schema changes using Ent's NamedDiff feature.

```bash
# Generate migration with custom name (required)
make migrate-gen name=add_user_avatar
```

**Use Case**: When you've modified your Ent schema files (e.g., `ent/schema/user.go`) and want to create a migration that captures those changes.

### 2. `migrate-up` - Apply Migrations

Executes all pending migrations against the database using Atlas CLI.

```bash
make migrate-up
```

**Use Case**: Apply schema changes to your database. Run this after generating migrations or when setting up a new environment.

### 3. `migrate-down` - Roll Back Migrations

Rolls back database migrations to a specific version for recovery or development purposes.

```bash
# Roll back to specific version (version is required)
make migrate-down version=20240101000000
```

**Use Case**: When you need to undo recent migrations for debugging, testing different approaches, or recovering from migration issues.

### 4. `migrate-new` - Create Empty Migration

Creates a new empty migration file for custom SQL changes using Atlas CLI.

```bash
# Create empty migration (name is required)
make migrate-new name=add_custom_indexes
```

**Use Case**: When you need to add custom SQL that isn't covered by Ent schema generation (indexes, triggers, custom functions, etc.).

### 5. `migrate-hash` - Regenerate Migration Hashes

Regenerates Atlas migration hashes for validation using Atlas CLI.

```bash
make migrate-hash
```

**Use Case**: When migration files have been modified and you need to update their checksums.

### 6. `migrate-status` - Check Migration Status

Shows the current migration status and lists pending migrations.

```bash
make migrate-status
```

**Use Case**: To understand which migrations have been applied and which are pending before making changes.

## Workflow Examples

### Adding a New Field to User Schema

1. **Modify the Ent schema**:
   ```go
   // ent/schema/user.go
   func (User) Fields() []ent.Field {
       return []ent.Field{
           field.String("username").Unique(),
           field.String("email").Unique(),
           field.String("avatar_url").Optional(), // New field
           // ... other fields
       }
   }
   ```

2. **Generate migration**:
   ```bash
   make migrate-gen name=add_user_avatar
   ```

3. **Review the generated migration** in `migrations/` directory.

4. **Apply migration**:
   ```bash
   make migrate-up
   ```

### Rolling Back a Problematic Migration

1. **Check current status**:
   ```bash
   make migrate-status
   ```

2. **Roll back to previous version**:
   ```bash
   make migrate-down version=20240101000000
   ```

3. **Verify rollback**:
   ```bash
   make migrate-status
   ```

### Adding Custom Indexes

1. **Create empty migration**:
   ```bash
   make migrate-new name=add_performance_indexes
   ```

2. **Edit the generated file** in `migrations/` directory:
   ```sql
   -- Add indexes for better query performance
   CREATE INDEX idx_users_email ON users(email);
   CREATE INDEX idx_users_created_at ON users(created_at);
   CREATE INDEX idx_sessions_user_id ON sessions(user_id);
   ```

3. **Apply migration**:
   ```bash
   make migrate-up
   ```

## Configuration

### Database Configuration

Update `config.yml` with your database settings:

```yaml
database:
  driver: "postgres"
  host: "localhost"
  port: 5433
  user: "username"
  password: "password"
  database: "account"
  sslmode: "disable"
```

### Environment Variables

For rollback operations, you can optionally set:
- `DEV_DATABASE_URL`: Development database URL for safe rollback operations

### Atlas Configuration

The `atlas.hcl` file contains environment-specific settings:

- **local**: For local development
- **dev**: For development environment  
- **prod**: For production environment

## Directory Structure

```
services/account/
├── atlas.hcl                    # Atlas configuration
├── cmd/migrate/migrate.go       # Migration tool (Atlas CLI integration)
├── migrations/                  # Migration files directory
│   └── YYYYMMDDHHMMSS_name.sql
├── ent/schema/                  # Ent schema definitions
├── config.yml                  # Service configuration
└── MIGRATION.md                 # This file
```

## Best Practices

### 1. **Always Review Generated Migrations**
- Check the generated SQL before applying
- Ensure the migration captures your intended changes
- Test in development first

### 2. **Use Descriptive Names**
```bash
# Good
make migrate-gen name=add_user_profile_fields
make migrate-new name=optimize_user_queries

# Avoid
make migrate-gen name=update
make migrate-new name=fix
```

### 3. **Check Status Before Changes**
```bash
# Always check status first
make migrate-status

# Then proceed with changes
make migrate-gen name=my_changes
make migrate-up
```

### 4. **Backup Before Major Changes**
```bash
# Create database backup before running migrations
pg_dump account > backup_$(date +%Y%m%d_%H%M%S).sql
make migrate-up
```

### 5. **Version Control**
- Always commit migration files to git
- Never modify existing migration files
- Create new migrations for changes

### 6. **Environment Consistency**
- Use the same migration process across all environments
- Apply migrations in the correct order
- Keep environments in sync

## Troubleshooting

### Common Issues

**1. Atlas CLI not found:**
```bash
# Install Atlas CLI
curl -sSf https://atlasgo.sh | sh

# Verify installation
atlas version
```

**2. Migration fails to apply:**
```bash
# Check database connectivity and migration status
make migrate-status

# Check migration file syntax
atlas migrate validate --dir file://migrations
```

**3. Schema drift detected:**
```bash
# Regenerate hashes
make migrate-hash

# Check for conflicts
make migrate-status
```

**4. Ent schema changes not reflected:**
```bash
# Regenerate Ent code first
make generate

# Then create migration
make migrate-gen name=schema_updates
```

**5. Rollback issues:**
```bash
# Set DEV_DATABASE_URL for safe rollbacks
export DEV_DATABASE_URL="postgres://user:pass@localhost:5432/account_dev"

# Then perform rollback
make migrate-down version=target_version
```

## Advanced Usage

### Direct Atlas CLI Usage

For advanced features, you can use Atlas CLI directly:

```bash
# Validate migrations
atlas migrate validate --dir file://migrations

# Linear migration history
atlas migrate lint --dir file://migrations

# Generate migration with specific options
atlas migrate diff --to file://migrations --dev-url "docker://postgres/15/dev"
```

## Safety Guidelines

1. **Never run migrations directly in production without testing**
2. **Always have rollback plans**
3. **Use transactions for complex migrations**  
4. **Monitor migration performance on large tables**
5. **Keep migration files small and focused**
6. **Test rollback procedures in development**

---

For more information about Atlas and Ent:
- [Atlas Documentation](https://atlasgo.io/)
- [Ent Documentation](https://entgo.io/) 
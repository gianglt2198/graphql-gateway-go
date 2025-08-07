package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"github.com/gianglt2198/federation-go/services/account/config"
	"github.com/gianglt2198/federation-go/services/account/generated/ent/migrate"
)

var (
	migrationDir = "migrations"
	rootCmd      = &cobra.Command{
		Use:   "migrate",
		Short: "Database migration tool using Atlas and Ent",
		Long:  "Comprehensive database migration tool for account service using Atlas CLI and Ent schema generation",
	}
)

func main() {
	rootCmd.AddCommand(
		migrateGenCmd(),
		migrateUpCmd(),
		migrateNewCmd(),
		migrateHashCmd(),
		migrateStatusCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// migrateGenCmd creates a new migration file from Ent schema using NamedDiff
func migrateGenCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "migrate-gen",
		Short: "Generate a new migration file from Ent schema",
		Long:  "Create a new migration file based on the current Ent schema changes using Ent's NamedDiff",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateMigrationFromSchema(name)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the migration (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// migrateUpCmd runs pending migrations
func migrateUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-up",
		Short: "Apply all pending migrations",
		Long:  "Execute all pending migrations against the database using Atlas CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return applyMigrations()
		},
	}
	return cmd
}

// migrateNewCmd creates an empty migration file
func migrateNewCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "migrate-new",
		Short: "Create a new empty migration file",
		Long:  "Create a new empty migration file for custom SQL changes using Atlas CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("migration name is required")
			}
			return createNewMigration(name)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Name for the migration (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// migrateHashCmd regenerates migration hashes
func migrateHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-hash",
		Short: "Regenerate migration hashes",
		Long:  "Regenerate Atlas migration hashes for the migration directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return regenerateMigrationHashes()
		},
	}
	return cmd
}

// migrateStatusCmd shows migration status
func migrateStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-status",
		Short: "Show migration status",
		Long:  "Display the current migration status and pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showMigrationStatus()
		},
	}
	return cmd
}

// generateMigrationFromSchema creates a migration using Ent's NamedDiff
func generateMigrationFromSchema(name string) error {

	if name == "" {
		return fmt.Errorf("migration name is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := regenerateMigrationHashes(); err != nil {
		return err
	}

	// Create atlas migration directory
	localDir, err := atlas.NewLocalDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to create atlas migration directory: %w", err)
	}

	// Configure migration options
	opts := []schema.MigrateOption{
		schema.WithDir(localDir),
		schema.WithMigrationMode(schema.ModeInspect),
		schema.WithDialect(dialect.Postgres),
		schema.WithFormatter(atlas.DefaultFormatter),
		schema.WithIndent("  "),
		schema.WithDropIndex(true),
		schema.WithDropColumn(true),
	}

	fmt.Printf("Generating migration: %s\n %v\n", name, cfg.Database.GetURL())

	ctx := context.Background()
	if err := migrate.NamedDiff(ctx, cfg.Database.GetURL(), name, opts...); err != nil {
		return fmt.Errorf("migration file generation failed: %w", err)
	}

	fmt.Printf("Migration '%s' generated successfully in %s/\n", name, migrationDir)
	return nil
}

// applyMigrations runs all pending migrations
func applyMigrations() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := ensureMigrationDir(); err != nil {
		return err
	}

	return runAtlasCommand("apply",
		"--dir", "file://"+migrationDir,
		"--url", cfg.Database.GetURL(),
	)
}

// createNewMigration creates an empty migration file
func createNewMigration(name string) error {
	if err := ensureMigrationDir(); err != nil {
		return err
	}

	return runAtlasCommand("new",
		"--dir", "file://"+migrationDir,
		name,
	)
}

// regenerateMigrationHashes updates migration hashes
func regenerateMigrationHashes() error {
	if err := ensureMigrationDir(); err != nil {
		return err
	}

	return runAtlasCommand("hash",
		"--dir", "file://"+migrationDir,
	)
}

// showMigrationStatus displays current migration status
func showMigrationStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := ensureMigrationDir(); err != nil {
		return err
	}

	return runAtlasCommand("status",
		"--url", cfg.Database.GetURL(),
		"--dir", "file://"+migrationDir,
	)
}

// runAtlasCommand executes an Atlas CLI command
func runAtlasCommand(command string, args ...string) error {
	cmdArgs := append([]string{"migrate", command}, args...)
	cmd := exec.Command("atlas", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Executing: atlas %s\n", joinArgs(cmdArgs))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("atlas command failed: %w", err)
	}

	return nil
}

// ensureMigrationDir creates the migration directory if it doesn't exist
func ensureMigrationDir() error {
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}
	return nil
}

// joinArgs joins command arguments for display
func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

package migrate

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/spf13/cobra"
)

const migrationsPath = "file://internal/db/migrations"

func NewCmdMigrate(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration management",
		Long: `Manage database migrations with various operations:
  up    - Apply all pending migrations or specific number
  down  - Rollback specific number of migrations
  drop  - Drop all tables (DESTRUCTIVE!)
  force - Force set migration version
  version - Show current migration version
  create - Create new migration files`,
	}

	cmd.AddCommand(
		newUpCmd(cfg),
		newDownCmd(cfg),
		newDropCmd(cfg),
		newForceCmd(cfg),
		newVersionCmd(cfg),
		newCreateCmd(cfg),
	)

	return cmd
}

func newUpCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "up [N]",
		Short: "Apply all pending migrations or N migrations",
		Long:  "Apply all pending migrations. Optionally specify number of migrations to apply.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getMigrator(cfg)
			if err != nil {
				return err
			}
			defer m.Close()

			if len(args) == 0 {
				// Apply all migrations
				if err := m.Up(); err != nil && err != migrate.ErrNoChange {
					return fmt.Errorf("failed to apply migrations: %w", err)
				}
				fmt.Println("✅ All migrations applied successfully")
			} else {
				// Apply N migrations
				n, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid number: %s", args[0])
				}
				if err := m.Steps(n); err != nil && err != migrate.ErrNoChange {
					return fmt.Errorf("failed to apply %d migrations: %w", n, err)
				}
				fmt.Printf("✅ Applied %d migrations successfully\n", n)
			}

			return nil
		},
	}
}

func newDownCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "down N",
		Short: "Rollback N migrations",
		Long:  "Rollback specified number of migrations. This is a DESTRUCTIVE operation!",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid number: %s", args[0])
			}

			// Safety check
			if n <= 0 {
				return fmt.Errorf("number of migrations must be positive")
			}

			fmt.Printf("⚠️  WARNING: This will rollback %d migration(s). This is DESTRUCTIVE!\n", n)
			fmt.Print("Are you sure? (yes/no): ")

			confirm := readInputLine("Are you sure? (yes/no): ")
			if confirm != "yes" {
				fmt.Println("❌ Operation cancelled")
				return nil
			}

			m, err := getMigrator(cfg)
			if err != nil {
				return err
			}
			defer m.Close()

			if err := m.Steps(-n); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to rollback %d migrations: %w", n, err)
			}

			fmt.Printf("✅ Rolled back %d migrations successfully\n", n)
			return nil
		},
	}
}

func newDropCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "drop",
		Short: "Drop all tables (DESTRUCTIVE!)",
		Long:  "Drop all tables in the database. This will completely destroy all data!",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⚠️  DANGER: This will DROP ALL TABLES and DELETE ALL DATA!")
			fmt.Print("Type 'DROP ALL DATA' to confirm: ")

			confirm := readInputLine("Are you sure? (yes/no): ")
			if confirm != "DROP ALL DATA" {
				fmt.Println("❌ Operation cancelled")
				return nil
			}

			m, err := getMigrator(cfg)
			if err != nil {
				return err
			}
			defer m.Close()

			if err := m.Drop(); err != nil {
				return fmt.Errorf("failed to drop database: %w", err)
			}

			fmt.Println("✅ All tables dropped successfully")
			return nil
		},
	}
}

func newForceCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "force VERSION",
		Short: "Force set migration version",
		Long:  "Force set the migration version without running migrations. Use with caution!",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid version: %s", args[0])
			}

			fmt.Printf("⚠️  WARNING: This will force set migration version to %d\n", version)
			fmt.Print("Are you sure? (yes/no): ")

			confirm := readInputLine("Are you sure? (yes/no): ")
			if confirm != "yes" {
				fmt.Println("❌ Operation cancelled")
				return nil
			}

			m, err := getMigrator(cfg)
			if err != nil {
				return err
			}
			defer m.Close()

			if err := m.Force(version); err != nil {
				return fmt.Errorf("failed to force version %d: %w", version, err)
			}

			fmt.Printf("✅ Migration version forced to %d\n", version)
			return nil
		},
	}
}

func newVersionCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show current migration version",
		Long:  "Display the current migration version and dirty state.",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := getMigrator(cfg)
			if err != nil {
				return err
			}
			defer m.Close()

			version, dirty, err := m.Version()
			if err != nil {
				if err == migrate.ErrNilVersion {
					fmt.Println("📋 Migration version: No migrations applied yet")
					return nil
				}
				return fmt.Errorf("failed to get version: %w", err)
			}

			fmt.Printf("📋 Migration version: %d\n", version)
			if dirty {
				fmt.Println("⚠️  State: DIRTY (migration failed)")
			} else {
				fmt.Println("✅ State: CLEAN")
			}

			return nil
		},
	}
}

func newCreateCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "create NAME",
		Short: "Create new migration files",
		Long:  "Create new up and down migration files with timestamp prefix.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			timestamp := time.Now().Format("20060102150405")

			upFile := fmt.Sprintf("internal/db/migrations/%s_%s.up.sql", timestamp, name)
			downFile := fmt.Sprintf("internal/db/migrations/%s_%s.down.sql", timestamp, name)

			// Create up migration file
			upContent := fmt.Sprintf(`-- Migration: %s
-- Created at: %s
-- Description: Add your migration description here

-- Add your UP migration SQL here
-- Example:
-- CREATE TABLE example (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL
-- );
`, name, time.Now().Format("2006-01-02 15:04:05"))

			if err := writeFile(upFile, upContent); err != nil {
				return fmt.Errorf("failed to create up migration: %w", err)
			}

			// Create down migration file
			downContent := fmt.Sprintf(`-- Migration: %s (DOWN)
-- Created at: %s
-- Description: Rollback for %s

-- Add your DOWN migration SQL here
-- Example:
-- DROP TABLE IF EXISTS example;
`, name, time.Now().Format("2006-01-02 15:04:05"), name)

			if err := writeFile(downFile, downContent); err != nil {
				return fmt.Errorf("failed to create down migration: %w", err)
			}

			fmt.Printf("✅ Created migration files:\n")
			fmt.Printf("   📄 %s\n", upFile)
			fmt.Printf("   📄 %s\n", downFile)

			return nil
		},
	}
}

func getMigrator(cfg *config.Config) (*migrate.Migrate, error) {
	m, err := migrate.New(migrationsPath, cfg.Database.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	return m, nil
}

func writeFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0o644)
}

func readInputLine(prompt string) string {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

package cli

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	// Required for registering PostgreSQL driver with golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Required for reading migration files from disk.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

func newCmdMigrate() *cobra.Command {
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
		newUpCmd(),
		newDownCmd(),
		newDropCmd(),
		newForceCmd(),
		newVersionCmd(),
		newCreateCmd(),
	)

	return cmd
}

func newUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up [N]",
		Short: "Apply all (or N) pending migrations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			r, err := newRunner()
			if err != nil {
				return err
			}
			defer r.Close()

			steps := 0
			if len(args) == 1 {
				steps, err = strconv.Atoi(args[0])
				if err != nil {
					return err
				}
			}
			return r.Up(steps)
		},
	}
}

func newDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down N",
		Short: "Rollback N migrations (DESTRUCTIVE)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			if !confirm(
				fmt.Sprintf("⚠️ This will rollback %d migration(s). Type 'yes' to proceed: ", n),
				"yes",
			) {
				slog.Default().Error("❌ cancelled")
				return nil
			}

			r, err := newRunner()
			if err != nil {
				return err
			}
			defer r.Close()

			return r.Down(n)
		},
	}
}

func newDropCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drop",
		Short: "Drop **ALL** tables (DANGER)",
		RunE: func(_ *cobra.Command, _ []string) error {
			if !confirm("TYPE 'DROP ALL DATA' to erase everything: ", "DROP ALL DATA") {
				slog.Default().Error("❌ cancelled")
				return nil
			}

			r, err := newRunner()
			if err != nil {
				return err
			}
			defer r.Close()

			return r.Drop()
		},
	}
}

func newForceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "force VERSION",
		Short: "Set schema version without running migrations",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			v, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			if !confirm("Type 'yes' to force version: ", "yes") {
				return nil
			}
			r, err := newRunner()
			if err != nil {
				return err
			}
			defer r.Close()

			return r.Force(v)
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show current migration version",
		RunE: func(_ *cobra.Command, _ []string) error {
			r, err := newRunner()
			if err != nil {
				return err
			}
			defer r.Close()

			v, dirty, err := r.Version()
			if err != nil {
				return err
			}
			slog.Default().Info("📋 version", "version", v, "dirty", dirty)
			return nil
		},
	}
}

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create NAME",
		Short: "Create new migration files",
		Long:  "Create new up and down migration files with timestamp prefix.",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			timestamp := time.Now().Format("20060102150405")

			upFile := fmt.Sprintf("db/migrations/%s_%s.up.sql", timestamp, name)
			downFile := fmt.Sprintf("db/migrations/%s_%s.down.sql", timestamp, name)

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

			slog.Default().Info("✅ Created migration files", "upFile", upFile, "downFile", downFile)

			return nil
		},
	}
}

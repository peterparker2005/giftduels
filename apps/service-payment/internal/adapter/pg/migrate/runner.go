package migratepg

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	// Required for registering PostgreSQL driver with golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Required for reading migration files from disk.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const migrationsPath = "file://db/migrations"

type Runner struct {
	m *migrate.Migrate
}

func NewWithDSN(dsn string) (*Runner, error) {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	return &Runner{m: m}, nil
}

func (r *Runner) Close() {
	_, _ = r.m.Close()
}

func (r *Runner) Down(steps int) error {
	if steps <= 0 {
		return errors.New("steps must be positive")
	}
	err := r.m.Steps(-steps)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (r *Runner) Drop() error {
	return r.m.Drop()
}

func (r *Runner) Force(version int) error {
	return r.m.Force(version)
}

func (r *Runner) Up(steps int) error {
	if steps == 0 {
		err := r.m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	}
	return r.m.Steps(steps)
}

func (r *Runner) Version() (uint, bool, error) {
	return r.m.Version()
}

package migrate

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
)

func Run(cfg *config.Config) error {
	m, err := migrate.New(
		"file://internal/db/migrations",
		cfg.Database.Address(), // e.g. "postgres://user:pw@host:5432/dbname?sslmode=disable"
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

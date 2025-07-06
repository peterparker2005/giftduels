package db

import (
	"database/sql"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"go.uber.org/fx"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Module предоставляет database зависимости
var Module = fx.Module("database",
	fx.Provide(
		func(cfg *config.Config) (*sql.DB, error) {
			db, err := sql.Open("postgres", cfg.Database.Address())
			if err != nil {
				return nil, err
			}

			// Проверяем соединение
			if err := db.Ping(); err != nil {
				return nil, err
			}

			return db, nil
		},
		func(sqlDB *sql.DB) DBTX {
			return sqlDB
		},
		New, // *Queries
	),
)

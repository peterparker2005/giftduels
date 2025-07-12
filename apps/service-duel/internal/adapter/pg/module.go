package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Module("pg",
	fx.Provide(
		NewDuelRepository,
		NewPgxTxManager,
		func(cfg *config.Config) (*pgxpool.Pool, error) {
			return Connect(context.Background(), Config{
				DSN:             cfg.Database.DSN(),
				MaxConns:        cfg.Database.MaxConns,
				ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
			})
		},
	),
)

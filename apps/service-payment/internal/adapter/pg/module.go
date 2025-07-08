package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Module("pg",
	fx.Provide(
		func(cfg *config.Config) (*pgxpool.Pool, error) {
			return Connect(context.Background(), Config{
				DSN:             cfg.Database.DSN(),
				MaxConns:        cfg.Database.MaxConns,
				ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
			})
		},
		NewBalanceRepository,
	),
)

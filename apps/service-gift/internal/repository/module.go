package repository

import (
	"github.com/peterparker2005/giftduels/packages/logger-go"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/db"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/pricing"
	giftRepo "github.com/peterparker2005/giftduels/apps/service-gift/internal/repository/gift"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/stubs"
	"go.uber.org/fx"
)

// Module предоставляет repository зависимости
var Module = fx.Module("repositories",
	fx.Provide(
		// Gift repository
		func(queries *db.Queries) gift.Repository {
			return giftRepo.NewSQLRepository(queries)
		},
		// Pricing repository
		func(cfg *config.Config, logger *logger.Logger) pricing.Repository {
			return stubs.NewPricingFake()
		},
	),
)

package service

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/domain/gift"
	giftService "github.com/peterparker2005/giftduels/apps/service-gift/internal/service/gift"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
)

// Module предоставляет service зависимости
var Module = fx.Module("services",
	fx.Provide(
		func(repo gift.GiftRepository, txMgr pg.TxManager, log *logger.Logger) *giftService.Service {
			return giftService.New(repo, txMgr, log)
		},
	),
)

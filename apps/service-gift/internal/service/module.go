package service

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/portals"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/command"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/query"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/saga"
	"go.uber.org/fx"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("services",
	fx.Provide(
		command.NewGiftEventHandler,
		command.NewGiftStakeCommand,
		command.NewGiftWithdrawCommand,
		command.NewGiftReturnFromGameCommand,

		query.NewGiftReadService,
		query.NewUserGiftsService,

		saga.NewWithdrawalSaga,

		portals.NewPortalsPriceService,
	),
)

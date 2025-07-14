package service

import (
	"go.uber.org/fx"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/duel"
)

var Module = fx.Module("service",
	fx.Provide(
		duel.NewDuelService,
	),
)

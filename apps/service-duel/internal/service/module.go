package service

import (
	"go.uber.org/fx"

	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/command"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/query"
)

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("service",
	fx.Provide(
		// Commands
		command.NewDuelCreateCommand,
		command.NewDuelAutoRollCommand,
		command.NewDuelJoinCommand,
		// Queries
		query.NewDuelQueryService,
	),
)

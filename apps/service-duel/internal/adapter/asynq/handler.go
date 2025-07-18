package asynq

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	dueldomain "github.com/peterparker2005/giftduels/apps/service-duel/internal/domain/duel"
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/service/command"
	"github.com/peterparker2005/giftduels/packages/logger-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func RegisterHandlers(
	lc fx.Lifecycle,
	srv *asynq.Server,
	autoRollCmd *command.DuelAutoRollCommand,
	log *logger.Logger,
) {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeAutoRoll, func(ctx context.Context, task *asynq.Task) error {
		var p PayloadAutoRoll
		if err := json.Unmarshal(task.Payload(), &p); err != nil {
			return err
		}
		return autoRollCmd.Execute(ctx, dueldomain.ID(p.DuelID))
	})

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.Run(mux); err != nil {
					log.Error("failed to run asynq server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			srv.Shutdown()
			return nil
		},
	})
}

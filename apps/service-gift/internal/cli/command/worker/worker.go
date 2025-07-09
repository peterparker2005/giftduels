package worker

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/adapter/pg"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/app"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/eventhandler"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func NewCmdWorker(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Run event worker",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWorker(cfg)
		},
	}
}

func runWorker(cfg *config.Config) error {
	app := fx.New(
		app.LoggerModule,
		pg.Module,
		eventhandler.Module,
		fx.Provide(func() *config.Config { return cfg }),
	)

	app.Run()
	return nil
}

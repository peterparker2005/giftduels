package consume

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/app"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/db"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/event"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/repository"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func NewCmdConsume(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "consume",
		Short: "Run event consumer",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConsumer(cfg)
		},
	}
}

func runConsumer(cfg *config.Config) error {
	app := fx.New(
		app.LoggerModule,
		db.Module,
		repository.Module,
		event.Module,
		fx.Provide(func() *config.Config { return cfg }),
	)

	app.Run()
	return nil
}

package worker

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/service/eventhandler"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func NewCmdWorker() *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Run event worker",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWorker()
		},
	}
}

func runWorker() error {
	app := fx.New(
		eventhandler.Module,
	)

	app.Run()
	return nil
}

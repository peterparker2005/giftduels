package worker

import (
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
	app := fx.New()

	app.Run()
	return nil
}

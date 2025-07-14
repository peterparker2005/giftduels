package cli

import (
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/app"
	"github.com/spf13/cobra"
)

func newCmdServe() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		RunE: func(_ *cobra.Command, _ []string) error {
			grpcApp := app.NewGRPCApp()
			grpcApp.Run()
			return nil
		},
	}
}

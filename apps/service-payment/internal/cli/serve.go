package cli

import (
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/app"
	"github.com/spf13/cobra"
)

func newCmdServe() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcApp := app.NewGRPCApp()
			grpcApp.Run()
			return nil
		},
	}
}

package serve

import (
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/app"
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/config"
	"github.com/spf13/cobra"
)

func NewCmdServe(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.Run(cfg)
			return nil
		},
	}
}

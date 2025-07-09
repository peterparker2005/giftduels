package serve

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/app"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.Run()
			return nil
		},
	}
}

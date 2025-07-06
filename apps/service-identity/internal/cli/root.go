package cli

import (
	"log"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/cli/command/serve"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/packages/cli-go"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := cli.RootCmd()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	cmd.AddCommand(
		serve.NewCmdServe(cfg),
	)

	return cmd
}

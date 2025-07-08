package cli

import (
	"log"

	"github.com/peterparker2005/giftduels/apps/service-gift/internal/cli/command/migrate"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/cli/command/serve"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/cli/command/worker"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/config"
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
		worker.NewCmdWorker(cfg),
		migrate.NewCmdMigrate(cfg),
	)

	return cmd
}

package cli

import (
	"log/slog"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/cli/command/migrate"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/cli/command/serve"
	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
	"github.com/peterparker2005/giftduels/packages/cli-go"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := cli.RootCmd()

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Default().Error("failed to load config", "error", err)
		return nil
	}

	cmd.AddCommand(
		serve.NewCmdServe(cfg),
		migrate.NewCmdMigrate(cfg),
	)

	return cmd
}

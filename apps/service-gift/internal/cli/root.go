package cli

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/cli/command/migrate"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/cli/command/serve"
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/cli/command/worker"
	"github.com/peterparker2005/giftduels/packages/cli-go"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := cli.RootCmd()

	cmd.AddCommand(
		serve.NewCmdServe(),
		worker.NewCmdWorker(),
		migrate.NewCmdMigrate(),
	)

	return cmd
}

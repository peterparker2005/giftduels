package cli

import (
	"github.com/peterparker2005/giftduels/apps/service-event/internal/cli/command/serve"
	"github.com/peterparker2005/giftduels/apps/service-event/internal/cli/command/worker"
	"github.com/peterparker2005/giftduels/packages/cli-go"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := cli.RootCmd()

	cmd.AddCommand(
		serve.NewCmdServe(),
		worker.NewCmdWorker(),
	)

	return cmd
}

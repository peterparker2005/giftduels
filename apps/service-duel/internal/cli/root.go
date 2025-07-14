package cli

import (
	"github.com/peterparker2005/giftduels/packages/cli-go"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	cmd := cli.RootCmd()

	cmd.AddCommand(
		newCmdServe(),
		newCmdMigrate(),
		newCmdWorker(),
	)

	return cmd
}

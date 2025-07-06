package cli

import (
	"github.com/peterparker2005/giftduels/packages/cli-go/command/version"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	cmd := newRootCmd()

	return cmd
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "CLI for service-identity",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.PersistentFlags().Bool("debug", false, "enable debug logging")

	cmd.AddCommand(
		version.NewCmdVersion(),
	)
	return cmd
}

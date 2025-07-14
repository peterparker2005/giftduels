package version

import (
	"log/slog"

	"github.com/peterparker2005/giftduels/packages/version-go"
	"github.com/spf13/cobra"
)

func NewCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version of the service",
		Run: func(_ *cobra.Command, _ []string) {
			slog.Default().Info("version", "version", version.Version)
		},
	}
}

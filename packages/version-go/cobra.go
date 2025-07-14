package version

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// AddCommand добавляет в rootCmd sub-команду `version`.
func AddCommand(root *cobra.Command, name, short string) {
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: short,
		Run: func(_ *cobra.Command, _ []string) {
			slog.Default().Info(name, "version", Version)
		},
	})
}

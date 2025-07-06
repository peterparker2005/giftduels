// packages/version/cobra.go
package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// AddCommand добавляет в rootCmd sub-команду `version`
func AddCommand(root *cobra.Command, name, short string) {
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(name, Version)
		},
	})
}

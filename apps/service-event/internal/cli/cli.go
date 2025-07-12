package cli

import (
	"os"
)

func Run() {
	rootCmd := rootCmd()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

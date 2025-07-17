package cli

import (
	"github.com/peterparker2005/giftduels/apps/service-event/internal/app"
	"github.com/spf13/cobra"
)

func newCmdWorker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Run event worker",
		Run: func(_ *cobra.Command, _ []string) {
			workerApp := app.NewWorkerApp()
			workerApp.Run()
		},
	}

	return cmd
}

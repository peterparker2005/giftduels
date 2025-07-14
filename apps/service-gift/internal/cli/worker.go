package cli

import (
	"github.com/peterparker2005/giftduels/apps/service-gift/internal/app"
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

	cmd.AddCommand(
		newCmdWorkerEvent(),
	)

	return cmd
}

func newCmdWorkerEvent() *cobra.Command {
	return &cobra.Command{
		Use:   "event",
		Short: "Run event worker",
		Run: func(_ *cobra.Command, _ []string) {
			workerApp := app.NewWorkerApp()
			workerApp.Run()
		},
	}
}

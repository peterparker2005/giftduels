package cli

import (
	"github.com/peterparker2005/giftduels/apps/service-payment/internal/app"
	"github.com/spf13/cobra"
)

func newCmdWorker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Run event worker",
		Run: func(cmd *cobra.Command, _ []string) {
			workerApp := app.NewWorkerApp()
			workerApp.Run()
		},
	}

	cmd.AddCommand(
		newCmdWorkerEvent(),
		newCmdWorkerTon(),
	)

	return cmd
}

func newCmdWorkerEvent() *cobra.Command {
	return &cobra.Command{
		Use:   "event",
		Short: "Run event worker",
		Run: func(cmd *cobra.Command, args []string) {
			workerApp := app.NewWorkerApp()
			workerApp.Run()
		},
	}
}

func newCmdWorkerTon() *cobra.Command {
	return &cobra.Command{
		Use:   "ton",
		Short: "Run TON blocсkchain poller",
		Run: func(cmd *cobra.Command, args []string) {
			workerTonApp := app.NewWorkerTonApp()
			workerTonApp.Run()
		},
	}
}

package app

import (
	"go.uber.org/fx"
)

func NewWorkerApp() *fx.App {
	return fx.New(
		fx.Options(
			CommonModule,
		),
	)
}

package stream

import "go.uber.org/fx"

func ProvideMessageMapper() MessageMapper {
	return DuelMessageMapper
}

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Options(
	fx.Provide(ProvideMessageMapper),
)

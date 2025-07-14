package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/peterparker2005/giftduels/packages/configs"
	"go.uber.org/fx"
)

type Config struct {
	configs.ServiceBaseConfig

	Logger configs.LoggerConfig `yaml:"logger"`

	// shared configs
	Database configs.DatabaseConfig `yaml:"database"`
	GRPC     configs.GRPCConfig     `yaml:"grpc"`
	AMQP     configs.AMQPConfig     `yaml:"amqp"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	_ = cleanenv.ReadConfig(".env", &cfg)

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

//nolint:gochecknoglobals // fx module pattern
var Module = fx.Module("config",
	fx.Provide(LoadConfig),
)

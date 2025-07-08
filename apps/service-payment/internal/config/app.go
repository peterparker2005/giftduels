package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/peterparker2005/giftduels/packages/configs"
)

type Config struct {
	configs.ServiceBaseConfig
	Logger   configs.LoggerConfig   `yaml:"logger"`
	Database configs.DatabaseConfig `yaml:"database"`
	AMQP     configs.AMQPConfig     `yaml:"amqp"`
	GRPC     configs.GRPCConfig     `yaml:"grpc"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	_ = cleanenv.ReadConfig(".env", &cfg)

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

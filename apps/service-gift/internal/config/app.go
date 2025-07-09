package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/peterparker2005/giftduels/packages/configs"
	"go.uber.org/fx"
)

type TelegramConfig struct {
	BotToken string `yaml:"bot_token" env:"TELEGRAM_BOT_TOKEN"`
}

type TonnelApiConfig struct {
	InitData string `yaml:"init_data" env:"TONNEL_API_INIT_DATA"`
}

type Config struct {
	configs.ServiceBaseConfig
	Logger          configs.LoggerConfig `yaml:"logger"`
	Telegram        TelegramConfig       `yaml:"telegram"`
	TonnelApiConfig TonnelApiConfig      `yaml:"tonnel"`

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

var Module = fx.Module("config",
	fx.Provide(LoadConfig),
)

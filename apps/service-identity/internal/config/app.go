package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/peterparker2005/giftduels/packages/configs"
)

type TelegramConfig struct {
	BotToken string `yaml:"bot_token" env:"TELEGRAM_BOT_TOKEN"`
}

type JWTConfig struct {
	Secret     string        `yaml:"secret"     env:"JWT_SECRET"     env-default:"supersecret"`
	Expiration time.Duration `yaml:"expiration" env:"JWT_EXPIRATION" env-default:"24h"`
}

type Config struct {
	configs.ServiceBaseConfig

	Database configs.DatabaseConfig `yaml:"database"`
	Logger   configs.LoggerConfig   `yaml:"logger"`
	Telegram TelegramConfig         `yaml:"telegram"`
	JWT      JWTConfig              `yaml:"jwt"`
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

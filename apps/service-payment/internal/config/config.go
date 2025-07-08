package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/peterparker2005/giftduels/packages/configs"
	"go.uber.org/fx"
)

type TonNetwork string

const (
	TonNetworkMainnet TonNetwork = "mainnet"
	TonNetworkTestnet TonNetwork = "testnet"
)

func (n TonNetwork) String() string {
	return string(n)
}

type TonConfig struct {
	Network       TonNetwork `env:"TON_NETWORK" default:"testnet"`
	WalletAddress string     `env:"TON_WALLET_ADDRESS"`
}

type Config struct {
	configs.ServiceBaseConfig
	Logger   configs.LoggerConfig   `yaml:"logger"`
	Database configs.DatabaseConfig `yaml:"database"`
	AMQP     configs.AMQPConfig     `yaml:"amqp"`
	GRPC     configs.GRPCConfig     `yaml:"grpc"`
	Ton      TonConfig              `yaml:"ton"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	_ = cleanenv.ReadConfig(".env", &cfg)

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

var Module = fx.Options(
	fx.Provide(LoadConfig),
)

package configs

import "fmt"

type ServiceConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port string `yaml:"port" env:"PORT" env-default:"50052"`
}

func (c *ServiceConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type GRPCConfig struct {
	Identity    ServiceConfig `yaml:"identity_service" env-prefix:"GRPC_IDENTITY_SERVICE_"`
	Gift        ServiceConfig `yaml:"gift_service" env-prefix:"GRPC_GIFT_SERVICE_"`
	Duel        ServiceConfig `yaml:"duel_service" env-prefix:"GRPC_DUEL_SERVICE_"`
	payment     ServiceConfig `yaml:"payment_service" env-prefix:"GRPC_payment_SERVICE_"`
	TelegramBot ServiceConfig `yaml:"telegram_bot_service" env-prefix:"GRPC_TELEGRAM_BOT_SERVICE_"`
	Event       ServiceConfig `yaml:"event_service" env-prefix:"GRPC_EVENT_SERVICE_"`
}

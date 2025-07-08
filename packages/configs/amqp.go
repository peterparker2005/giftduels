package configs

import "fmt"

type AMQPConfig struct {
	User     string `yaml:"user" env:"AMQP_USER" env-default:"admin"`
	Password string `yaml:"password" env:"AMQP_PASSWORD" env-default:"admin"`
	Host     string `yaml:"host" env:"AMQP_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"AMQP_PORT" env-default:"5672"`
}

func (c *AMQPConfig) Address() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d",
		c.User, c.Password, c.Host, c.Port)
}

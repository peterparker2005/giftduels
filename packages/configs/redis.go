package configs

import (
	"net"
	"strconv"
)

type RedisConfig struct {
	Host     string `yaml:"host"     env:"REDIS_HOST"     env-default:"localhost"`
	Port     int    `yaml:"port"     env:"REDIS_PORT"     env-default:"6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	DB       int    `yaml:"db"       env:"REDIS_DB"       env-default:"0"`
	Username string `yaml:"username" env:"REDIS_USERNAME" env-default:""`
}

func (c *RedisConfig) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

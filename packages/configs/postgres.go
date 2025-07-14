package configs

import (
	"fmt"
	"time"
)

type DatabaseConfig struct {
	User            string        `yaml:"user"              env:"DB_USER"              env-default:"user"`
	Password        string        `yaml:"password"          env:"DB_PASSWORD"          env-default:"pass"`
	Name            string        `yaml:"name"              env:"DB_NAME"              env-default:"db"`
	Host            string        `yaml:"host"              env:"DB_HOST"              env-default:"localhost"`
	Port            uint16        `yaml:"port"              env:"DB_PORT"              env-default:"5432"`
	SSLMode         string        `yaml:"ssl_mode"          env:"DB_SSL_MODE"          env-default:"disable"`
	MaxConns        int32         `yaml:"max_conns"         env:"DB_MAX_CONNS"         env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" env-default:"1h"`
}

func (db *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}

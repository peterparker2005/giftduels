package configs

import (
	"fmt"
	"net"
	"strconv"
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
	hostPort := net.JoinHostPort(db.Host, strconv.Itoa(int(db.Port)))
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		db.User, db.Password, hostPort, db.Name, db.SSLMode,
	)
}

package configs

import "fmt"

type DatabaseConfig struct {
	User     string `yaml:"user" env:"DB_USER" env-default:"user"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-default:"pass"`
	Name     string `yaml:"name" env:"DB_NAME" env-default:"db"`
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     uint16 `yaml:"port" env:"DB_PORT" env-default:"5432"`
}

func (db *DatabaseConfig) Address() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		db.User, db.Password, db.Host, db.Port, db.Name,
	)
}

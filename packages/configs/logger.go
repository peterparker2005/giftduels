package configs

// LoggerConfig keeps logger settings.
type LoggerConfig struct {
	LogLevel string `env:"LOG_LEVEL" env-default:"info" yaml:"level"`    // Log level: debug,info,warn,error,fatal
	Pretty   bool   `env:"LOG_PRETTY" env-default:"false" yaml:"pretty"` // Human-friendly (pretty) output
}

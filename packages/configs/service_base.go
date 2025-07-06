package configs

// ServiceBaseConfig keeps base service settings.
type ServiceBaseConfig struct {
	ServiceName string `env:"SERVICE_NAME" env-default:"identity" yaml:"service_name"`
	Environment string `env:"ENVIRONMENT" env-default:"development" yaml:"environment"`
}

package configs

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

type ServiceName string

// ServiceBaseConfig keeps base service settings.
type ServiceBaseConfig struct {
	ServiceName ServiceName `env:"SERVICE_NAME" env-default:"identity"    yaml:"service_name"`
	Environment Environment `env:"ENVIRONMENT"  env-default:"development" yaml:"environment"`
}

func (c Environment) IsDev() bool {
	return c == EnvironmentDevelopment
}

func (c Environment) IsProd() bool {
	return c == EnvironmentProduction
}

func (c Environment) String() string {
	return string(c)
}

func (c ServiceName) String() string {
	return "service-" + string(c)
}

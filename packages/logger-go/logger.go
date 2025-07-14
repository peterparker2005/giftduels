package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config is the common logger settings.
type Config struct {
	Service     string // required
	Level       string // debug|info|warn|error|fatal
	Pretty      bool   // human-readable output
	Environment string // development|staging|production (optional)
	Version     string // service version (optional)
}

// Logger is a thin wrapper over zap.Logger with cached baseFields.
type Logger struct {
	zap        *zap.Logger
	baseFields []zap.Field
}

// NewLogger configures zap-config and immediately calculates baseFields.
func NewLogger(cfg Config) (*Logger, error) {
	var zapCfg zap.Config
	if cfg.Pretty {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Encoding = "console"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapCfg.EncoderConfig.ConsoleSeparator = " "
	} else {
		zapCfg = zap.NewProductionConfig()
		zapCfg.Encoding = "json"
		zapCfg.OutputPaths = []string{"stdout"}
		zapCfg.ErrorOutputPaths = []string{"stderr"}

		enc := &zapCfg.EncoderConfig
		enc.TimeKey = "timestamp"
		enc.EncodeTime = zapcore.RFC3339TimeEncoder
		enc.LevelKey = "level"
		enc.EncodeLevel = zapcore.LowercaseLevelEncoder
		enc.MessageKey = "message"
		enc.CallerKey = ""
		enc.StacktraceKey = ""
	}

	if lvl, err := zap.ParseAtomicLevel(cfg.Level); err == nil {
		zapCfg.Level = lvl
	}

	zl, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	base := []zap.Field{
		zap.String("service", cfg.Service),
	}
	if cfg.Environment != "" {
		base = append(base, zap.String("environment", cfg.Environment))
	}
	if cfg.Version != "" {
		base = append(base, zap.String("version", cfg.Version))
	}

	return &Logger{
		zap:        zl,
		baseFields: base,
	}, nil
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, append(l.baseFields, fields...)...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, append(l.baseFields, fields...)...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, append(l.baseFields, fields...)...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, append(l.baseFields, fields...)...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, append(l.baseFields, fields...)...)
}

// With returns a new Logger with additional fields, but preserves baseFields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		zap:        l.zap.With(fields...),
		baseFields: l.baseFields,
	}
}

// Sync is called before shutdown.
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

func (l *Logger) Zap() *zap.Logger {
	return l.zap
}

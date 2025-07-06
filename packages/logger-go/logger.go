package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config — общие настройки логгера.
type Config struct {
	Service     string // обязательное
	Level       string // debug|info|warn|error|fatal
	Pretty      bool   // вывод в человекочитаемом виде
	Environment string // development|staging|production (опционально)
	Version     string // версия сервиса (опционально)
}

// Logger — thin wrapper над zap.Logger с кэшированными baseFields.
type Logger struct {
	zap        *zap.Logger
	baseFields []zap.Field
}

// NewLogger настраивает zap-конфиг и сразу вычисляет baseFields.
func NewLogger(cfg Config) (*Logger, error) {
	// 1) Собираем zap.Config
	var zapCfg zap.Config
	if cfg.Pretty {
		zapCfg = zap.NewDevelopmentConfig()
		// Для pretty логирования используем консольный энкодер
		zapCfg.Encoding = "console"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapCfg.EncoderConfig.ConsoleSeparator = " "
	} else {
		zapCfg = zap.NewProductionConfig()
		zapCfg.Encoding = "json"                     // Структурированный JSON
		zapCfg.OutputPaths = []string{"stdout"}      // один источник — STDOUT
		zapCfg.ErrorOutputPaths = []string{"stderr"} // ошибки во STDERR

		enc := &zapCfg.EncoderConfig
		enc.TimeKey = "timestamp"
		enc.EncodeTime = zapcore.RFC3339TimeEncoder
		enc.LevelKey = "level"
		enc.EncodeLevel = zapcore.LowercaseLevelEncoder
		enc.MessageKey = "message"
		enc.CallerKey = ""     // не выводим caller
		enc.StacktraceKey = "" // не выводим stacktrace
	}

	if lvl, err := zap.ParseAtomicLevel(cfg.Level); err == nil {
		zapCfg.Level = lvl
	}

	// 2) Строим сам Logger
	zl, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	// 3) Вычисляем неизменяемые базовые поля
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

// Info/Debug/Warn/Error/Fatal работают через variadic zap.Field.
// Всегда подтягивают baseFields первым аргументом.
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

// With возвращает новый Logger с доп. полями, но сохраняет baseFields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		zap:        l.zap.With(fields...),
		baseFields: l.baseFields,
	}
}

// Sync — не забывайте вызывать перед shutdown.
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

func (l *Logger) Zap() *zap.Logger {
	return l.zap
}

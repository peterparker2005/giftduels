package logger

import (
	"github.com/ThreeDotsLabs/watermill"
	"go.uber.org/zap"
)

// WatermillLogger is a zap logger adapter for Watermill.
type WatermillLogger struct {
	logger *Logger
}

// NewWatermill creates a new logger adapter.
func NewWatermill(logger *Logger) watermill.LoggerAdapter {
	return &WatermillLogger{
		logger: logger,
	}
}

func (l *WatermillLogger) Error(msg string, err error, fields watermill.LogFields) {
	l.logger.With(l.convertFields(fields)...).With(zap.Error(err)).Error(msg)
}

func (l *WatermillLogger) Info(msg string, fields watermill.LogFields) {
	l.logger.With(l.convertFields(fields)...).Info(msg)
}

func (l *WatermillLogger) Debug(msg string, fields watermill.LogFields) {
	l.logger.With(l.convertFields(fields)...).Debug(msg)
}

func (l *WatermillLogger) Trace(msg string, fields watermill.LogFields) {
	l.logger.With(l.convertFields(fields)...).Debug(msg)
}

func (l *WatermillLogger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return &WatermillLogger{
		logger: l.logger.With(l.convertFields(fields)...),
	}
}

func (l *WatermillLogger) convertFields(fields watermill.LogFields) []zap.Field {
	var zapFields []zap.Field
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return zapFields
}

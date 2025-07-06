package logger

import (
	"github.com/ThreeDotsLabs/watermill"
	"go.uber.org/zap"
)

// WatermillLogger адаптер zap логгера для Watermill
type WatermillLogger struct {
	logger *zap.SugaredLogger
}

// NewWatermillLogger создает новый адаптер логгера
func NewWatermill(logger *zap.Logger) watermill.LoggerAdapter {
	return &WatermillLogger{
		logger: logger.WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
}

func (l *WatermillLogger) Error(msg string, err error, fields watermill.LogFields) {
	l.logger.With(l.convertFields(fields)...).With("error", err).Error(msg)
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

func (l *WatermillLogger) convertFields(fields watermill.LogFields) []interface{} {
	var zapFields []interface{}
	for key, value := range fields {
		zapFields = append(zapFields, key, value)
	}
	return zapFields
}

package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const (
	TraceIDKey   ctxKey = "x-trace-id"
	RequestIDKey ctxKey = "x-request-id"
)

// WithContext returns a new Logger with the trace_id and request_id fields
// added from ctx (if they are present).
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := []zap.Field{}
	if tid, ok := ctx.Value(TraceIDKey).(string); ok && tid != "" {
		fields = append(fields, zap.String("trace_id", tid))
	}
	if rid, ok := ctx.Value(RequestIDKey).(string); ok && rid != "" {
		fields = append(fields, zap.String("request_id", rid))
	}
	// create a new zap.Logger with these fields
	return &Logger{
		zap:        l.zap.With(fields...),
		baseFields: l.baseFields,
	}
}

package logger

import "go.uber.org/zap"

type LogErrorDetail struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
	Stack   string `json:"stack,omitempty"`
	Code    string `json:"code,omitempty"`
}

// Поле trace_id (32 hex)
func TraceID(v string) zap.Field { return zap.String("trace_id", v) }

// Поле span_id (16 hex)
func SpanID(v string) zap.Field { return zap.String("span_id", v) }

// Поле request_id (UUID)
func RequestID(v string) zap.Field { return zap.String("request_id", v) }

// Поле user_id (идентификатор пользователя)
func UserID(v string) zap.Field { return zap.String("user_id", v) }

// Поле session_id (идентификатор сессии)
func SessionID(v string) zap.Field { return zap.String("session_id", v) }

// Поле operation
func Operation(v string) zap.Field { return zap.String("operation", v) }

// Поле duration_ms (>=0)
func DurationMs(v float64) zap.Field {
	if v < 0 {
		v = 0
	}
	return zap.Float64("duration_ms", v)
}

// Поле status_code
func StatusCode(v int) zap.Field { return zap.Int("status_code", v) }

// Поле metadata (любые доп. пары key–value)
func Metadata(m map[string]interface{}) zap.Field { return zap.Any("metadata", m) }

// Поле error (объект с деталями)
func ErrorDetail(e LogErrorDetail) zap.Field { return zap.Any("error", e) }

// Поле error (простая ошибка)
func Error(err error) zap.Field { return zap.Error(err) }

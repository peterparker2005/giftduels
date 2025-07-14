package logger

import "go.uber.org/zap"

// RequestID is the request_id field.
func RequestID(v string) zap.Field { return zap.String("request_id", v) }

// TelegramUserID is the user_id field.
func TelegramUserID(v string) zap.Field { return zap.String("telegram_user_id", v) }

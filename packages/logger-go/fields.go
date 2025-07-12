package logger

import "go.uber.org/zap"

// Поле request_id (UUID)
func RequestID(v string) zap.Field { return zap.String("request_id", v) }

// Поле user_id (идентификатор пользователя)
func TelegramUserID(v string) zap.Field { return zap.String("telegram_user_id", v) }

package jwtutil

import (
	"github.com/golang-jwt/jwt/v5"
)

type SessionClaims struct {
	jwt.RegisteredClaims

	UserID         string `json:"uid"`
	TelegramUserID int64  `json:"telegram_user_id"`
}

func ParseToken(tokenStr string, secret string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok {
		return nil, jwt.ErrInvalidKey
	}

	return claims, nil
}

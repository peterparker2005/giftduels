package token

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/peterparker2005/giftduels/apps/service-identity/internal/config"
)

type SessionClaims struct {
	jwt.RegisteredClaims

	TelegramUserID int64 `json:"tg_uid"`
}

type Service interface {
	Generate(telegramUserID int64) (string, error)
	Validate(tokenStr string) (*SessionClaims, error)
}

type JWTService struct {
	secret     []byte
	expiration time.Duration
	logger     *zap.Logger
}

func NewJWTService(cfg *config.JWTConfig, logger *zap.Logger) *JWTService {
	return &JWTService{
		secret:     []byte(cfg.Secret),
		expiration: cfg.Expiration,
		logger:     logger,
	}
}

func (s *JWTService) Generate(telegramUserID int64) (string, error) {
	claims := &SessionClaims{
		TelegramUserID: telegramUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(telegramUserID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) Validate(tokenStr string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return s.secret, nil
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

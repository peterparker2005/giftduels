package duel

import (
	"errors"
)

var (
	ErrInvalidID             = errors.New("id is required")
	ErrInvalidMaxPlayers     = errors.New("max players must be between 2 and 4")
	ErrInvalidMaxGifts       = errors.New("max gifts must be between 1 and 10")
	ErrInvalidTelegramUserID = errors.New("telegram user id must be greater than 0")
)

type MaxPlayers int32

const (
	MinPlayers      = 2
	MaxPlayersLimit = 4
)

func NewMaxPlayers(n int32) (MaxPlayers, error) {
	if n < MinPlayers || n > MaxPlayersLimit {
		return 0, ErrInvalidMaxPlayers
	}
	return MaxPlayers(n), nil
}

type MaxGifts int32

const (
	MinGifts      = 1
	MaxGiftsLimit = 10
)

func NewMaxGifts(n int32) (MaxGifts, error) {
	if n < MinGifts || n > MaxGiftsLimit {
		return 0, ErrInvalidMaxGifts
	}
	return MaxGifts(n), nil
}

func NewID(id string) (ID, error) {
	if id == "" {
		return "", ErrInvalidID
	}
	return ID(id), nil
}

func NewTelegramUserID(id int64) (TelegramUserID, error) {
	if id <= 0 {
		return 0, ErrInvalidTelegramUserID
	}
	return TelegramUserID(id), nil
}

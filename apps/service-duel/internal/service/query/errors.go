package query

import "errors"

var (
	ErrDuelNotFound           = errors.New("duel not found")
	ErrInvalidFilterType      = errors.New("invalid filter type")
	ErrGetGifts               = errors.New("failed to fetch gifts")
	ErrGiftNotFoundInResponse = errors.New("gift not found in response")
	ErrParseGiftPrice         = errors.New("failed to parse gift price")
	ErrBuildStakedGift        = errors.New("failed to build staked gift")
	ErrGetUsers               = errors.New("failed to fetch users")
	ErrDatabase               = errors.New("database operation failed")
)

func IsDuelNotFound(err error) bool {
	return errors.Is(err, ErrDuelNotFound)
}

func IsDatabase(err error) bool {
	return errors.Is(err, ErrDatabase)
}

func IsInvalidFilterType(err error) bool {
	return errors.Is(err, ErrInvalidFilterType)
}

func IsGetGifts(err error) bool {
	return errors.Is(err, ErrGetGifts)
}

func IsGiftNotFoundInResponse(err error) bool {
	return errors.Is(err, ErrGiftNotFoundInResponse)
}

func IsParseGiftPrice(err error) bool {
	return errors.Is(err, ErrParseGiftPrice)
}

func IsBuildStakedGift(err error) bool {
	return errors.Is(err, ErrBuildStakedGift)
}

func IsGetUsers(err error) bool {
	return errors.Is(err, ErrGetUsers)
}

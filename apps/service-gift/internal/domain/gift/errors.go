package gift

import "errors"

var (
	ErrGiftAlreadyWithdrawn      = errors.New("gift already withdrawn")
	ErrGiftNotOwned              = errors.New("gift not owned")
	ErrGiftNotFound              = errors.New("gift not found")
	ErrGiftAlreadyOwned          = errors.New("gift already owned")
	ErrTitleRequired             = errors.New("title is required")
	ErrSlugRequired              = errors.New("slug is required")
	ErrGiftNotWithdrawPending    = errors.New("gift is not in withdraw pending status")
	ErrGiftCannotBeWithdrawn     = errors.New("gift cannot be withdrawn")
	ErrGiftCannotStake           = errors.New("gift cannot be staked")
	ErrInvalidCommissionCurrency = errors.New("invalid commission currency")
)

func IsInvalidCommissionCurrency(err error) bool {
	return errors.Is(err, ErrInvalidCommissionCurrency)
}

func IsGiftNotOwned(err error) bool {
	return errors.Is(err, ErrGiftNotOwned)
}

func IsGiftNotFound(err error) bool {
	return errors.Is(err, ErrGiftNotFound)
}

func IsGiftAlreadyWithdrawn(err error) bool {
	return errors.Is(err, ErrGiftAlreadyWithdrawn)
}

func IsGiftAlreadyOwned(err error) bool {
	return errors.Is(err, ErrGiftAlreadyOwned)
}

func IsGiftNotWithdrawPending(err error) bool {
	return errors.Is(err, ErrGiftNotWithdrawPending)
}

func IsGiftCannotBeWithdrawn(err error) bool {
	return errors.Is(err, ErrGiftCannotBeWithdrawn)
}

func IsGiftCannotStake(err error) bool {
	return errors.Is(err, ErrGiftCannotStake)
}

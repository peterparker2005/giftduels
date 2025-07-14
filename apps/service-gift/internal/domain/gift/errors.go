package gift

import "errors"

var (
	ErrGiftAlreadyWithdrawn   = errors.New("gift already withdrawn")
	ErrGiftNotOwned           = errors.New("gift not owned")
	ErrGiftAlreadyOwned       = errors.New("gift already owned")
	ErrTitleRequired          = errors.New("title is required")
	ErrSlugRequired           = errors.New("slug is required")
	ErrGiftNotWithdrawPending = errors.New("gift is not in withdraw pending status")
	ErrGiftCannotBeWithdrawn  = errors.New("gift cannot be withdrawn")
)

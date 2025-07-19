package payment

import "errors"

var ErrInsufficientBalance = errors.New("insufficient balance")

func IsInsufficientBalance(err error) bool {
	return errors.Is(err, ErrInsufficientBalance)
}

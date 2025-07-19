package proto

import "errors"

var ErrInvalidFilterType = errors.New("invalid filter type")

func IsInvalidFilterType(err error) bool {
	return errors.Is(err, ErrInvalidFilterType)
}

package errors

import "errors"

var ErrRequiredField = errors.New("required enum field is unspecified")

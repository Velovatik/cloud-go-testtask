package entity

import "errors"

var (
	ErrNullEntity = errors.New("cannot set entity to nil")
)

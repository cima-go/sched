package sched

import (
	"errors"
)

var (
	ErrTypeNotRegistered  = errors.New("type has not registered")
	ErrTypeBeenRegistered = errors.New("type has been registered")
)

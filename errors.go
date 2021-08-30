package sched

import (
	"errors"
)

var (
	ErrTypeNotRegister    = errors.New("type has not register")
	ErrTypeBeenRegistered = errors.New("type has been registered")
)

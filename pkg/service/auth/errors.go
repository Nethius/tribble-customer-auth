package auth

import (
	"errors"
)

var (
	ErrAlreadyRegistered  = errors.New("user already registered")
	ErrNotExist           = errors.New("user does not exist")
	ErrWrongPassword      = errors.New("wrong password")
	ErrMissingTokenSecret = errors.New("missing token secret")
)

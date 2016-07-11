package servicer

import (
	"errors"
)

var (
	UserNotFound         = errors.New("user not found")
	CustomerNotFound     = errors.New("customer not found")
	TeamNotFound         = errors.New("team not found")
	SignupNotFound       = errors.New("customer not found")
	SignupExists         = errors.New("signup exists with that email")
	SignupAlreadyClaimed = errors.New("signup already claimed")
	SignupInvalidToken   = errors.New("invalid token for signup")
)

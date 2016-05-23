package gesclient

import "errors"

var (
	WrongExpectedVersion = errors.New("Wrong expected version")
	StreamDeleted        = errors.New("Stream deleted")
	InvalidTransaction   = errors.New("Invalid transaction")
	AccessDenied         = errors.New("Access denied")
	AuthenticationError  = errors.New("Authentication error")
	BadRequest           = errors.New("Bad request")
)

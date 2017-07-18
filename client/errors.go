package client

import (
	"errors"
	"fmt"
)

var (
	WrongExpectedVersion = errors.New("Wrong expected version")
	StreamDeleted        = errors.New("Stream deleted")
	InvalidTransaction   = errors.New("Invalid transaction")
	AccessDenied         = errors.New("Access denied")
	AuthenticationError  = errors.New("Authentication error")
	BadRequest           = errors.New("Bad request")
)

type ServerError struct {
	msg string
}

func NewServerError(msg string) error {
	if msg == "" {
		msg = "<no message>"
	}
	return &ServerError{msg}
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("Unexpected error on server: %s", e.msg)
}

type NotModified struct {
	stream string
}

func NewNotModified(stream string) error {
	return &NotModified{stream}
}

func (e *NotModified) Error() string {
	return fmt.Sprintf("Stream not modified: %s", e.stream)
}

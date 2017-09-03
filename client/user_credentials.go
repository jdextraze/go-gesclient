package client

import "fmt"

type UserCredentials struct {
	username string
	password string
}

func NewUserCredentials(username string, password string) *UserCredentials {
	return &UserCredentials{username, password}
}

func (uc *UserCredentials) Username() string {
	return uc.username
}

func (uc *UserCredentials) Password() string {
	return uc.password
}

func (uc *UserCredentials) String() string {
	return fmt.Sprintf("UserCredentials{username: %s password: %s}", uc.username, uc.password)
}

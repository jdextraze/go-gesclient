package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	"testing"
)

func TestWrongExpectedVersion_Error(t *testing.T) {
	if client.WrongExpectedVersion.Error() != "Wrong expected version" {
		t.FailNow()
	}
}

func TestStreamDeleted_Error(t *testing.T) {
	if client.StreamDeleted.Error() != "Stream deleted" {
		t.FailNow()
	}
}

func TestInvalidTransaction_Error(t *testing.T) {
	if client.InvalidTransaction.Error() != "Invalid transaction" {
		t.FailNow()
	}
}

func TestAccessDenied_Error(t *testing.T) {
	if client.AccessDenied.Error() != "Access denied" {
		t.FailNow()
	}
}

func TestAuthenticationError_Error(t *testing.T) {
	if client.AuthenticationError.Error() != "Authentication error" {
		t.FailNow()
	}
}

func TestBadRequest_Error(t *testing.T) {
	if client.BadRequest.Error() != "Bad request" {
		t.FailNow()
	}
}

func TestServerError_Error(t *testing.T) {
	err := client.NewServerError("")
	if err.Error() != "Unexpected error on server: <no message>" {
		t.FailNow()
	}

	err = client.NewServerError("test")
	if err.Error() != "Unexpected error on server: test" {
		t.FailNow()
	}
}

func TestNotModified_Error(t *testing.T) {
	err := client.NewNotModified("test")
	if err.Error() != "Stream not modified: test" {
		t.FailNow()
	}
}

package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	"testing"
)

func TestNewPosition(t *testing.T) {
	p := client.NewPosition(10, 10)
	if p == nil {
		t.Fatalf("NewPosition returned nil")
	}
	if p.CommitPosition() != 10 {
		t.Error("CommitPosition returns doesn't match")
	}
	if p.PreparePosition() != 10 {
		t.Error("PreparePosition returns doesn't match")
	}

	expectToPanic(t, func() { client.NewPosition(10, 11) },
		"The commit position cannot be less than the prepare position (10 < 11)")
}

func TestPosition_Equals(t *testing.T) {
	p1 := client.NewPosition(10, 10)
	if !p1.Equals(client.NewPosition(10, 10)) {
		t.Error("Position Equals should be true")
	}
	if p1.Equals(client.NewPosition(9, 9)) {
		t.Error("Position Equals should be false")
	}
}

func TestPosition_GreaterThan(t *testing.T) {
	p1 := client.NewPosition(10, 10)
	if !p1.GreaterThan(client.NewPosition(9, 9)) {
		t.Error("Position GreaterThan should be true")
	}
	if p1.GreaterThan(client.NewPosition(10, 10)) {
		t.Error("Position GreaterThan should be false")
	}
}

func TestPosition_GreaterThanOrEquals(t *testing.T) {
	p1 := client.NewPosition(10, 10)
	if !p1.GreaterThanOrEquals(client.NewPosition(9, 9)) {
		t.Error("Position GreaterThanOrEquals should be true")
	}
	if !p1.GreaterThanOrEquals(client.NewPosition(10, 10)) {
		t.Error("Position GreaterThanOrEquals should be true")
	}
	if p1.GreaterThanOrEquals(client.NewPosition(11, 11)) {
		t.Error("Position GreaterThanOrEquals should be false")
	}
}

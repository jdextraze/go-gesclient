package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"testing"
)

func TestNewStreamEventsSlice(t *testing.T) {
	s := client.NewStreamEventsSlice(
		client.SliceReadStatus_Success,
		"Test",
		1,
		client.ReadDirection_Forward,
		[]*messages.ResolvedIndexedEvent{{}},
		12,
		123,
		true,
	)
	if s == nil {
		t.Fatal("NewStreamEventsSlice returned nil")
	}

	if s.IsEndOfStream() != true {
		t.Error("IsEndOfStream")
	}
	if s.Stream() != "Test" {
		t.Error("Stream")
	}
	if s.Status() != client.SliceReadStatus_Success {
		t.Error("Status")
	}
	if len(s.Events()) != 1 {
		t.Error("Events")
	}
	if s.ReadDirection() != client.ReadDirection_Forward {
		t.Error("ReadDirection")
	}
	if s.FromEventNumber() != 1 {
		t.Error("FromEventNumber")
	}
	if s.LastEventNumber() != 123 {
		t.Error("LastEventNumber")
	}
	if s.NextEventNumber() != 12 {
		t.Error("NextEventNumber")
	}
}

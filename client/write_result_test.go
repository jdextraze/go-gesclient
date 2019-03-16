package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	"testing"
)

func TestNewWriteResult(t *testing.T) {
	var wr *client.WriteResult
	wr = client.NewWriteResult(12, nil)
	if wr == nil {
		t.Fail()
	}
}

func TestWriteResult_NextExpectedVersion(t *testing.T) {
	wr := client.NewWriteResult(12, nil)
	if wr.NextExpectedVersion() != 12 {
		t.Fail()
	}
}

func TestWriteResult_LogPosition(t *testing.T) {
	p := client.NewPosition(1, 1)
	wr := client.NewWriteResult(12, p)
	if wr.LogPosition() != p {
		t.Fail()
	}
}

func TestWriteResult_String(t *testing.T) {
	wr := client.NewWriteResult(12, nil)
	if wr.String() != "&{nextExpectedVersion:12 logPosition:<nil>}" {
		t.Fail()
	}
}
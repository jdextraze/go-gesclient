package client_test

import (
	"bytes"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/guid"
	"github.com/satori/go.uuid"
	"testing"
)

func expectToPanic(t *testing.T, fn func(), v interface{}) {
	defer func() {
		if recover() != v {
			t.FailNow()
		}
	}()
	fn()
}

func TestNewTcpPackage(t *testing.T) {
	cmd := client.Command_BadRequest
	flags := client.FlagsNone
	correlationId := uuid.Must(uuid.NewV4())
	data := []byte(`data`)
	userCredentials := client.NewUserCredentials("1", "2")

	p := client.NewTcpPackage(cmd, flags, correlationId, data, nil)
	if p == nil {
		t.Fatalf("NewTcpPackage returned nil")
	}

	if p.Command() != cmd {
		t.Errorf("Package command doesn't match: %v != %v", p.Command(), cmd)
	}
	if p.Flags() != client.FlagsNone {
		t.Errorf("Package flags doesn't match: %v != %v", p.Flags(), client.FlagsNone)
	}
	if !uuid.Equal(p.CorrelationId(), correlationId) {
		t.Errorf("Package correlationId doesn't match: %s != %s", p.CorrelationId(), correlationId)
	}
	if p.Size() != 22 {
		t.Errorf("Package size doesn't match: %d != %d", p.Size(), 22)
	}
	if p.Username() != "" {
		t.Errorf("Package username doesn't match: %s != %s", p.Username(), "")
	}
	if p.Password() != "" {
		t.Errorf("Package password doesn't match: %s != %s", p.Password(), "")
	}
	if !bytes.Equal(p.Data(), data) {
		t.Errorf("Package data doesn't match: %v != %v", p.Data(), data)
	}

	expectToPanic(t, func() {
		client.NewTcpPackage(cmd, client.FlagsNone, correlationId, data, userCredentials)
	}, "userCredentials provided for non-authorized TcpPackage.")

	expectToPanic(t, func() {
		client.NewTcpPackage(cmd, client.FlagsAuthenticated, correlationId, data, nil)
	}, "userCredentials are missing")

	expectToPanic(t, func() {
		client.NewTcpPackage(cmd, client.FlagsAuthenticated, correlationId, data, client.NewUserCredentials("", "pwd"))
	}, "username is missing")

	expectToPanic(t, func() {
		client.NewTcpPackage(cmd, client.FlagsAuthenticated, correlationId, data, client.NewUserCredentials("user", ""))
	}, "password is missing")

	p = client.NewTcpPackage(cmd, client.FlagsAuthenticated, correlationId, data, userCredentials)
	if p.Command() != cmd {
		t.Errorf("Package command doesn't match: %v != %v", p.Command(), cmd)
	}
	if p.Flags() != client.FlagsAuthenticated {
		t.Errorf("Package flags doesn't match: %v != %v", p.Flags(), client.FlagsAuthenticated)
	}
	if !uuid.Equal(p.CorrelationId(), correlationId) {
		t.Errorf("Package correlationId doesn't match: %s != %s", p.CorrelationId(), correlationId)
	}
	if p.Size() != 26 {
		t.Errorf("Package size doesn't match: %d != %d", p.Size(), 26)
	}
	if p.Username() != userCredentials.Username() {
		t.Errorf("Package username doesn't match: %s != %s", p.Username(), userCredentials.Username())
	}
	if p.Password() != userCredentials.Password() {
		t.Errorf("Package password doesn't match: %s != %s", p.Password(), userCredentials.Password())
	}
	if !bytes.Equal(p.Data(), data) {
		t.Errorf("Package data doesn't match: %v != %v", p.Data(), data)
	}
}

func TestTcpPacketFromBytes(t *testing.T) {
	correlationId := uuid.Must(uuid.NewV4())
	data := []byte{1,2,3}
	b := make([]byte, 21)
	b[client.PackageCommandOffset] = byte(client.Command_BadRequest)
	b[client.PackageFlagsOffset] = byte(client.FlagsNone)
	copy(b[client.PackageCorrelationOffset:], correlationId.Bytes())
	copy(b[client.PackageMandatorySize:], data)
	p, err := client.TcpPacketFromBytes(b)
	if err != nil {
		t.Fatalf("TcpPacketFromBytes failed: %v", err)
	}

	if p.Command() != client.Command_BadRequest {
		t.Errorf("Package command doesn't match: %v != %v", p.Command(), client.Command_BadRequest)
	}
	if p.Flags() != client.FlagsNone {
		t.Errorf("Package flags doesn't match: %v != %v", p.Flags(), client.FlagsNone)
	}
	if !uuid.Equal(p.CorrelationId(), guid.FromBytes(correlationId.Bytes())) {
		t.Errorf("Package correlationId doesn't match: %s != %s", p.CorrelationId(), correlationId)
	}
	if p.Size() != int32(len(b)) {
		t.Errorf("Package size doesn't match: %d != %d", p.Size(), len(b))
	}
	if p.Username() != "" {
		t.Errorf("Package username doesn't match: %s != %s", p.Username(), "")
	}
	if p.Password() != "" {
		t.Errorf("Package password doesn't match: %s != %s", p.Password(), "")
	}
	if !bytes.Equal(p.Data(), data) {
		t.Errorf("Package data doesn't match: %v != %v", p.Data(), data)
	}

	b[client.PackageFlagsOffset] = byte(client.FlagsAuthenticated)
	b[client.PackageMandatorySize] = 3
	p, err = client.TcpPacketFromBytes(b)
	if err == nil || err.Error() != "Username length is too big, it does not fit into TcpPackage." {
		t.Fail()
	}

	b[client.PackageMandatorySize] = 1
	b[client.PackageMandatorySize+2] = 3
	p, err = client.TcpPacketFromBytes(b)
	if err == nil || err.Error() != "Password length is too big, it does not fit into TcpPackage." {
		t.Fail()
	}
}
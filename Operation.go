package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"reflect"
)

type operation interface {
	GetCorrelationId() uuid.UUID
	GetRequestCommand() tcpCommand
	GetRequestMessage() proto.Message
	ParseResponse(p *tcpPacket)
	IsCompleted() bool
	Fail(err error)
	Retry() bool
	UserCredentials() *UserCredentials
}

type baseOperation struct {
	correlationId   uuid.UUID
	isCompleted     bool
	retry           bool
	userCredentials *UserCredentials
}

func (o *baseOperation) GetCorrelationId() uuid.UUID { return o.correlationId }

func (o *baseOperation) IsCompleted() bool { return o.isCompleted }

func (o *baseOperation) Retry() bool { return o.retry }

func (o *baseOperation) UserCredentials() *UserCredentials { return o.userCredentials }

func (o *baseOperation) HandleError(p *tcpPacket, expectedCommand tcpCommand) error {
	switch p.Command {
	case tcpCommand_NotAuthenticated:
		return o.handleNotAuthenticated(p)
	case tcpCommand_BadRequest:
		return o.handleBadRequest(p)
	case tcpCommand_NotHandled:
		return o.handleNotHandled(p)
	default:
		return o.handleUnexpectedCommand(p, expectedCommand)
	}
}

func (o *baseOperation) handleNotAuthenticated(p *tcpPacket) error {
	msg := string(p.Payload)
	if msg == "" {
		return AuthenticationError
	}
	return errors.New(msg)
}

func (o *baseOperation) handleBadRequest(p *tcpPacket) error {
	msg := string(p.Payload)
	if msg == "" {
		return BadRequest
	}
	return errors.New(msg)
}

func (o *baseOperation) handleNotHandled(p *tcpPacket) (err error) {
	msg := &protobuf.NotHandled{}
	if err := proto.Unmarshal(p.Payload, msg); err != nil {
		return fmt.Errorf("Invalid payload for NotHandled: %v", err)
	}
	switch *msg.Reason {
	case protobuf.NotHandled_NotReady:
		o.retry = true
	case protobuf.NotHandled_TooBusy:
		o.retry = true
	case protobuf.NotHandled_NotMaster:
		// TODO support reconnect
		err = errors.New("NotHandled - NotMaster not supported")
	default:
		log.Error("Unknown NotHandledReason: %s", msg.Reason.String())
		o.retry = true
	}
	return
}

func (o *baseOperation) handleUnexpectedCommand(p *tcpPacket, expectedCommand tcpCommand) error {
	log.Error(`Unexpected TcpCommand received.
Expected: %v, Actual: %v, CorrelationId: %v
Operation (%s): %v,
TcpPackage data dump: %v`,
		expectedCommand, p.Command, p.CorrelationId, reflect.TypeOf(o).Name(), o, p.Payload,
	)
	return fmt.Errorf("Command not expected. Expected: %v, Actual: %v", expectedCommand, p.Command)
}

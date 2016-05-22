package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"reflect"
)

type Operation interface {
	GetCorrelationId() uuid.UUID
	GetRequestCommand() tcpCommand
	GetRequestMessage() proto.Message
	ParseResponse(p *tcpPacket)
	IsCompleted() bool
	Fail(err error)
	Retry() bool
	UserCredentials() *UserCredentials
}

type BaseOperation struct {
	correlationId   uuid.UUID
	isCompleted     bool
	retry           bool
	userCredentials *UserCredentials
}

func (o *BaseOperation) GetCorrelationId() uuid.UUID { return o.correlationId }

func (o *BaseOperation) IsCompleted() bool { return o.isCompleted }

func (o *BaseOperation) Retry() bool { return o.retry }

func (o *BaseOperation) UserCredentials() *UserCredentials { return o.userCredentials }

func (o *BaseOperation) HandleError(p *tcpPacket, expectedCommand tcpCommand) error {
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

func (o *BaseOperation) handleNotAuthenticated(p *tcpPacket) error {
	msg := string(p.Payload)
	if msg == "" {
		msg = "Authentication Error"
	}
	return errors.New(msg)
}

func (o *BaseOperation) handleBadRequest(p *tcpPacket) error {
	msg := string(p.Payload)
	if msg == "" {
		msg = "<no message>"
	}
	return errors.New(msg)
}

func (o *BaseOperation) handleNotHandled(p *tcpPacket) (err error) {
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
		err = errors.New("NotHandled - NotMaster not supported")
	default:
		log.Error("Unknown NotHandledReason: %s", msg.Reason.String())
		o.retry = true
	}
	return
}

func (o *BaseOperation) handleUnexpectedCommand(p *tcpPacket, expectedCommand tcpCommand) error {
	log.Error(`Unexpected TcpCommand received.
Expected: %v, Actual: %v, CorrelationId: %v
Operation (%s): %v,
TcpPackage data dump: %v`,
		expectedCommand, p.Command, p.CorrelationId, reflect.TypeOf(o).Name(), o, p.Payload,
	)
	return fmt.Errorf("Command not expected. Expected: %v, Actual: %v", expectedCommand, p.Command)
}

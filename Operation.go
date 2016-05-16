package gesclient

import (
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type Operation interface {
	GetCorrelationId() uuid.UUID
	GetRequestCommand() tcpCommand
	GetRequestMessage() proto.Message
	ParseResponse(tcpCommand, proto.Message)
	IsCompleted() bool
	SetError(err error) error
}

type BaseOperation struct {
	correlationId uuid.UUID
	isCompleted   bool
}

func (o *BaseOperation) GetCorrelationId() uuid.UUID { return o.correlationId }

func (o *BaseOperation) IsCompleted() bool { return o.isCompleted }

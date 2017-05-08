package models

import (
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type Operation interface {
	GetCorrelationId() uuid.UUID
	GetRequestCommand() Command
	GetRequestMessage() proto.Message
	ParseResponse(p *Package)
	IsCompleted() bool
	Fail(err error)
	Retry() bool
	UserCredentials() *UserCredentials
}

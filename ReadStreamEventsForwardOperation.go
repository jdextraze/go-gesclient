package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type readStreamEventsForwardOperation struct {
	*BaseOperation
	stream string
	start  int
	max    int
	c      chan *StreamEventsSlice
}

func newReadStreamEventsForwardOperation(
	stream string,
	start int,
	max int,
	c chan *StreamEventsSlice,
	userCredentials *UserCredentials,
) *readStreamEventsForwardOperation {
	return &readStreamEventsForwardOperation{
		BaseOperation: &BaseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream: stream,
		start:  start,
		max:    max,
		c:      c,
	}
}

func (o *readStreamEventsForwardOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_ReadStreamEventsForward
}

func (o *readStreamEventsForwardOperation) GetRequestMessage() proto.Message {
	no := false
	start := int32(o.start)
	max := int32(o.max)
	return &protobuf.ReadStreamEvents{
		EventStreamId:   &o.stream,
		FromEventNumber: &start,
		MaxCount:        &max,
		ResolveLinkTos:  &no,
		RequireMaster:   &no,
	}
}

func (o *readStreamEventsForwardOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_ReadStreamEventsForwardCompleted {
		err := o.HandleError(p, tcpCommand_ReadStreamEventsForwardCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}
	msg := &protobuf.ReadStreamEventsCompleted{}
	err := proto.Unmarshal(p.Payload, msg)
	o.c <- newStreamEventsSlice(
		SliceReadStatus(msg.GetResult()),
		o.stream,
		o.start,
		ReadDirectionForward,
		msg.Events,
		int(msg.GetNextEventNumber()),
		int(msg.GetLastEventNumber()),
		msg.GetIsEndOfStream(),
		err,
	)
	close(o.c)
	o.isCompleted = true
}

func (o *readStreamEventsForwardOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.c <- newStreamEventsSlice(
		SliceReadStatusError,
		o.stream,
		o.start,
		ReadDirectionForward,
		[]*protobuf.ResolvedIndexedEvent{},
		0,
		0,
		false,
		err,
	)
	close(o.c)
}

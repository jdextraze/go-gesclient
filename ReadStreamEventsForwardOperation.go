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
) *readStreamEventsForwardOperation {
	return &readStreamEventsForwardOperation{
		BaseOperation: &BaseOperation{
			correlationId: uuid.NewV4(),
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

func (o *readStreamEventsForwardOperation) ParseResponse(cmd tcpCommand, msg proto.Message) {
	readStreamEventsCompleted := msg.(*protobuf.ReadStreamEventsCompleted)
	// TODO handle error
	streamEventsSlice, _ := newStreamEventsSlice(
		SliceReadStatus(readStreamEventsCompleted.GetResult()),
		o.stream,
		o.start,
		ReadDirectionForward,
		readStreamEventsCompleted.Events,
		int(readStreamEventsCompleted.GetNextEventNumber()),
		int(readStreamEventsCompleted.GetLastEventNumber()),
		readStreamEventsCompleted.GetIsEndOfStream(),
	)
	o.c <- streamEventsSlice
	close(o.c)
	o.isCompleted = true
}

func (o *readStreamEventsForwardOperation) SetError(err error) error {
	if o.isCompleted {
		return err
	}
	close(o.c)
	return err
}

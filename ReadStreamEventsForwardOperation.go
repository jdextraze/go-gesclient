package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type readStreamEventsForwardOperation struct {
	*baseOperation
	stream        string
	start         int
	max           int
	resultChannel chan *StreamEventsSlice
}

func newReadStreamEventsForwardOperation(
	stream string,
	start int,
	max int,
	userCredentials *UserCredentials,
) *readStreamEventsForwardOperation {
	return &readStreamEventsForwardOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		start:         start,
		max:           max,
		resultChannel: make(chan *StreamEventsSlice, 1),
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
		err := o.handleError(p, tcpCommand_ReadStreamEventsForwardCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.ReadStreamEventsCompleted{}
	err := proto.Unmarshal(p.Payload, msg)
	if err != nil {
		o.Fail(err)
		return
	}

	switch *msg.Result {
	case protobuf.ReadStreamEventsCompleted_Success:
		o.succeed(msg)
	case protobuf.ReadStreamEventsCompleted_StreamDeleted:
		o.succeed(msg)
	case protobuf.ReadStreamEventsCompleted_NoStream:
		o.succeed(msg)
	case protobuf.ReadStreamEventsCompleted_Error:
		o.Fail(NewServerError(msg.GetError()))
	case protobuf.ReadStreamEventsCompleted_NotModified:
		o.Fail(NewNotModified(o.stream))
	case protobuf.ReadStreamEventsCompleted_AccessDenied:
		o.Fail(AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result))
	}
}

func (o *readStreamEventsForwardOperation) succeed(msg *protobuf.ReadStreamEventsCompleted) {
	o.resultChannel <- newStreamEventsSlice(
		SliceReadStatus(msg.GetResult()),
		o.stream,
		o.start,
		ReadDirectionForward,
		msg.Events,
		int(msg.GetNextEventNumber()),
		int(msg.GetLastEventNumber()),
		msg.GetIsEndOfStream(),
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *readStreamEventsForwardOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newStreamEventsSlice(
		SliceReadStatus_Error,
		o.stream,
		o.start,
		ReadDirectionForward,
		[]*protobuf.ResolvedIndexedEvent{},
		0,
		0,
		false,
		err,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

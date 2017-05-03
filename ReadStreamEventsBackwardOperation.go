package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type readStreamEventsBackwardOperation struct {
	*baseOperation
	stream        string
	start         int
	max           int
	resultChannel chan *StreamEventsSlice
}

func newReadStreamEventsBackwardOperation(
	stream string,
	start int,
	max int,
	userCredentials *UserCredentials,
) *readStreamEventsBackwardOperation {
	return &readStreamEventsBackwardOperation{
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

func (o *readStreamEventsBackwardOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_ReadStreamEventsBackward
}

func (o *readStreamEventsBackwardOperation) GetRequestMessage() proto.Message {
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

func (o *readStreamEventsBackwardOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_ReadStreamEventsBackwardCompleted {
		err := o.handleError(p, tcpCommand_ReadStreamEventsBackwardCompleted)
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

func (o *readStreamEventsBackwardOperation) succeed(msg *protobuf.ReadStreamEventsCompleted) {
	o.resultChannel <- newStreamEventsSlice(
		SliceReadStatus(msg.GetResult()),
		o.stream,
		o.start,
		ReadDirectionBackward,
		msg.Events,
		int(msg.GetNextEventNumber()),
		int(msg.GetLastEventNumber()),
		msg.GetIsEndOfStream(),
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *readStreamEventsBackwardOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newStreamEventsSlice(
		SliceReadStatus_Error,
		o.stream,
		o.start,
		ReadDirectionBackward,
		[]*protobuf.ResolvedIndexedEvent{},
		0,
		0,
		false,
		err,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

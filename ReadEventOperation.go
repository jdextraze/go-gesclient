package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type ReadEventOperation struct {
	*baseOperation
	stream        string
	eventNumber   int
	resolveTos    bool
	resultChannel chan *EventReadResult
}

func newReadEventOperation(
	stream string,
	eventNumber int,
	resolveTos bool,
	userCredentials *UserCredentials,
) *ReadEventOperation {
	return &ReadEventOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		eventNumber:   eventNumber,
		resolveTos:    resolveTos,
		resultChannel: make(chan *EventReadResult, 1),
	}
}

func (o *ReadEventOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_ReadEvent
}

func (o *ReadEventOperation) GetRequestMessage() proto.Message {
	eventNumber := int32(o.eventNumber)
	requireMaster := false
	return &protobuf.ReadEvent{
		EventStreamId:  &o.stream,
		EventNumber:    &eventNumber,
		ResolveLinkTos: &o.resolveTos,
		RequireMaster:  &requireMaster,
	}
}

func (o *ReadEventOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_ReadEventCompleted {
		if err := o.handleError(p, tcpCommand_ReadEventCompleted); err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.ReadEventCompleted{}
	if err := proto.Unmarshal(p.Payload, msg); err != nil {
		o.Fail(err)
		return
	}

	switch *msg.Result {
	case protobuf.ReadEventCompleted_Success:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_NotFound:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_StreamDeleted:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_NoStream:
		o.succeed(msg)
	case protobuf.ReadEventCompleted_Error:
		o.Fail(NewServerError(msg.GetError()))
	case protobuf.ReadEventCompleted_AccessDenied:
		o.Fail(AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected ReadStreamResult: %v", *msg.Result))
	}
}

func (o *ReadEventOperation) succeed(msg *protobuf.ReadEventCompleted) {
	o.resultChannel <- newEventReadResult(
		EventReadStatus(msg.GetResult()),
		o.stream,
		o.eventNumber,
		msg.Event,
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *ReadEventOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newEventReadResult(EventReadStatus_Success, "", -1, nil, err)
	close(o.resultChannel)
	o.isCompleted = true
}

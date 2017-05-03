package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type readAllEventsForwardOperation struct {
	*baseOperation
	pos           *Position
	max           int
	resolveTos    bool
	resultChannel chan *AllEventsSlice
}

func newReadAllEventsForwardOperation(
	pos *Position,
	max int,
	resolveTos bool,
	userCredentials *UserCredentials,
) *readAllEventsForwardOperation {
	return &readAllEventsForwardOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		pos:           pos,
		max:           max,
		resolveTos:    resolveTos,
		resultChannel: make(chan *AllEventsSlice, 1),
	}
}

func (o *readAllEventsForwardOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_ReadAllEventsForward
}

func (o *readAllEventsForwardOperation) GetRequestMessage() proto.Message {
	commitPos := o.pos.CommitPosition()
	preparePos := o.pos.PreparePosition()
	no := false
	max := int32(o.max)
	return &protobuf.ReadAllEvents{
		CommitPosition:  &commitPos,
		PreparePosition: &preparePos,
		MaxCount:        &max,
		ResolveLinkTos:  &no,
		RequireMaster:   &no,
	}
}

func (o *readAllEventsForwardOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_ReadAllEventsForwardCompleted {
		err := o.handleError(p, tcpCommand_ReadAllEventsForwardCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.ReadAllEventsCompleted{}
	err := proto.Unmarshal(p.Payload, msg)
	if err != nil {
		o.Fail(err)
		return
	}

	if msg.Result == nil {
		o.succeed(msg)
		return
	}

	switch *msg.Result {
	case protobuf.ReadAllEventsCompleted_Success:
		o.succeed(msg)
	case protobuf.ReadAllEventsCompleted_Error:
		o.Fail(NewServerError(msg.GetError()))
	case protobuf.ReadAllEventsCompleted_AccessDenied:
		o.Fail(AccessDenied)
	default:
		o.Fail(fmt.Errorf("Unexpected ReadAllResult: %v", *msg.Result))
	}
}

func (o *readAllEventsForwardOperation) succeed(msg *protobuf.ReadAllEventsCompleted) {
	fromPosition, _ := NewPosition(*msg.CommitPosition, *msg.PreparePosition)
	nextPosition, _ := NewPosition(*msg.NextCommitPosition, *msg.NextPreparePosition)
	o.resultChannel <- newAllEventsSlice(
		ReadDirectionForward,
		fromPosition,
		nextPosition,
		msg.Events,
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *readAllEventsForwardOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newAllEventsSlice(
		ReadDirectionForward,
		nil,
		nil,
		[]*protobuf.ResolvedEvent{},
		err,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

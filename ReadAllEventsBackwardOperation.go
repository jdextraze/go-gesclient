package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type readAllEventsBackwardOperation struct {
	*baseOperation
	pos           *Position
	max           int
	resolveTos    bool
	resultChannel chan *AllEventsSlice
}

func newReadAllEventsBackwardOperation(
	pos *Position,
	max int,
	resolveTos bool,
	userCredentials *UserCredentials,
) *readAllEventsBackwardOperation {
	return &readAllEventsBackwardOperation{
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

func (o *readAllEventsBackwardOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_ReadAllEventsBackward
}

func (o *readAllEventsBackwardOperation) GetRequestMessage() proto.Message {
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

func (o *readAllEventsBackwardOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_ReadAllEventsBackwardCompleted {
		err := o.handleError(p, tcpCommand_ReadAllEventsBackwardCompleted)
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

func (o *readAllEventsBackwardOperation) succeed(msg *protobuf.ReadAllEventsCompleted) {
	fromPosition, _ := NewPosition(*msg.CommitPosition, *msg.PreparePosition)
	nextPosition, _ := NewPosition(*msg.NextCommitPosition, *msg.NextPreparePosition)
	o.resultChannel <- newAllEventsSlice(
		ReadDirectionBackward,
		fromPosition,
		nextPosition,
		msg.Events,
		nil,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *readAllEventsBackwardOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newAllEventsSlice(
		ReadDirectionBackward,
		nil,
		nil,
		[]*protobuf.ResolvedEvent{},
		err,
	)
	close(o.resultChannel)
	o.isCompleted = true
}

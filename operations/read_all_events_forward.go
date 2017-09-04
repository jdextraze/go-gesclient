package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type readAllEventsForward struct {
	*baseOperation
	pos        *client.Position
	max        int
	resolveTos bool
}

func NewReadAllEventsForward(
	source *tasks.CompletionSource,
	pos *client.Position,
	max int,
	resolveTos bool,
	userCredentials *client.UserCredentials,
) *readAllEventsForward {
	obj := &readAllEventsForward{
		pos:        pos,
		max:        max,
		resolveTos: resolveTos,
	}
	obj.baseOperation = newBaseOperation(client.Command_ReadAllEventsForward,
		client.Command_ReadAllEventsForwardCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *readAllEventsForward) createRequestDto() proto.Message {
	commitPos := o.pos.CommitPosition()
	preparePos := o.pos.PreparePosition()
	no := false
	max := int32(o.max)
	return &messages.ReadAllEvents{
		CommitPosition:  &commitPos,
		PreparePosition: &preparePos,
		MaxCount:        &max,
		ResolveLinkTos:  &no,
		RequireMaster:   &no,
	}
}

func (o *readAllEventsForward) inspectResponse(message proto.Message) (res *client.InspectionResult, err error) {
	msg := message.(*messages.ReadAllEventsCompleted)
	switch msg.GetResult() {
	case messages.ReadAllEventsCompleted_Success:
		err = o.succeed()
	case messages.ReadAllEventsCompleted_Error:
		err = o.Fail(client.NewServerError(msg.GetError()))
	case messages.ReadAllEventsCompleted_AccessDenied:
		err = o.Fail(client.AccessDenied)
	default:
		err = fmt.Errorf("Unexpected ReadAllResult: %v", *msg.Result)
	}
	if res == nil && err == nil {
		res = client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil)
	}
	return
}

func (o *readAllEventsForward) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*messages.ReadAllEventsCompleted)
	return client.NewAllEventsSlice(
		client.ReadDirection_Forward,
		client.NewPosition(msg.GetCommitPosition(), msg.GetPreparePosition()),
		client.NewPosition(msg.GetNextCommitPosition(), msg.GetNextPreparePosition()),
		msg.Events,
	), nil
}

func (o *readAllEventsForward) createResponse() proto.Message {
	return &messages.ReadAllEventsCompleted{}
}

func (o *readAllEventsForward) String() string {
	return "ReadAllEventsForward"
}

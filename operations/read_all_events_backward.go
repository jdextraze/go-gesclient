package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type readAllEventsBackward struct {
	*baseOperation
	pos        *client.Position
	max        int
	resolveTos bool
}

func NewReadAllEventsBackward(
	source *tasks.CompletionSource,
	pos *client.Position,
	max int,
	resolveTos bool,
	userCredentials *client.UserCredentials,
) *readAllEventsBackward {
	obj := &readAllEventsBackward{
		pos:        pos,
		max:        max,
		resolveTos: resolveTos,
	}
	obj.baseOperation = newBaseOperation(client.Command_ReadAllEventsBackward,
		client.Command_ReadAllEventsBackwardCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *readAllEventsBackward) createRequestDto() proto.Message {
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

func (o *readAllEventsBackward) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
	msg := message.(*messages.ReadAllEventsCompleted)
	switch msg.GetResult() {
	case messages.ReadAllEventsCompleted_Success:
		o.succeed()
	case messages.ReadAllEventsCompleted_Error:
		o.Fail(client.NewServerError(msg.GetError()))
	case messages.ReadAllEventsCompleted_AccessDenied:
		o.Fail(client.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected ReadAllResult: %v", *msg.Result)
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *readAllEventsBackward) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*messages.ReadAllEventsCompleted)
	return client.NewAllEventsSlice(
		client.ReadDirection_Backward,
		client.NewPosition(msg.GetCommitPosition(), msg.GetPreparePosition()),
		client.NewPosition(msg.GetNextCommitPosition(), msg.GetNextPreparePosition()),
		msg.Events,
	), nil
}

func (o *readAllEventsBackward) createResponse() proto.Message {
	return &messages.ReadAllEventsCompleted{}
}

func (o *readAllEventsBackward) String() string {
	return "ReadAllEventsBackward"
}

package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/protobuf"
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
	return &protobuf.ReadAllEvents{
		CommitPosition:  &commitPos,
		PreparePosition: &preparePos,
		MaxCount:        &max,
		ResolveLinkTos:  &no,
		RequireMaster:   &no,
	}
}

func (o *readAllEventsForward) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
	msg := message.(*protobuf.ReadAllEventsCompleted)
	switch msg.GetResult() {
	case protobuf.ReadAllEventsCompleted_Success:
		o.succeed()
	case protobuf.ReadAllEventsCompleted_Error:
		o.Fail(client.NewServerError(msg.GetError()))
	case protobuf.ReadAllEventsCompleted_AccessDenied:
		o.Fail(client.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected ReadAllResult: %v", *msg.Result)
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *readAllEventsForward) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*protobuf.ReadAllEventsCompleted)
	return client.NewAllEventsSlice(
		client.ReadDirectionForward,
		client.NewPosition(msg.GetCommitPosition(), msg.GetPreparePosition()),
		client.NewPosition(msg.GetNextCommitPosition(), msg.GetNextPreparePosition()),
		msg.Events,
	), nil
}

func (o *readAllEventsForward) createResponse() proto.Message {
	return &protobuf.ReadAllEventsCompleted{}
}

func (o *readAllEventsForward) String() string {
	return "ReadAllEventsForward"
}

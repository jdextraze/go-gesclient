package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/tasks"
)

type readAllEventsBackward struct {
	*baseOperation
	pos           *models.Position
	max           int
	resolveTos    bool
}

func NewReadAllEventsBackward(
	source *tasks.CompletionSource,
	pos *models.Position,
	max int,
	resolveTos bool,
	userCredentials *models.UserCredentials,
) *readAllEventsBackward {
	obj := &readAllEventsBackward{
		pos:           pos,
		max:           max,
		resolveTos:    resolveTos,
	}
	obj.baseOperation = newBaseOperation(models.Command_ReadAllEventsBackward,
		models.Command_ReadAllEventsBackwardCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *readAllEventsBackward) createRequestDto() proto.Message {
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

func (o *readAllEventsBackward) inspectResponse(message proto.Message) (*models.InspectionResult, error) {
	msg := message.(*protobuf.ReadAllEventsCompleted)
	switch msg.GetResult() {
	case protobuf.ReadAllEventsCompleted_Success:
		o.succeed()
	case protobuf.ReadAllEventsCompleted_Error:
		o.Fail(models.NewServerError(msg.GetError()))
	case protobuf.ReadAllEventsCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected ReadAllResult: %v", *msg.Result)
	}
	return models.NewInspectionResult(models.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *readAllEventsBackward) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*protobuf.ReadAllEventsCompleted)
	return models.NewAllEventsSlice(
		models.ReadDirectionBackward,
		models.NewPosition(msg.GetCommitPosition(), msg.GetPreparePosition()),
		models.NewPosition(msg.GetNextCommitPosition(), msg.GetNextPreparePosition()),
		msg.Events,
	), nil
}

func (o *readAllEventsBackward) createResponse() proto.Message {
	return &protobuf.ReadAllEventsCompleted{}
}

func (o *readAllEventsBackward) String() string {
	return "ReadAllEventsBackward"
}

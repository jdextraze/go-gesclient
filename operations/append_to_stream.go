package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/tasks"
)

type appendToStream struct {
	*baseOperation
	requireMaster    bool
	events           []*models.EventData
	stream           string
	expectedVersion  int
	wasCommitTimeout bool
}

func NewAppendToStream(
	source *tasks.CompletionSource,
	requireMaster bool,
	stream string,
	expectedVersion int,
	events []*models.EventData,
	userCredentials *models.UserCredentials,
) *appendToStream {
	obj := &appendToStream{
		requireMaster:   requireMaster,
		stream:          stream,
		expectedVersion: expectedVersion,
		events:          events,
	}
	obj.baseOperation = newBaseOperation(models.Command_WriteEvents, models.Command_WriteEventsCompleted,
		userCredentials, source, obj.createRequestDto, obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *appendToStream) createRequestDto() proto.Message {
	newEvents := make([]*protobuf.NewEvent, len(o.events))
	for i, evt := range o.events {
		newEvents[i] = evt.ToNewEvent()
	}
	expectedVersion := int32(o.expectedVersion)
	return &protobuf.WriteEvents{
		EventStreamId:   &o.stream,
		ExpectedVersion: &expectedVersion,
		Events:          newEvents,
		RequireMaster:   &o.requireMaster,
	}
}

func (o *appendToStream) inspectResponse(message proto.Message) (*models.InspectionResult, error) {
	msg := message.(*protobuf.WriteEventsCompleted)
	switch msg.GetResult() {
	case protobuf.OperationResult_Success:
		if o.wasCommitTimeout {
			log.Debugf("IDEMPOTENT WRITE SUCCEEDED FOR %s.", o)
		}
		if err := o.succeed(); err != nil {
			return nil, err
		}
	case protobuf.OperationResult_PrepareTimeout, protobuf.OperationResult_ForwardTimeout:
		return models.NewInspectionResult(models.InspectionDecision_Retry, msg.GetResult().String(), nil, nil), nil
	case protobuf.OperationResult_CommitTimeout:
		o.wasCommitTimeout = true
		return models.NewInspectionResult(models.InspectionDecision_Retry, msg.GetResult().String(), nil, nil), nil
	case protobuf.OperationResult_WrongExpectedVersion:
		o.Fail(models.WrongExpectedVersion)
	case protobuf.OperationResult_StreamDeleted:
		o.Fail(models.StreamDeleted)
	case protobuf.OperationResult_InvalidTransaction:
		o.Fail(models.InvalidTransaction)
	case protobuf.OperationResult_AccessDenied:
		o.Fail(models.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected OperationResult: %s", msg.GetResult())
	}
	return models.NewInspectionResult(models.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *appendToStream) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*protobuf.WriteEventsCompleted)
	pos := models.NewPosition(msg.GetCommitPosition(), msg.GetPreparePosition())
	return models.NewWriteResult(int(msg.GetLastEventNumber()), pos), nil
}

func (o *appendToStream) createResponse() proto.Message {
	return &protobuf.WriteEventsCompleted{}
}

func (o *appendToStream) String() string {
	return fmt.Sprintf("AppendToStream '%s'", o.stream)
}

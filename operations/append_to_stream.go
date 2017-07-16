package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/jdextraze/go-gesclient/tasks"
)

type appendToStream struct {
	*baseOperation
	requireMaster    bool
	events           []*client.EventData
	stream           string
	expectedVersion  int
	wasCommitTimeout bool
}

func NewAppendToStream(
	source *tasks.CompletionSource,
	requireMaster bool,
	stream string,
	expectedVersion int,
	events []*client.EventData,
	userCredentials *client.UserCredentials,
) *appendToStream {
	obj := &appendToStream{
		requireMaster:   requireMaster,
		stream:          stream,
		expectedVersion: expectedVersion,
		events:          events,
	}
	obj.baseOperation = newBaseOperation(client.Command_WriteEvents, client.Command_WriteEventsCompleted,
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

func (o *appendToStream) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
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
		return client.NewInspectionResult(client.InspectionDecision_Retry, msg.GetResult().String(), nil, nil), nil
	case protobuf.OperationResult_CommitTimeout:
		o.wasCommitTimeout = true
		return client.NewInspectionResult(client.InspectionDecision_Retry, msg.GetResult().String(), nil, nil), nil
	case protobuf.OperationResult_WrongExpectedVersion:
		o.Fail(client.WrongExpectedVersion)
	case protobuf.OperationResult_StreamDeleted:
		o.Fail(client.StreamDeleted)
	case protobuf.OperationResult_InvalidTransaction:
		o.Fail(client.InvalidTransaction)
	case protobuf.OperationResult_AccessDenied:
		o.Fail(client.AccessDenied)
	default:
		return nil, fmt.Errorf("Unexpected OperationResult: %s", msg.GetResult())
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *appendToStream) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*protobuf.WriteEventsCompleted)
	pos := client.NewPosition(msg.GetCommitPosition(), msg.GetPreparePosition())
	return client.NewWriteResult(int(msg.GetLastEventNumber()), pos), nil
}

func (o *appendToStream) createResponse() proto.Message {
	return &protobuf.WriteEventsCompleted{}
}

func (o *appendToStream) String() string {
	return fmt.Sprintf("AppendToStream '%s'", o.stream)
}

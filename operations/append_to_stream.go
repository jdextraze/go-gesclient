package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
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
	newEvents := make([]*messages.NewEvent, len(o.events))
	for i, evt := range o.events {
		newEvents[i] = evt.ToNewEvent()
	}
	expectedVersion := int32(o.expectedVersion)
	return &messages.WriteEvents{
		EventStreamId:   &o.stream,
		ExpectedVersion: &expectedVersion,
		Events:          newEvents,
		RequireMaster:   &o.requireMaster,
	}
}

func (o *appendToStream) inspectResponse(message proto.Message) (res *client.InspectionResult, err error) {
	msg := message.(*messages.WriteEventsCompleted)
	switch msg.GetResult() {
	case messages.OperationResult_Success:
		if o.wasCommitTimeout {
			log.Debugf("IDEMPOTENT WRITE SUCCEEDED FOR %s.", o)
		}
		err = o.succeed()
	case messages.OperationResult_PrepareTimeout, messages.OperationResult_ForwardTimeout:
		res = client.NewInspectionResult(client.InspectionDecision_Retry, msg.GetResult().String(), nil, nil)
	case messages.OperationResult_CommitTimeout:
		o.wasCommitTimeout = true
		res = client.NewInspectionResult(client.InspectionDecision_Retry, msg.GetResult().String(), nil, nil)
	case messages.OperationResult_WrongExpectedVersion:
		err = o.Fail(client.WrongExpectedVersion)
	case messages.OperationResult_StreamDeleted:
		err = o.Fail(client.StreamDeleted)
	case messages.OperationResult_InvalidTransaction:
		err = o.Fail(client.InvalidTransaction)
	case messages.OperationResult_AccessDenied:
		err = o.Fail(client.AccessDenied)
	default:
		err = fmt.Errorf("Unexpected OperationResult: %s", msg.GetResult())
	}
	if res == nil && err == nil {
		res = client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil)
	}
	return
}

func (o *appendToStream) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*messages.WriteEventsCompleted)
	pos := client.NewPosition(msg.GetCommitPosition(), msg.GetPreparePosition())
	return client.NewWriteResult(int(msg.GetLastEventNumber()), pos), nil
}

func (o *appendToStream) createResponse() proto.Message {
	return &messages.WriteEventsCompleted{}
}

func (o *appendToStream) String() string {
	return fmt.Sprintf("AppendToStream '%s'", o.stream)
}

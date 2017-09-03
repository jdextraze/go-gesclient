package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type deleteStream struct {
	*baseOperation
	stream          string
	expectedVersion int
	hardDelete      bool
}

func NewDeleteStream(
	source *tasks.CompletionSource,
	stream string,
	expectedVersion int,
	hardDelete bool,
	userCredentials *client.UserCredentials,
) *deleteStream {
	obj := &deleteStream{
		stream:          stream,
		expectedVersion: expectedVersion,
		hardDelete:      hardDelete,
	}
	obj.baseOperation = newBaseOperation(client.Command_DeleteStream, client.Command_DeleteStreamCompleted,
		userCredentials, source, obj.createRequestDto, obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *deleteStream) createRequestDto() proto.Message {
	expectedVersion := int32(o.expectedVersion)
	requireMaster := false
	return &messages.DeleteStream{
		EventStreamId:   &o.stream,
		ExpectedVersion: &expectedVersion,
		RequireMaster:   &requireMaster,
		HardDelete:      &o.hardDelete,
	}
}

func (o *deleteStream) inspectResponse(message proto.Message) (res *client.InspectionResult, err error) {
	msg := message.(*messages.DeleteStreamCompleted)
	switch msg.GetResult() {
	case messages.OperationResult_Success:
		err = o.succeed()
	case messages.OperationResult_PrepareTimeout,
		messages.OperationResult_ForwardTimeout,
		messages.OperationResult_CommitTimeout:
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
		err = fmt.Errorf("Unexpected Operation result: %v", msg.GetResult())
	}
	if res == nil && err == nil {
		res = client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil)
	}
	return
}

func (o *deleteStream) transformResponse(message proto.Message) (interface{}, error) {
	response := message.(*messages.DeleteStreamCompleted)
	return client.NewDeleteResult(client.NewPosition(response.GetCommitPosition(), response.GetPreparePosition())), nil
}

func (o *deleteStream) createResponse() proto.Message {
	return &messages.DeleteStreamCompleted{}
}

func (o *deleteStream) String() string {
	return fmt.Sprintf("DeleteStream '%s'", o.stream)
}

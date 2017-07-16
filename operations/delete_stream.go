package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/protobuf"
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
	return &protobuf.DeleteStream{
		EventStreamId:   &o.stream,
		ExpectedVersion: &expectedVersion,
		RequireMaster:   &requireMaster,
		HardDelete:      &o.hardDelete,
	}
}

func (o *deleteStream) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
	msg := message.(*protobuf.DeleteStreamCompleted)
	switch msg.GetResult() {
	case protobuf.OperationResult_Success:
		if err := o.succeed(); err != nil {
			return nil, err
		}
	case protobuf.OperationResult_PrepareTimeout,
		protobuf.OperationResult_ForwardTimeout,
		protobuf.OperationResult_CommitTimeout:
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
		o.Fail(fmt.Errorf("Unexpected Operation result: %v", msg.GetResult()))
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *deleteStream) transformResponse(message proto.Message) (interface{}, error) {
	response := message.(*protobuf.DeleteStreamCompleted)
	return client.NewDeleteResult(client.NewPosition(response.GetCommitPosition(), response.GetPreparePosition())), nil
}

func (o *deleteStream) createResponse() proto.Message {
	return &protobuf.DeleteStreamCompleted{}
}

func (o *deleteStream) String() string {
	return fmt.Sprintf("DeleteStream '%s'", o.stream)
}

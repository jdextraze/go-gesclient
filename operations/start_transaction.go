package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type StartTransaction struct {
	*baseOperation
	requireMaster    bool
	stream           string
	expectedVersion  int
	parentConnection client.TransactionConnection
}

func NewStartTransaction(
	source *tasks.CompletionSource,
	requireMaster bool,
	stream string,
	expectedVersion int,
	parentConnection client.TransactionConnection,
	userCredentials *client.UserCredentials,
) *StartTransaction {
	obj := &StartTransaction{
		requireMaster:    requireMaster,
		stream:           stream,
		expectedVersion:  expectedVersion,
		parentConnection: parentConnection,
	}
	obj.baseOperation = newBaseOperation(client.Command_TransactionStart, client.Command_TransactionStartCompleted,
		userCredentials, source, obj.createRequestDto, obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *StartTransaction) createRequestDto() proto.Message {
	expectedVersion := int32(o.expectedVersion)
	return &messages.TransactionStart{
		EventStreamId:   &o.stream,
		RequireMaster:   &o.requireMaster,
		ExpectedVersion: &expectedVersion,
	}
}

func (o *StartTransaction) inspectResponse(message proto.Message) (res *client.InspectionResult, err error) {
	msg := message.(*messages.TransactionStartCompleted)
	switch msg.GetResult() {
	case messages.OperationResult_Success:
		o.succeed()
	case messages.OperationResult_PrepareTimeout, messages.OperationResult_CommitTimeout,
		messages.OperationResult_ForwardTimeout:
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
		err = fmt.Errorf("unexpected OperationResult: %s", msg.GetResult())
	}
	if res == nil && err == nil {
		res = client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil)
	}
	return
}

func (o *StartTransaction) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*messages.TransactionStartCompleted)
	return client.NewTransaction(msg.GetTransactionId(), o.userCredentials, o.parentConnection), nil
}

func (o *StartTransaction) createResponse() proto.Message {
	return &messages.TransactionStartCompleted{}
}

func (o *StartTransaction) String() string {
	return fmt.Sprintf("Start transaction on stream '%s' expecting version %d", o.stream, o.expectedVersion)
}

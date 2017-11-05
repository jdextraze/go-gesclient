package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type CommitTransaction struct {
	*baseOperation
	requireMaster bool
	transactionId int64
}

func NewCommitTransaction(
	source *tasks.CompletionSource,
	requireMaster bool,
	transactionId int64,
	userCredentials *client.UserCredentials,
) *CommitTransaction {
	obj := &CommitTransaction{
		requireMaster: requireMaster,
		transactionId: transactionId,
	}
	obj.baseOperation = newBaseOperation(client.Command_TransactionCommit, client.Command_TransactionCommitCompleted,
		userCredentials, source, obj.createRequestDto, obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *CommitTransaction) createRequestDto() proto.Message {
	return &messages.TransactionCommit{
		TransactionId: &o.transactionId,
		RequireMaster: &o.requireMaster,
	}
}

func (o *CommitTransaction) inspectResponse(message proto.Message) (res *client.InspectionResult, err error) {
	msg := message.(*messages.TransactionCommitCompleted)
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

func (o *CommitTransaction) transformResponse(message proto.Message) (interface{}, error) {
	msg := message.(*messages.TransactionCommitCompleted)
	preparePosition := int64(-1)
	if msg.PreparePosition != nil {
		preparePosition = *msg.PreparePosition
	}
	commitPosition := int64(-1)
	if msg.CommitPosition != nil {
		commitPosition = *msg.CommitPosition
	}
	return client.NewWriteResult(int(msg.GetLastEventNumber()), client.NewPosition(preparePosition, commitPosition)), nil
}

func (o *CommitTransaction) createResponse() proto.Message {
	return &messages.TransactionCommitCompleted{}
}

func (o *CommitTransaction) String() string {
	return fmt.Sprintf("Commit transaction #%d", o.transactionId)
}

package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type TransactionalWrite struct {
	*baseOperation
	requireMaster bool
	transactionId int64
	events        []*client.EventData
}

func NewTransactionalWrite(
	source *tasks.CompletionSource,
	requireMaster bool,
	transactionId int64,
	events []*client.EventData,
	userCredentials *client.UserCredentials,
) *TransactionalWrite {
	obj := &TransactionalWrite{
		requireMaster: requireMaster,
		transactionId: transactionId,
		events:        events,
	}
	obj.baseOperation = newBaseOperation(client.Command_TransactionWrite, client.Command_TransactionWriteCompleted,
		userCredentials, source, obj.createRequestDto, obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *TransactionalWrite) createRequestDto() proto.Message {
	events := make([]*messages.NewEvent, len(o.events))
	for i, e := range o.events {
		var (
			dataContentType     int32
			metadataContentType int32
		)
		eventType := e.Type()
		if e.IsJson() {
			dataContentType = 1
		}
		events[i] = &messages.NewEvent{
			EventId:             e.EventId().Bytes(),
			EventType:           &eventType,
			DataContentType:     &dataContentType,
			MetadataContentType: &metadataContentType,
			Data:                e.Data(),
			Metadata:            e.Metadata(),
		}
	}
	return &messages.TransactionWrite{
		TransactionId: &o.transactionId,
		RequireMaster: &o.requireMaster,
		Events:        events,
	}
}

func (o *TransactionalWrite) inspectResponse(message proto.Message) (res *client.InspectionResult, err error) {
	msg := message.(*messages.TransactionWriteCompleted)
	switch msg.GetResult() {
	case messages.OperationResult_Success:
		o.succeed()
	case messages.OperationResult_PrepareTimeout, messages.OperationResult_CommitTimeout,
		messages.OperationResult_ForwardTimeout:
		res = client.NewInspectionResult(client.InspectionDecision_Retry, msg.GetResult().String(), nil, nil)
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

func (o *TransactionalWrite) transformResponse(message proto.Message) (interface{}, error) {
	return nil, nil
}

func (o *TransactionalWrite) createResponse() proto.Message {
	return &messages.TransactionWriteCompleted{}
}

func (o *TransactionalWrite) String() string {
	return fmt.Sprintf("Transactional write on transaction #%d", o.transactionId)
}

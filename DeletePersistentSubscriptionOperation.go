package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type deletePersistentSubscriptionOperation struct {
	*baseOperation
	stream        string
	groupName     string
	resultChannel chan *PersistentSubscriptionDeleteResult
}

func newDeletePersistentSubscriptionOperation(
	stream string,
	groupName string,
	userCredentials *UserCredentials,
) *deletePersistentSubscriptionOperation {
	return &deletePersistentSubscriptionOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		groupName:     groupName,
		resultChannel: make(chan *PersistentSubscriptionDeleteResult, 1),
	}
}

func (o *deletePersistentSubscriptionOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_DeletePersistentSubscription
}

func (o *deletePersistentSubscriptionOperation) GetRequestMessage() proto.Message {
	return &protobuf.DeletePersistentSubscription{
		EventStreamId:         &o.stream,
		SubscriptionGroupName: &o.groupName,
	}
}

func (o *deletePersistentSubscriptionOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_DeletePersistentSubscriptionCompleted {
		err := o.handleError(p, tcpCommand_DeletePersistentSubscriptionCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.DeletePersistentSubscriptionCompleted{}
	if err := proto.Unmarshal(p.Payload, msg); err != nil {
		o.Fail(err)
		return
	}

	switch msg.GetResult() {
	case protobuf.DeletePersistentSubscriptionCompleted_Success:
		o.succeed(msg)
	case protobuf.DeletePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case protobuf.DeletePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(AccessDenied)
	case protobuf.DeletePersistentSubscriptionCompleted_DoesNotExist:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s doesn't exists", o.groupName, o.stream))
	default:
		o.Fail(fmt.Errorf("Unexpected operation result: %v", msg.GetResult()))
	}
}

func (o *deletePersistentSubscriptionOperation) succeed(msg *protobuf.DeletePersistentSubscriptionCompleted) {
	o.resultChannel <- newPersistentSubscriptionDeleteResult(PersistentSubscriptionCreateStatus_Success, nil)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *deletePersistentSubscriptionOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newPersistentSubscriptionDeleteResult(PersistentSubscriptionDeleteStatus_Failure, err)
	close(o.resultChannel)
	o.isCompleted = true
}

package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type deletePersistentSubscription struct {
	*baseOperation
	stream    string
	groupName string
}

func NewDeletePersistentSubscription(
	source *tasks.CompletionSource,
	stream string,
	groupName string,
	userCredentials *client.UserCredentials,
) *deletePersistentSubscription {
	obj := &deletePersistentSubscription{
		stream:    stream,
		groupName: groupName,
	}
	obj.baseOperation = newBaseOperation(client.Command_DeletePersistentSubscription,
		client.Command_DeletePersistentSubscriptionCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *deletePersistentSubscription) createRequestDto() proto.Message {
	return &messages.DeletePersistentSubscription{
		EventStreamId:         &o.stream,
		SubscriptionGroupName: &o.groupName,
	}
}

func (o *deletePersistentSubscription) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
	msg := message.(*messages.DeletePersistentSubscriptionCompleted)
	switch msg.GetResult() {
	case messages.DeletePersistentSubscriptionCompleted_Success:
		o.succeed()
	case messages.DeletePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case messages.DeletePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(client.AccessDenied)
	case messages.DeletePersistentSubscriptionCompleted_DoesNotExist:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s doesn't exists", o.groupName, o.stream))
	default:
		return nil, fmt.Errorf("Unexpected Operation result: %v", msg.GetResult())
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *deletePersistentSubscription) transformResponse(message proto.Message) (interface{}, error) {
	return client.NewPersistentSubscriptionDeleteResult(client.PersistentSubscriptionDeleteStatus_Success), nil
}

func (o *deletePersistentSubscription) createResponse() proto.Message {
	return &messages.DeletePersistentSubscriptionCompleted{}
}

func (o *deletePersistentSubscription) String() string {
	return fmt.Sprintf("DeletePersistentSubscription Stream: %s, Group Name: %s", o.stream, o.groupName)
}

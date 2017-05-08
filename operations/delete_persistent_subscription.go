package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/jdextraze/go-gesclient/models"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type deletePersistentSubscription struct {
	*baseOperation
	stream        string
	groupName     string
	resultChannel chan *models.PersistentSubscriptionDeleteResult
}

func NewDeletePersistentSubscription(
	stream string,
	groupName string,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.PersistentSubscriptionDeleteResult,
) *deletePersistentSubscription {
	return &deletePersistentSubscription{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		groupName:     groupName,
		resultChannel: resultChannel,
	}
}

func (o *deletePersistentSubscription) GetRequestCommand() models.Command {
	return models.Command_DeletePersistentSubscription
}

func (o *deletePersistentSubscription) GetRequestMessage() proto.Message {
	return &protobuf.DeletePersistentSubscription{
		EventStreamId:         &o.stream,
		SubscriptionGroupName: &o.groupName,
	}
}

func (o *deletePersistentSubscription) ParseResponse(p *models.Package) {
	if p.Command != models.Command_DeletePersistentSubscriptionCompleted {
		err := o.handleError(p, models.Command_DeletePersistentSubscriptionCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.DeletePersistentSubscriptionCompleted{}
	if err := proto.Unmarshal(p.Data, msg); err != nil {
		o.Fail(err)
		return
	}

	switch msg.GetResult() {
	case protobuf.DeletePersistentSubscriptionCompleted_Success:
		o.succeed(msg)
	case protobuf.DeletePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case protobuf.DeletePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	case protobuf.DeletePersistentSubscriptionCompleted_DoesNotExist:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s doesn't exists", o.groupName, o.stream))
	default:
		o.Fail(fmt.Errorf("Unexpected Operation result: %v", msg.GetResult()))
	}
}

func (o *deletePersistentSubscription) succeed(msg *protobuf.DeletePersistentSubscriptionCompleted) {
	o.resultChannel <- models.NewPersistentSubscriptionDeleteResult(models.PersistentSubscriptionCreateStatus_Success, nil)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *deletePersistentSubscription) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewPersistentSubscriptionDeleteResult(models.PersistentSubscriptionDeleteStatus_Failure, err)
	close(o.resultChannel)
	o.isCompleted = true
}

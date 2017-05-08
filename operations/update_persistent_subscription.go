package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"time"
	"github.com/jdextraze/go-gesclient/models"
)

type updatePersistentSubscription struct {
	*baseOperation
	stream        string
	groupName     string
	settings      models.PersistentSubscriptionSettings
	resultChannel chan *models.PersistentSubscriptionUpdateResult
}

func NewUpdatePersistentSubscription(
	stream string,
	groupName string,
	settings models.PersistentSubscriptionSettings,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.PersistentSubscriptionUpdateResult,
) *updatePersistentSubscription {
	return &updatePersistentSubscription{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		groupName:     groupName,
		settings:      settings,
		resultChannel: resultChannel,
	}
}

func (o *updatePersistentSubscription) GetRequestCommand() models.Command {
	return models.Command_UpdatePersistentSubscription
}

func (o *updatePersistentSubscription) GetRequestMessage() proto.Message {
	messageTimeoutMs := int32(o.settings.MessageTimeout().Nanoseconds() / int64(time.Millisecond))
	checkpointAfterMs := int32(o.settings.CheckPointAfter().Nanoseconds() / int64(time.Millisecond))
	isRoundRobin := o.settings.NamedConsumerStrategy.IsRoundRobin()
	namedConsumerStrategy := o.settings.NamedConsumerStrategy.ToString()
	resolvedLinkTos := o.settings.ResolveLinkTos()
	startFrom := o.settings.StartFrom()
	extraStatistics := o.settings.ExtraStatistics()
	maxCheckpointCount := o.settings.MaxCheckPointCount()
	minCheckpointCount := o.settings.MinCheckPointCount()
	maxSubscriberCount := o.settings.MaxSubscriberCount()
	return &protobuf.UpdatePersistentSubscription{
		EventStreamId:              &o.stream,
		SubscriptionGroupName:      &o.groupName,
		ResolveLinkTos:             &resolvedLinkTos,
		StartFrom:                  &startFrom,
		MessageTimeoutMilliseconds: &messageTimeoutMs,
		RecordStatistics:           &extraStatistics,
		LiveBufferSize:             &o.settings.LiveBufferSize,
		ReadBatchSize:              &o.settings.ReadBatchSize,
		BufferSize:                 &o.settings.LiveBufferSize,
		MaxRetryCount:              &o.settings.MaxRetryCount,
		PreferRoundRobin:           &isRoundRobin,
		CheckpointAfterTime:        &checkpointAfterMs,
		CheckpointMaxCount:         &maxCheckpointCount,
		CheckpointMinCount:         &minCheckpointCount,
		SubscriberMaxCount:         &maxSubscriberCount,
		NamedConsumerStrategy:      &namedConsumerStrategy,
	}
}

func (o *updatePersistentSubscription) ParseResponse(p *models.Package) {
	if p.Command != models.Command_UpdatePersistentSubscriptionCompleted {
		err := o.handleError(p, models.Command_UpdatePersistentSubscriptionCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.UpdatePersistentSubscriptionCompleted{}
	if err := proto.Unmarshal(p.Data, msg); err != nil {
		o.Fail(err)
		return
	}

	switch msg.GetResult() {
	case protobuf.UpdatePersistentSubscriptionCompleted_Success:
		o.succeed(msg)
	case protobuf.UpdatePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case protobuf.UpdatePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	case protobuf.UpdatePersistentSubscriptionCompleted_DoesNotExist:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s doesn't exists", o.groupName, o.stream))
	default:
		o.Fail(fmt.Errorf("Unexpected Operation result: %v", msg.GetResult()))
	}
}

func (o *updatePersistentSubscription) succeed(msg *protobuf.UpdatePersistentSubscriptionCompleted) {
	o.resultChannel <- models.NewPersistentSubscriptionUpdateResult(models.PersistentSubscriptionUpdateStatus_Success, nil)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *updatePersistentSubscription) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewPersistentSubscriptionUpdateResult(models.PersistentSubscriptionUpdateStatus_Failure, err)
	close(o.resultChannel)
	o.isCompleted = true
}

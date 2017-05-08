package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"time"
	"github.com/jdextraze/go-gesclient/models"
)

type createPersistentSubscription struct {
	*baseOperation
	stream        string
	groupName     string
	settings      models.PersistentSubscriptionSettings
	resultChannel chan *models.PersistentSubscriptionCreateResult
}

func NewCreatePersistentSubscription(
	stream string,
	groupName string,
	settings models.PersistentSubscriptionSettings,
	userCredentials *models.UserCredentials,
	resultChannel chan *models.PersistentSubscriptionCreateResult,
) *createPersistentSubscription {
	return &createPersistentSubscription{
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

func (o *createPersistentSubscription) GetRequestCommand() models.Command {
	return models.Command_CreatePersistentSubscription
}

func (o *createPersistentSubscription) GetRequestMessage() proto.Message {
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
	return &protobuf.CreatePersistentSubscription{
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

func (o *createPersistentSubscription) ParseResponse(p *models.Package) {
	if p.Command != models.Command_CreatePersistentSubscriptionCompleted {
		err := o.handleError(p, models.Command_CreatePersistentSubscriptionCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.CreatePersistentSubscriptionCompleted{}
	if err := proto.Unmarshal(p.Data, msg); err != nil {
		o.Fail(err)
		return
	}

	switch msg.GetResult() {
	case protobuf.CreatePersistentSubscriptionCompleted_Success:
		o.succeed(msg)
	case protobuf.CreatePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case protobuf.CreatePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	case protobuf.CreatePersistentSubscriptionCompleted_AlreadyExists:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s already exists", o.groupName, o.stream))
	default:
		o.Fail(fmt.Errorf("Unexpected Operation result: %v", msg.GetResult()))
	}
}

func (o *createPersistentSubscription) succeed(msg *protobuf.CreatePersistentSubscriptionCompleted) {
	o.resultChannel <- models.NewPersistentSubscriptionCreateResult(models.PersistentSubscriptionCreateStatus_Success, nil)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *createPersistentSubscription) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- models.NewPersistentSubscriptionCreateResult(models.PersistentSubscriptionCreateStatus_Failure, err)
	close(o.resultChannel)
	o.isCompleted = true
}

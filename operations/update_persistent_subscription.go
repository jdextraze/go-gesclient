package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"time"
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/common"
	"github.com/jdextraze/go-gesclient/tasks"
)

type updatePersistentSubscription struct {
	*baseOperation
	stream                     string
	groupName                  string
	resolveLinkTos             bool
	startFromBeginning         int32
	messageTimeoutMilliseconds int32
	recordStatistics           bool
	maxRetryCount              int32
	liveBufferSize             int32
	readBatchSize              int32
	bufferSize                 int32
	checkPointAfter            int32
	minCheckPointCount         int32
	maxCheckPointCount         int32
	maxSubscriberCount         int32
	namedConsumerStrategy      string
}

func NewUpdatePersistentSubscription(
	source *tasks.CompletionSource,
	stream string,
	groupName string,
	settings *models.PersistentSubscriptionSettings,
	userCredentials *models.UserCredentials,
) *updatePersistentSubscription {
	if settings == nil {
		panic("settings is nil")
	}
	obj := &updatePersistentSubscription{
		stream:                     stream,
		groupName:                  groupName,
		resolveLinkTos:             settings.ResolveLinkTos(),
		startFromBeginning:         settings.StartFrom(),
		maxRetryCount:              settings.MaxRetryCount,
		liveBufferSize:             settings.LiveBufferSize,
		readBatchSize:              settings.ReadBatchSize,
		bufferSize:                 settings.HistoryBufferSize,
		recordStatistics:           settings.ExtraStatistics(),
		messageTimeoutMilliseconds: int32(settings.MessageTimeout().Nanoseconds() / int64(time.Millisecond)),
		checkPointAfter:            int32(settings.CheckPointAfter().Nanoseconds() / int64(time.Millisecond)),
		minCheckPointCount:         settings.MinCheckPointCount(),
		maxCheckPointCount:         settings.MaxCheckPointCount(),
		maxSubscriberCount:         settings.MaxSubscriberCount(),
		namedConsumerStrategy:      settings.NamedConsumerStrategy.ToString(),
	}
	obj.baseOperation = newBaseOperation(models.Command_UpdatePersistentSubscription,
		models.Command_UpdatePersistentSubscriptionCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *updatePersistentSubscription) createRequestDto() proto.Message {
	preferRoundRobin := o.namedConsumerStrategy == common.SystemConsumerStrategies_RoundRobin.ToString()
	return &protobuf.UpdatePersistentSubscription{
		EventStreamId:              &o.stream,
		SubscriptionGroupName:      &o.groupName,
		ResolveLinkTos:             &o.resolveLinkTos,
		StartFrom:                  &o.startFromBeginning,
		MessageTimeoutMilliseconds: &o.messageTimeoutMilliseconds,
		RecordStatistics:           &o.recordStatistics,
		LiveBufferSize:             &o.liveBufferSize,
		ReadBatchSize:              &o.readBatchSize,
		BufferSize:                 &o.bufferSize,
		MaxRetryCount:              &o.maxRetryCount,
		PreferRoundRobin:           &preferRoundRobin,
		CheckpointAfterTime:        &o.checkPointAfter,
		CheckpointMaxCount:         &o.maxCheckPointCount,
		CheckpointMinCount:         &o.minCheckPointCount,
		SubscriberMaxCount:         &o.maxSubscriberCount,
		NamedConsumerStrategy:      &o.namedConsumerStrategy,
	}
}

func (o *updatePersistentSubscription) inspectResponse(message proto.Message) (*models.InspectionResult, error) {
	msg := message.(*protobuf.UpdatePersistentSubscriptionCompleted)
	switch msg.GetResult() {
	case protobuf.UpdatePersistentSubscriptionCompleted_Success:
		if err := o.succeed(); err != nil {
			return nil, err
		}
	case protobuf.UpdatePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case protobuf.UpdatePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(models.AccessDenied)
	case protobuf.UpdatePersistentSubscriptionCompleted_DoesNotExist:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s does not exists", o.groupName, o.stream))
	default:
		return nil, fmt.Errorf("Unexpected Operation result: %v", msg.GetResult())
	}
	return models.NewInspectionResult(models.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *updatePersistentSubscription) transformResponse(message proto.Message) (interface{}, error) {
	return models.NewPersistentSubscriptionUpdateResult(models.PersistentSubscriptionUpdateStatus_Success), nil
}

func (o *updatePersistentSubscription) createResponse() proto.Message {
	return &protobuf.UpdatePersistentSubscriptionCompleted{}
}

func (o *updatePersistentSubscription) String() string {
	return fmt.Sprintf("UpdatePersistentSubscription Stream: %s, Group Name: %s", o.stream, o.groupName)
}

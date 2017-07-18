package operations

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/common"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type createPersistentSubscription struct {
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

func NewCreatePersistentSubscription(
	source *tasks.CompletionSource,
	stream string,
	groupName string,
	settings *client.PersistentSubscriptionSettings,
	userCredentials *client.UserCredentials,
) *createPersistentSubscription {
	if settings == nil {
		panic("settings is nil")
	}
	obj := &createPersistentSubscription{
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
	obj.baseOperation = newBaseOperation(client.Command_CreatePersistentSubscription,
		client.Command_CreatePersistentSubscriptionCompleted, userCredentials, source, obj.createRequestDto,
		obj.inspectResponse, obj.transformResponse, obj.createResponse)
	return obj
}

func (o *createPersistentSubscription) createRequestDto() proto.Message {
	preferRoundRobin := o.namedConsumerStrategy == common.SystemConsumerStrategies_RoundRobin.ToString()
	return &messages.CreatePersistentSubscription{
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

func (o *createPersistentSubscription) inspectResponse(message proto.Message) (*client.InspectionResult, error) {
	msg := message.(*messages.CreatePersistentSubscriptionCompleted)
	switch msg.GetResult() {
	case messages.CreatePersistentSubscriptionCompleted_Success:
		if err := o.succeed(); err != nil {
			return nil, err
		}
	case messages.CreatePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case messages.CreatePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(client.AccessDenied)
	case messages.CreatePersistentSubscriptionCompleted_AlreadyExists:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s already exists", o.groupName, o.stream))
	default:
		return nil, fmt.Errorf("Unexpected Operation result: %v", msg.GetResult())
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, msg.GetResult().String(), nil, nil), nil
}

func (o *createPersistentSubscription) transformResponse(message proto.Message) (interface{}, error) {
	return client.NewPersistentSubscriptionCreateResult(client.PersistentSubscriptionCreateStatus_Success), nil
}

func (o *createPersistentSubscription) createResponse() proto.Message {
	return &messages.CreatePersistentSubscriptionCompleted{}
}

func (o *createPersistentSubscription) String() string {
	return fmt.Sprintf("CreatePersistentSubscription Stream: %s, Group Name: %s", o.stream, o.groupName)
}

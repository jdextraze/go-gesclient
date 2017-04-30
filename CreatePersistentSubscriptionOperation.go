package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"time"
)

type createPersistentSubscriptionOperation struct {
	*baseOperation
	stream        string
	groupName     string
	settings      PersistentSubscriptionSettings
	resultChannel chan *PersistentSubscriptionCreateResult
}

func newCreatePersistentSubscriptionOperation(
	stream string,
	groupName string,
	settings PersistentSubscriptionSettings,
	userCredentials *UserCredentials,
) *createPersistentSubscriptionOperation {
	return &createPersistentSubscriptionOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		groupName:     groupName,
		settings:      settings,
		resultChannel: make(chan *PersistentSubscriptionCreateResult, 1),
	}
}

func (o *createPersistentSubscriptionOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_CreatePersistentSubscription
}

func (o *createPersistentSubscriptionOperation) GetRequestMessage() proto.Message {
	messageTimeoutMs := int32(o.settings.messageTimeout.Nanoseconds() / int64(time.Millisecond))
	checkpointAfterMs := int32(o.settings.checkPointAfter.Nanoseconds() / int64(time.Millisecond))
	isRoundRobin := o.settings.NamedConsumerStrategy.IsRoundRobin()
	namedConsumerStrategy := o.settings.NamedConsumerStrategy.ToString()
	return &protobuf.CreatePersistentSubscription{
		EventStreamId:              &o.stream,
		SubscriptionGroupName:      &o.groupName,
		ResolveLinkTos:             &o.settings.resolveLinkTos,
		StartFrom:                  &o.settings.startFrom,
		MessageTimeoutMilliseconds: &messageTimeoutMs,
		RecordStatistics:           &o.settings.extraStatistics,
		LiveBufferSize:             &o.settings.LiveBufferSize,
		ReadBatchSize:              &o.settings.ReadBatchSize,
		BufferSize:                 &o.settings.LiveBufferSize,
		MaxRetryCount:              &o.settings.MaxRetryCount,
		PreferRoundRobin:           &isRoundRobin,
		CheckpointAfterTime:        &checkpointAfterMs,
		CheckpointMaxCount:         &o.settings.maxCheckPointCount,
		CheckpointMinCount:         &o.settings.minCheckPointCount,
		SubscriberMaxCount:         &o.settings.maxSubscriberCount,
		NamedConsumerStrategy:      &namedConsumerStrategy,
	}
}

func (o *createPersistentSubscriptionOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_CreatePersistentSubscriptionCompleted {
		err := o.handleError(p, tcpCommand_CreatePersistentSubscriptionCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.CreatePersistentSubscriptionCompleted{}
	if err := proto.Unmarshal(p.Payload, msg); err != nil {
		o.Fail(err)
		return
	}

	switch msg.GetResult() {
	case protobuf.CreatePersistentSubscriptionCompleted_Success:
		o.succeed(msg)
	case protobuf.CreatePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case protobuf.CreatePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(AccessDenied)
	case protobuf.CreatePersistentSubscriptionCompleted_AlreadyExists:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s already exists", o.groupName, o.stream))
	default:
		o.Fail(fmt.Errorf("Unexpected operation result: %v", msg.GetResult()))
	}
}

func (o *createPersistentSubscriptionOperation) succeed(msg *protobuf.CreatePersistentSubscriptionCompleted) {
	o.resultChannel <- newPersistentSubscriptionCreateResult(PersistentSubscriptionCreateStatus_Success, nil)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *createPersistentSubscriptionOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newPersistentSubscriptionCreateResult(PersistentSubscriptionCreateStatus_Failure, err)
	close(o.resultChannel)
	o.isCompleted = true
}

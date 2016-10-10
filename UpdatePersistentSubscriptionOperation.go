package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"time"
)

type updatePersistentSubscriptionOperation struct {
	*baseOperation
	stream        string
	groupName     string
	settings      PersistentSubscriptionSettings
	resultChannel chan *PersistentSubscriptionUpdateResult
}

func newUpdatePersistentSubscriptionOperation(
	stream string,
	groupName string,
	settings PersistentSubscriptionSettings,
	userCredentials *UserCredentials,
) *updatePersistentSubscriptionOperation {
	return &updatePersistentSubscriptionOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		groupName:     groupName,
		settings:      settings,
		resultChannel: make(chan *PersistentSubscriptionUpdateResult, 1),
	}
}

func (o *updatePersistentSubscriptionOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_UpdatePersistentSubscription
}

func (o *updatePersistentSubscriptionOperation) GetRequestMessage() proto.Message {
	messageTimeoutMs := int32(o.settings.messageTimeout.Nanoseconds() / int64(time.Millisecond))
	checkpointAfterMs := int32(o.settings.checkPointAfter.Nanoseconds() / int64(time.Millisecond))
	isRoundRobin := o.settings.NamedConsumerStrategy.IsRoundRobin()
	namedConsumerStrategy := o.settings.NamedConsumerStrategy.ToString()
	return &protobuf.UpdatePersistentSubscription{
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

func (o *updatePersistentSubscriptionOperation) ParseResponse(p *tcpPacket) {
	if p.Command != tcpCommand_UpdatePersistentSubscriptionCompleted {
		err := o.handleError(p, tcpCommand_UpdatePersistentSubscriptionCompleted)
		if err != nil {
			o.Fail(err)
		}
		return
	}

	msg := &protobuf.UpdatePersistentSubscriptionCompleted{}
	if err := proto.Unmarshal(p.Payload, msg); err != nil {
		o.Fail(err)
		return
	}

	switch msg.GetResult() {
	case protobuf.UpdatePersistentSubscriptionCompleted_Success:
		o.succeed(msg)
	case protobuf.UpdatePersistentSubscriptionCompleted_Fail:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s failed '%s'", o.groupName, o.stream, *msg.Reason))
	case protobuf.UpdatePersistentSubscriptionCompleted_AccessDenied:
		o.Fail(AccessDenied)
	case protobuf.UpdatePersistentSubscriptionCompleted_DoesNotExist:
		o.Fail(fmt.Errorf("Subscription group %s on stream %s doesn't exists", o.groupName, o.stream))
	default:
		o.Fail(fmt.Errorf("Unexpected operation result: %v", msg.GetResult()))
	}
}

func (o *updatePersistentSubscriptionOperation) succeed(msg *protobuf.UpdatePersistentSubscriptionCompleted) {
	o.resultChannel <- newPersistentSubscriptionUpdateResult(PersistentSubscriptionUpdateStatus_Success, nil)
	close(o.resultChannel)
	o.isCompleted = true
}

func (o *updatePersistentSubscriptionOperation) Fail(err error) {
	if o.isCompleted {
		return
	}
	o.resultChannel <- newPersistentSubscriptionUpdateResult(PersistentSubscriptionUpdateStatus_Failure, err)
	close(o.resultChannel)
	o.isCompleted = true
}

package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/jdextraze/go-gesclient/models"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"time"
)

type connectToPersistentSubscription struct {
	*baseOperation
	stream         string
	groupName      string
	resultChannel  chan models.PersistentSubscription
	confirmation   *models.SubscriptionConfirmation
	conn           models.InternalConnection
	events         chan *models.ResolvedEvent
	dropped        chan *models.SubscriptionDropped
	confirmed      bool
	error          error
	unsubscribe    bool
	autoAck        bool
	ackEvents      []*models.ResolvedEvent
	nakEvents      []*models.ResolvedEvent
	nakAction      models.PersistentSubscriptionNakEventAction
	nakMessage     string
	subscriptionId string
}

func NewConnectToPersistentSubscription(
	stream string,
	groupName string,
	autoAck bool,
	conn models.InternalConnection,
	userCredentials *models.UserCredentials,
	resultChannel chan models.PersistentSubscription,
) *connectToPersistentSubscription {
	return &connectToPersistentSubscription{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		groupName:     groupName,
		autoAck:       autoAck,
		resultChannel: resultChannel,
		conn:          conn,
		events:        make(chan *models.ResolvedEvent, 4096),
		dropped:       make(chan *models.SubscriptionDropped, 1),
	}
}

func (o *connectToPersistentSubscription) Start() (chan models.PersistentSubscription, error) {
	// TODO
	return o.resultChannel, nil
}

func (o *connectToPersistentSubscription) GetRequestCommand() models.Command {
	if len(o.ackEvents) > 0 {
		o.ackEvents = nil
		return models.Command_PersistentSubscriptionAckEvents
	}
	if len(o.nakEvents) > 0 {
		o.nakEvents = nil
		return models.Command_PersistentSubscriptionNakEvents
	}
	if o.unsubscribe {
		return models.Command_UnsubscribeFromStream
	}
	return models.Command_ConnectToPersistentSubscription
}

func (o *connectToPersistentSubscription) GetRequestMessage() proto.Message {
	if len(o.ackEvents) > 0 {
		eventIds := make([][]byte, len(o.ackEvents))
		for i, evt := range o.ackEvents {
			eventIds[i] = evt.OriginalEvent().EventId().Bytes()
		}
		return &protobuf.PersistentSubscriptionAckEvents{
			SubscriptionId:    &o.subscriptionId,
			ProcessedEventIds: eventIds,
		}
	}
	if len(o.nakEvents) > 0 {
		eventIds := make([][]byte, len(o.nakEvents))
		for i, evt := range o.nakEvents {
			eventIds[i] = evt.OriginalEvent().EventId().Bytes()
		}
		action := protobuf.PersistentSubscriptionNakEvents_NakAction(o.nakAction)
		return &protobuf.PersistentSubscriptionNakEvents{
			SubscriptionId:    &o.subscriptionId,
			ProcessedEventIds: eventIds,
			Message:           &o.nakMessage,
			Action:            &action,
		}
	}
	if o.unsubscribe {
		return &protobuf.UnsubscribeFromStream{}
	}
	bufferSize := int32(10)
	return &protobuf.ConnectToPersistentSubscription{
		EventStreamId:           &o.stream,
		SubscriptionId:          &o.groupName,
		AllowedInFlightMessages: &bufferSize,
	}
}

func (o *connectToPersistentSubscription) Confirmation() *models.SubscriptionConfirmation {
	return o.confirmation
}

func (o *connectToPersistentSubscription) Events() <-chan *models.ResolvedEvent {
	return o.events
}

func (o *connectToPersistentSubscription) Dropped() <-chan *models.SubscriptionDropped {
	return o.dropped
}

func (o *connectToPersistentSubscription) ParseResponse(p *models.Package) {
	switch p.Command {
	case models.Command_PersistentSubscriptionConfirmation:
		o.subscriptionConfirmation(p.Data)
	case models.Command_PersistentSubscriptionStreamEventAppeared:
		o.streamEventAppeared(p.Data)
	case models.Command_SubscriptionDropped:
		o.subscriptionDropped(p.Data)
	default:
		o.handleError(p, models.Command_PersistentSubscriptionStreamEventAppeared)
	}
}

func (o *connectToPersistentSubscription) Error() error { return o.error }

func (o *connectToPersistentSubscription) Fail(err error) {
	if !o.confirmed {
		o.error = err
		o.resultChannel <- o
		close(o.resultChannel)
	}
	evt := &models.SubscriptionDropped{
		Reason: models.SubscriptionDropReason_Error,
		Error:  err,
	}
	o.dropped <- evt
	close(o.dropped)
	close(o.events)
	o.isCompleted = true
}

func (o *connectToPersistentSubscription) Acknowledge(events ...*models.ResolvedEvent) {
	o.ackEvents = events
	o.conn.EnqueueOperation(o)
}

func (o *connectToPersistentSubscription) Failure(
	action models.PersistentSubscriptionNakEventAction,
	reason string,
	events ...*models.ResolvedEvent,
) {
	o.nakEvents = events
	o.nakAction = action
	o.nakMessage = reason
	o.conn.EnqueueOperation(o)
}

func (o *connectToPersistentSubscription) Stop(timeout time.Duration) error {
	if !o.confirmed {
		return errors.New("Subscription not confirmed")
	}
	o.unsubscribe = true
	return o.conn.EnqueueOperation(o)
}

func (o *connectToPersistentSubscription) subscriptionConfirmation(payload []byte) {
	if o.confirmed {
		return
	}

	msg := &protobuf.PersistentSubscriptionConfirmation{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	o.subscriptionId = msg.GetSubscriptionId()
	o.confirmation = models.NewSubscriptionConfirmation(msg.GetLastCommitPosition(), msg.GetLastEventNumber())
	o.resultChannel <- o
	close(o.resultChannel)
	o.confirmed = true
}

func (o *connectToPersistentSubscription) streamEventAppeared(payload []byte) {
	msg := &protobuf.PersistentSubscriptionStreamEventAppeared{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	evt := models.NewResolvedEvent(msg.Event)
	o.events <- evt
	if o.autoAck {
		o.Acknowledge(evt)
	}
}

func (o *connectToPersistentSubscription) subscriptionDropped(payload []byte) {
	msg := &protobuf.SubscriptionDropped{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	d := &models.SubscriptionDropped{
		Reason: models.SubscriptionDropReason(msg.GetReason()),
	}
	o.dropped <- d
	close(o.dropped)
	close(o.events)
	o.isCompleted = true
}

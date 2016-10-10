package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"time"
)

type PersistentSubscriptionNakEventAction int

const (
	PersistentSubscriptionNakEventAction_Unknown = 0
	PersistentSubscriptionNakEventAction_Park    = 1
	PersistentSubscriptionNakEventAction_Retry   = 2
	PersistentSubscriptionNakEventAction_Skip    = 3
	PersistentSubscriptionNakEventAction_Stop    = 4
)

type PersistentSubscription interface {
	Confirmation() *SubscriptionConfirmation
	Events() <-chan *ResolvedEvent
	Dropped() <-chan *SubscriptionDropped
	Acknowledge(events ...*ResolvedEvent)
	Failure(action PersistentSubscriptionNakEventAction, reason string, events ...*ResolvedEvent)
	Stop(timeout time.Duration) error
	Error() error
}

type connectToPersistentSubscriptionOperation struct {
	*baseOperation
	stream        string
	groupName     string
	resultChannel chan PersistentSubscription
	confirmation  *SubscriptionConfirmation
	conn          *connection
	events        chan *ResolvedEvent
	dropped       chan *SubscriptionDropped
	confirmed     bool
	error         error
	unsubscribe   bool
	autoAck       bool
	ackEvents     []*ResolvedEvent
	nakEvents     []*ResolvedEvent
	nakAction     PersistentSubscriptionNakEventAction
	nakMessage    string
}

func newConnectToPersistentSubscriptionOperation(
	stream string,
	groupName string,
	autoAck bool,
	conn *connection,
	userCredentials *UserCredentials,
) *connectToPersistentSubscriptionOperation {
	return &connectToPersistentSubscriptionOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		groupName:     groupName,
		autoAck:       autoAck,
		resultChannel: make(chan PersistentSubscription, 1),
		conn:          conn,
		events:        make(chan *ResolvedEvent, 4096),
		dropped:       make(chan *SubscriptionDropped, 1),
	}
}

func (o *connectToPersistentSubscriptionOperation) GetRequestCommand() tcpCommand {
	if len(o.ackEvents) > 0 {
		o.ackEvents = nil
		return tcpCommand_PersistentSubscriptionAckEvents
	}
	if len(o.nakEvents) > 0 {
		o.nakEvents = nil
		return tcpCommand_PersistentSubscriptionNakEvents
	}
	if o.unsubscribe {
		return tcpCommand_UnsubscribeFromStream
	}
	return tcpCommand_ConnectToPersistentSubscription
}

func (o *connectToPersistentSubscriptionOperation) GetRequestMessage() proto.Message {
	if len(o.ackEvents) > 0 {
		eventIds := make([][]byte, len(o.ackEvents))
		for i, evt := range o.ackEvents {
			eventIds[i] = evt.Event().eventId.Bytes()
		}
		return &protobuf.PersistentSubscriptionAckEvents{
			SubscriptionId:    &o.groupName,
			ProcessedEventIds: eventIds,
		}
	}
	if len(o.nakEvents) > 0 {
		eventIds := make([][]byte, len(o.nakEvents))
		for i, evt := range o.nakEvents {
			eventIds[i] = evt.Event().eventId.Bytes()
		}
		action := protobuf.PersistentSubscriptionNakEvents_NakAction(o.nakAction)
		return &protobuf.PersistentSubscriptionNakEvents{
			SubscriptionId:    &o.groupName,
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

func (o *connectToPersistentSubscriptionOperation) Confirmation() *SubscriptionConfirmation {
	return o.confirmation
}

func (o *connectToPersistentSubscriptionOperation) Events() <-chan *ResolvedEvent {
	return o.events
}

func (o *connectToPersistentSubscriptionOperation) Dropped() <-chan *SubscriptionDropped {
	return o.dropped
}

func (o *connectToPersistentSubscriptionOperation) ParseResponse(p *tcpPacket) {
	switch p.Command {
	case tcpCommand_PersistentSubscriptionConfirmation:
		o.subscriptionConfirmation(p.Payload)
	case tcpCommand_PersistentSubscriptionStreamEventAppeared:
		o.streamEventAppeared(p.Payload)
	case tcpCommand_SubscriptionDropped:
		o.subscriptionDropped(p.Payload)
	default:
		o.handleError(p, tcpCommand_PersistentSubscriptionStreamEventAppeared)
	}
}

func (o *connectToPersistentSubscriptionOperation) Error() error { return o.error }

func (o *connectToPersistentSubscriptionOperation) Fail(err error) {
	if !o.confirmed {
		o.error = err
		o.resultChannel <- o
		close(o.resultChannel)
	}
	evt := &SubscriptionDropped{
		Reason: SubscriptionDropReason_Error,
		Error:  err,
	}
	o.dropped <- evt
	close(o.dropped)
	close(o.events)
	o.isCompleted = true
}

func (o *connectToPersistentSubscriptionOperation) Acknowledge(events ...*ResolvedEvent) {
	o.ackEvents = events
	o.conn.enqueueOperation(o, false)
}

func (o *connectToPersistentSubscriptionOperation) Failure(
	action PersistentSubscriptionNakEventAction,
	reason string,
	events ...*ResolvedEvent,
) {
	o.nakEvents = events
	o.nakAction = action
	o.nakMessage = reason
	o.conn.enqueueOperation(o, false)
}

func (o *connectToPersistentSubscriptionOperation) Stop(timeout time.Duration) error {
	if !o.confirmed {
		return errors.New("Subscription not confirmed")
	}
	o.unsubscribe = true
	return o.conn.enqueueOperation(o, false)
}

func (o *connectToPersistentSubscriptionOperation) subscriptionConfirmation(payload []byte) {
	if o.confirmed {
		return
	}

	msg := &protobuf.PersistentSubscriptionConfirmation{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	o.confirmation = &SubscriptionConfirmation{msg.GetLastCommitPosition(), msg.GetLastEventNumber()}
	o.resultChannel <- o
	close(o.resultChannel)
	o.confirmed = true
}

func (o *connectToPersistentSubscriptionOperation) streamEventAppeared(payload []byte) {
	msg := &protobuf.PersistentSubscriptionStreamEventAppeared{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	evt := newResolvedEvent(msg.Event)
	o.events <- evt
	if o.autoAck {
		o.Acknowledge(evt)
	}
}

func (o *connectToPersistentSubscriptionOperation) subscriptionDropped(payload []byte) {
	msg := &protobuf.SubscriptionDropped{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	d := &SubscriptionDropped{
		Reason: SubscriptionDropReason(msg.GetReason()),
	}
	o.dropped <- d
	close(o.dropped)
	close(o.events)
	o.isCompleted = true
}

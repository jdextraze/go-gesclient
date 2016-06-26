package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type SubscriptionConfirmation struct {
	lastCommitPosition int64
	lastEventNumber    int32
}

func (c *SubscriptionConfirmation) LastCommitPosition() int64 {
	return c.lastCommitPosition
}

func (c *SubscriptionConfirmation) LastEventNumber() int32 {
	return c.lastEventNumber
}

type SubscriptionDropReason int

// TODO review this
const (
	SubscriptionDropReason_Error                         SubscriptionDropReason = -1
	SubscriptionDropReason_Unsubscribed                  SubscriptionDropReason = 0
	SubscriptionDropReason_AccessDenied                  SubscriptionDropReason = 1
	SubscriptionDropReason_NotFound                      SubscriptionDropReason = 2
	SubscriptionDropReason_PersistentSubscriptionDeleted SubscriptionDropReason = 3
	SubscriptionDropReason_SubscriberMaxCountReached     SubscriptionDropReason = 4

	SubscriptionDropReason_UserInitiated           SubscriptionDropReason = 5
	SubscriptionDropReason_CatchUpError            SubscriptionDropReason = 6
	SubscriptionDropReason_ProcessingQueueOverflow SubscriptionDropReason = 7
)

type SubscriptionDropped struct {
	Reason SubscriptionDropReason
	Error  error
}

type Subscription interface {
	Confirmation() *SubscriptionConfirmation
	Events() <-chan *ResolvedEvent
	Unsubscribe() error
	Dropped() <-chan *SubscriptionDropped
	Error() error
}

type subscribeToStreamOperation struct {
	*baseOperation
	stream        string
	resultChannel chan Subscription
	confirmation  *SubscriptionConfirmation
	conn          *connection
	events        chan *ResolvedEvent
	dropped       chan *SubscriptionDropped
	confirmed     bool
	error         error
	unsubscribe   bool
}

func newSubscribeToStreamOperation(
	stream string,
	conn *connection,
	userCredentials *UserCredentials,
) *subscribeToStreamOperation {
	return &subscribeToStreamOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		resultChannel: make(chan Subscription, 1),
		conn:          conn,
		events:        make(chan *ResolvedEvent, 1000),
		dropped:       make(chan *SubscriptionDropped, 1),
	}
}

func (o *subscribeToStreamOperation) GetRequestCommand() tcpCommand {
	if o.unsubscribe {
		return tcpCommand_UnsubscribeFromStream
	}
	return tcpCommand_SubscribeToStream
}

func (o *subscribeToStreamOperation) GetRequestMessage() proto.Message {
	if o.unsubscribe {
		return &protobuf.UnsubscribeFromStream{}
	}
	no := false
	return &protobuf.SubscribeToStream{
		EventStreamId:  &o.stream,
		ResolveLinkTos: &no,
	}
}

func (o *subscribeToStreamOperation) ParseResponse(p *tcpPacket) {
	switch p.Command {
	case tcpCommand_SubscriptionConfirmation:
		o.subscriptionConfirmation(p.Payload)
	case tcpCommand_StreamEventAppeared:
		o.streamEventAppeared(p.Payload)
	case tcpCommand_SubscriptionDropped:
		o.subscriptionDropped(p.Payload)
	default:
		o.handleError(p, tcpCommand_StreamEventAppeared)
	}
}

func (o *subscribeToStreamOperation) Confirmation() *SubscriptionConfirmation {
	return o.confirmation
}

func (o *subscribeToStreamOperation) Events() <-chan *ResolvedEvent {
	return o.events
}

func (o *subscribeToStreamOperation) Unsubscribe() error {
	if !o.confirmed {
		return errors.New("Subscription not confirmed")
	}
	o.unsubscribe = true
	return o.conn.enqueueOperation(o, false)
}

func (o *subscribeToStreamOperation) Dropped() <-chan *SubscriptionDropped {
	return o.dropped
}

func (o *subscribeToStreamOperation) Error() error { return o.error }

func (o *subscribeToStreamOperation) Fail(err error) {
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

func (o *subscribeToStreamOperation) subscriptionConfirmation(payload []byte) {
	if o.confirmed {
		return
	}

	msg := &protobuf.SubscriptionConfirmation{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	o.confirmation = &SubscriptionConfirmation{msg.GetLastCommitPosition(), msg.GetLastEventNumber()}
	o.resultChannel <- o
	close(o.resultChannel)
	o.confirmed = true
}

func (o *subscribeToStreamOperation) streamEventAppeared(payload []byte) {
	msg := &protobuf.StreamEventAppeared{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	o.events <- newResolvedEventFrom(msg.Event)
}

func (o *subscribeToStreamOperation) subscriptionDropped(payload []byte) {
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

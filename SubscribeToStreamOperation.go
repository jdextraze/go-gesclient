package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

type SubscriptionConfirmation struct{}

type SubscriptionDropReason int

const (
	SubscriptionDropReasonError                         SubscriptionDropReason = -1
	SubscriptionDropReasonUnsubscribed                  SubscriptionDropReason = 0
	SubscriptionDropReasonAccessDenied                  SubscriptionDropReason = 1
	SubscriptionDropReasonNotFound                      SubscriptionDropReason = 2
	SubscriptionDropReasonPersistentSubscriptionDeleted SubscriptionDropReason = 3
	SubscriptionDropReasonSubscriberMaxCountReached     SubscriptionDropReason = 4
)

type SubscriptionDropped struct {
	Reason SubscriptionDropReason
	Error  error
}

type Subscription interface {
	Confirmation() *SubscriptionConfirmation
	Events() chan *ResolvedEvent
	Unsubscribe() error
	Dropped() chan *SubscriptionDropped
	Error() error
}

type subscribeToStreamOperation struct {
	*baseOperation
	stream       string
	c            chan Subscription
	confirmation *SubscriptionConfirmation
	conn         *connection
	events       []chan *ResolvedEvent
	dropped      []chan *SubscriptionDropped
	confirmed    bool
	error        error
}

func newSubscribeToStreamOperation(
	stream string,
	c chan Subscription,
	conn *connection,
	userCredentials *UserCredentials,
) *subscribeToStreamOperation {
	return &subscribeToStreamOperation{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:  stream,
		c:       c,
		conn:    conn,
		events:  make([]chan *ResolvedEvent, 0),
		dropped: make([]chan *SubscriptionDropped, 0),
	}
}

func (o *subscribeToStreamOperation) GetRequestCommand() tcpCommand {
	return tcpCommand_SubscribeToStream
}

func (o *subscribeToStreamOperation) GetRequestMessage() proto.Message {
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
	}
}

func (o *subscribeToStreamOperation) Confirmation() *SubscriptionConfirmation {
	return o.confirmation
}

func (o *subscribeToStreamOperation) Events() chan *ResolvedEvent {
	events := make(chan *ResolvedEvent)
	o.events = append(o.events, events)
	return events
}

func (o *subscribeToStreamOperation) Unsubscribe() error {
	if err := o.conn.assertConnected(); err != nil {
		return err
	}

	payload, err := proto.Marshal(&protobuf.UnsubscribeFromStream{})
	if err != nil {
		return err
	}

	userCredentials := o.UserCredentials()
	var authFlag byte = 0
	if userCredentials != nil {
		authFlag = 1
	}
	o.conn.output <- newTcpPacket(
		tcpCommand_UnsubscribeFromStream,
		authFlag,
		o.correlationId,
		payload,
		userCredentials,
	)

	return nil
}

func (o *subscribeToStreamOperation) Dropped() chan *SubscriptionDropped {
	dropped := make(chan *SubscriptionDropped)
	o.dropped = append(o.dropped, dropped)
	return dropped
}

func (o *subscribeToStreamOperation) Error() error { return o.error }

func (o *subscribeToStreamOperation) Fail(err error) {
	if !o.confirmed {
		o.error = err
		o.c <- o
		close(o.c)
	}
	evt := &SubscriptionDropped{
		Reason: SubscriptionDropReasonError,
		Error:  err,
	}
	for _, ch := range o.dropped {
		ch <- evt
		close(ch)
	}
	for _, ch := range o.events {
		close(ch)
	}
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
	o.confirmation = &SubscriptionConfirmation{}
	o.c <- o
	close(o.c)
	o.confirmed = true
}

func (o *subscribeToStreamOperation) streamEventAppeared(payload []byte) {
	msg := &protobuf.StreamEventAppeared{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}
	evt := newResolvedEventFrom(msg.Event)
	for _, ch := range o.events {
		ch <- evt
	}
}

func (o *subscribeToStreamOperation) subscriptionDropped(payload []byte) {
	msg := &protobuf.SubscriptionDropped{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}
	evt := &SubscriptionDropped{
		Reason: SubscriptionDropReason(msg.GetReason()),
	}
	for _, ch := range o.dropped {
		ch <- evt
		close(ch)
	}
	for _, ch := range o.events {
		close(ch)
	}
	o.isCompleted = true
}

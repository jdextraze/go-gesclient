package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

// TODO define content
type SubscriptionConfirmation struct{}

type SubscriptionDropped struct{}

type Subscription interface {
	Confirmation() *SubscriptionConfirmation
	Events() chan *ResolvedEvent
	Unsubscribe() error
	Dropped() chan *SubscriptionDropped
}

type subscribeToStreamOperation struct {
	*BaseOperation
	stream       string
	c            chan Subscription
	confirmation *SubscriptionConfirmation
	conn         *connection
	events       []chan *ResolvedEvent
	dropped      []chan *SubscriptionDropped
	confirmed    bool
}

func newSubscribeToStreamOperation(
	stream string,
	c chan Subscription,
	conn *connection,
) *subscribeToStreamOperation {
	return &subscribeToStreamOperation{
		BaseOperation: &BaseOperation{
			correlationId: uuid.NewV4(),
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

func (o *subscribeToStreamOperation) ParseResponse(cmd tcpCommand, msg proto.Message) {
	switch cmd {
	case tcpCommand_SubscriptionConfirmation:
		o.subscriptionConfirmation(msg.(*protobuf.SubscriptionConfirmation))
	case tcpCommand_StreamEventAppeared:
		o.streamEventAppeared(msg.(*protobuf.StreamEventAppeared))
	case tcpCommand_SubscriptionDropped:
		o.subscriptionDropped(msg.(*protobuf.SubscriptionDropped))
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

	o.conn.output <- newTcpPacket(
		tcpCommand_UnsubscribeFromStream,
		0,
		o.correlationId,
		payload,
	)

	return nil
}

func (o *subscribeToStreamOperation) Dropped() chan *SubscriptionDropped {
	dropped := make(chan *SubscriptionDropped)
	o.dropped = append(o.dropped, dropped)
	return dropped
}

func (o *subscribeToStreamOperation) SetError(err error) error {
	if !o.confirmed {
		close(o.c)
	}
	evt := &SubscriptionDropped{}
	for _, ch := range o.dropped {
		ch <- evt
		close(ch)
	}
	for _, ch := range o.events {
		close(ch)
	}
	return err
}

func (o *subscribeToStreamOperation) subscriptionConfirmation(msg *protobuf.SubscriptionConfirmation) {
	if o.confirmed {
		return
	}
	o.confirmation = &SubscriptionConfirmation{}
	o.c <- o
	close(o.c)
	o.confirmed = true
}

func (o *subscribeToStreamOperation) streamEventAppeared(msg *protobuf.StreamEventAppeared) {
	evt := newResolvedEventFrom(msg.Event)
	for _, ch := range o.events {
		ch <- evt
	}
}

func (o *subscribeToStreamOperation) subscriptionDropped(msg *protobuf.SubscriptionDropped) {
	evt := &SubscriptionDropped{}
	for _, ch := range o.dropped {
		ch <- evt
		close(ch)
	}
	for _, ch := range o.events {
		close(ch)
	}
	o.isCompleted = true
}

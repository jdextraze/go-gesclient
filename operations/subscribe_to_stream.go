package operations

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/jdextraze/go-gesclient/models"
)

type subscribeToStream struct {
	*baseOperation
	stream        string
	resultChannel chan models.Subscription
	confirmation  *models.SubscriptionConfirmation
	conn          models.InternalConnection
	events        chan *models.ResolvedEvent
	dropped       chan *models.SubscriptionDropped
	confirmed     bool
	error         error
	unsubscribe   bool
}

func NewSubscribeToStream(
	stream string,
	conn models.InternalConnection,
	userCredentials *models.UserCredentials,
	resultChannel chan models.Subscription,
) *subscribeToStream {
	return &subscribeToStream{
		baseOperation: &baseOperation{
			correlationId:   uuid.NewV4(),
			userCredentials: userCredentials,
		},
		stream:        stream,
		resultChannel: resultChannel,
		conn:          conn,
		events:        make(chan *models.ResolvedEvent, 1000),
		dropped:       make(chan *models.SubscriptionDropped, 1),
	}
}

func (o *subscribeToStream) GetRequestCommand() models.Command {
	if o.unsubscribe {
		return models.Command_UnsubscribeFromStream
	}
	return models.Command_SubscribeToStream
}

func (o *subscribeToStream) GetRequestMessage() proto.Message {
	if o.unsubscribe {
		return &protobuf.UnsubscribeFromStream{}
	}
	no := false
	return &protobuf.SubscribeToStream{
		EventStreamId:  &o.stream,
		ResolveLinkTos: &no,
	}
}

func (o *subscribeToStream) ParseResponse(p *models.Package) {
	switch p.Command {
	case models.Command_SubscriptionConfirmation:
		o.subscriptionConfirmation(p.Data)
	case models.Command_StreamEventAppeared:
		o.streamEventAppeared(p.Data)
	case models.Command_SubscriptionDropped:
		o.subscriptionDropped(p.Data)
	default:
		o.handleError(p, models.Command_StreamEventAppeared)
	}
}

func (o *subscribeToStream) Confirmation() *models.SubscriptionConfirmation {
	return o.confirmation
}

func (o *subscribeToStream) Events() <-chan *models.ResolvedEvent {
	return o.events
}

func (o *subscribeToStream) Unsubscribe() error {
	if !o.confirmed {
		return errors.New("Subscription not confirmed")
	}
	o.unsubscribe = true
	return o.conn.EnqueueOperation(o)
}

func (o *subscribeToStream) Dropped() <-chan *models.SubscriptionDropped {
	return o.dropped
}

func (o *subscribeToStream) Error() error { return o.error }

func (o *subscribeToStream) Fail(err error) {
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

func (o *subscribeToStream) subscriptionConfirmation(payload []byte) {
	if o.confirmed {
		return
	}

	msg := &protobuf.SubscriptionConfirmation{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	o.confirmation = models.NewSubscriptionConfirmation(msg.GetLastCommitPosition(), msg.GetLastEventNumber())
	o.resultChannel <- o
	close(o.resultChannel)
	o.confirmed = true
}

func (o *subscribeToStream) streamEventAppeared(payload []byte) {
	msg := &protobuf.StreamEventAppeared{}
	if err := proto.Unmarshal(payload, msg); err != nil {
		o.Fail(err)
		return
	}

	o.events <- models.NewResolvedEventFrom(msg.Event)
}

func (o *subscribeToStream) subscriptionDropped(payload []byte) {
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

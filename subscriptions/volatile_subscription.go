package subscriptions

import (
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
)

type VolatileSubscription struct {
	*subscriptionBase
}

func NewVolatileSubscription(
	source *tasks.CompletionSource,
	streamId string,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	eventAppeared client.EventAppearedHandler,
	subscriptionDropped client.SubscriptionDroppedHandler,
	verboseLogging bool,
	getConnection GetConnectionHandler,
) *VolatileSubscription {
	obj := &VolatileSubscription{}
	obj.subscriptionBase = newSubscriptionBase(source, streamId, resolveLinkTos, userCredentials, eventAppeared,
		subscriptionDropped, verboseLogging, getConnection, obj.createSubscriptionPackage, obj.inspectPackage,
		obj.createSubscriptionObject)
	return obj
}

func (s *VolatileSubscription) createSubscriptionPackage() (*client.Package, error) {
	dto := &messages.SubscribeToStream{
		EventStreamId:  &s.streamId,
		ResolveLinkTos: &s.resolveLinkTos,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return nil, err
	}
	var flags client.TcpFlag
	if s.userCredentials != nil {
		flags = client.FlagsAuthenticated
	}
	return client.NewTcpPackage(client.Command_SubscribeToStream, flags, s.correlationId, data, s.userCredentials), nil
}

func (s *VolatileSubscription) inspectPackage(p *client.Package) (bool, *client.InspectionResult, error) {
	if p.Command() == client.Command_SubscriptionConfirmation {
		dto := &messages.SubscriptionConfirmation{}
		if err := proto.Unmarshal(p.Data(), dto); err != nil {
			return false, nil, err
		}
		lastEventNumber := int(dto.GetLastEventNumber())
		if err := s.confirmSubscription(dto.GetLastCommitPosition(), &lastEventNumber); err != nil {
			return false, nil, err
		}
		return true, client.NewInspectionResult(client.InspectionDecision_Subscribed, "SubscriptionConfirmation",
			nil, nil), nil
	} else if p.Command() == client.Command_StreamEventAppeared {
		dto := &messages.StreamEventAppeared{}
		if err := proto.Unmarshal(p.Data(), dto); err != nil {
			return false, nil, err
		}
		if err := s.eventAppeared(client.NewResolvedEventFrom(dto.Event)); err != nil {
			return false, nil, err
		}
		return true, client.NewInspectionResult(client.InspectionDecision_DoNothing, "StreamEventAppeared",
			nil, nil), nil
	}
	return false, nil, nil
}

func (s *VolatileSubscription) createSubscriptionObject(lastCommitPosition int64, lastEventNumber *int,
) (interface{}, client.EventStoreSubscription, error) {
	obj := NewVolatileEventStoreSubscription(s, s.streamId, lastCommitPosition, lastEventNumber)
	return obj, obj.EventStoreSubscription, nil
}

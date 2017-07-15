package subscriptions

import (
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/tasks"
)

type VolatileSubscription struct {
	*subscriptionBase
}

func NewVolatileSubscription(
	source *tasks.CompletionSource,
	streamId string,
	resolveLinkTos bool,
	userCredentials *models.UserCredentials,
	eventAppeared models.EventAppearedHandler,
	subscriptionDropped models.SubscriptionDroppedHandler,
	verboseLogging bool,
	getConnection GetConnectionHandler,
) *VolatileSubscription {
	obj := &VolatileSubscription{}
	obj.subscriptionBase = newSubscriptionBase(source, streamId, resolveLinkTos, userCredentials, eventAppeared,
		subscriptionDropped, verboseLogging, getConnection, obj.createSubscriptionPackage, obj.inspectPackage,
		obj.createSubscriptionObject)
	return obj
}

func (s *VolatileSubscription) createSubscriptionPackage() (*models.Package, error) {
	dto := &protobuf.SubscribeToStream{
		EventStreamId: &s.streamId,
		ResolveLinkTos: &s.resolveLinkTos,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return nil, err
	}
	var flags byte
	if s.userCredentials != nil {
		flags = models.FlagsAuthenticated
	}
	return models.NewTcpPackage(models.Command_SubscribeToStream, flags, s.correlationId, data, s.userCredentials), nil
}

func (s *VolatileSubscription) inspectPackage(p *models.Package) (bool, *models.InspectionResult, error) {
	if p.Command() == models.Command_SubscriptionConfirmation {
		dto := &protobuf.SubscriptionConfirmation{}
		if err := proto.Unmarshal(p.Data(), dto); err != nil {
			return false, nil, err
		}
		lastEventNumber := int(dto.GetLastEventNumber())
		if err := s.confirmSubscription(dto.GetLastCommitPosition(), &lastEventNumber); err != nil {
			return false, nil, err
		}
		return true, models.NewInspectionResult(models.InspectionDecision_Subscribed, "SubscriptionConfirmation",
			nil, nil), nil
	} else if p.Command() == models.Command_StreamEventAppeared {
		dto := &protobuf.StreamEventAppeared{}
		if err := proto.Unmarshal(p.Data(), dto); err != nil {
			return false, nil, err
		}
		if err := s.eventAppeared(models.NewResolvedEventFrom(dto.Event)); err != nil {
			return false, nil, err
		}
		return true, models.NewInspectionResult(models.InspectionDecision_DoNothing, "StreamEventAppeared",
			nil, nil), nil
	}
	return false, nil, nil
}

func (s *VolatileSubscription) createSubscriptionObject(lastCommitPosition int64, lastEventNumber *int,
) (*models.EventStoreSubscription, error) {
	subscription := NewVolatileEventStoreSubscription(s, s.streamId, lastCommitPosition, lastEventNumber)
	return subscription.EventStoreSubscription, nil
}

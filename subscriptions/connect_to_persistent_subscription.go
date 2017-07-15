package subscriptions

import (
	"github.com/jdextraze/go-gesclient/models"
	"github.com/jdextraze/go-gesclient/protobuf"
	"github.com/golang/protobuf/proto"
	"errors"
	"github.com/satori/go.uuid"
	"github.com/jdextraze/go-gesclient/tasks"
)

type connectToPersistentSubscription struct {
	*subscriptionBase
	groupName      string
	bufferSize     int
	subscriptionId string
}

func NewConnectToPersistentSubscription(
	source *tasks.CompletionSource,
	groupName string,
	bufferSize int,
	streamId string,
	userCredentials *models.UserCredentials,
	eventAppeared models.EventAppearedHandler,
	subscriptionDropped models.SubscriptionDroppedHandler,
	verboseLogging bool,
	getConnection GetConnectionHandler,
) *connectToPersistentSubscription {
	obj := &connectToPersistentSubscription{
		groupName:  groupName,
		bufferSize: bufferSize,
	}
	obj.subscriptionBase = newSubscriptionBase(source, streamId, false, userCredentials, eventAppeared,
		subscriptionDropped, verboseLogging, getConnection, obj.createSubscriptionPackage,
		obj.inspectPackage, obj.createSubscriptionObject)
	return obj
}

func (s *connectToPersistentSubscription) createSubscriptionPackage() (*models.Package, error) {
	bufferSize := int32(s.bufferSize)
	dto := &protobuf.ConnectToPersistentSubscription{
		SubscriptionId: &s.groupName,
		EventStreamId: &s.streamId,
		AllowedInFlightMessages: &bufferSize,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return nil, err
	}
	var flags byte
	if s.userCredentials != nil {
		flags = models.FlagsAuthenticated
	}
	return models.NewTcpPackage(models.Command_ConnectToPersistentSubscription, flags, s.correlationId, data,
		s.userCredentials), nil
}

func (s *connectToPersistentSubscription) inspectPackage(
	p *models.Package,
) (ok bool, result *models.InspectionResult, err error) {
	switch p.Command() {
	case models.Command_PersistentSubscriptionConfirmation:
		dto := &protobuf.PersistentSubscriptionConfirmation{}
		if err = proto.Unmarshal(p.Data(), dto); err != nil {
			break
		}
		lastEventNumber := int(dto.GetLastEventNumber())
		if err = s.confirmSubscription(dto.GetLastCommitPosition(), &lastEventNumber); err != nil {
			break
		}
		s.subscriptionId = *dto.SubscriptionId
		result = models.NewInspectionResult(models.InspectionDecision_DoNothing, "SubscriptionConfirmation", nil, nil)
	case models.Command_PersistentSubscriptionStreamEventAppeared:
		dto := &protobuf.PersistentSubscriptionStreamEventAppeared{}
		if err = proto.Unmarshal(p.Data(), dto); err != nil {
			break
		}
		if err = s.eventAppeared(models.NewResolvedEvent(dto.Event)); err != nil {
			break
		}
		result = models.NewInspectionResult(models.InspectionDecision_DoNothing, "StreamEventAppeard", nil, nil)
	case models.Command_SubscriptionDropped:
		dto := &protobuf.SubscriptionDropped{}
		if err = proto.Unmarshal(p.Data(), dto); err != nil {
			break
		}
		switch dto.GetReason() {
		case protobuf.SubscriptionDropped_AccessDenied:
			err = s.DropSubscription(models.SubscriptionDropReason_AccessDenied,
				errors.New("You do not have access to the stream."), nil)
		case protobuf.SubscriptionDropped_NotFound:
			err = s.DropSubscription(models.SubscriptionDropReason_NotFound,
				errors.New("Subscription not found"), nil)
		case protobuf.SubscriptionDropped_PersistentSubscriptionDeleted:
			err = s.DropSubscription(models.SubscriptionDropReason_PersistentSubscriptionDeleted,
				errors.New("Persistent subscription deleted"), nil)
		case protobuf.SubscriptionDropped_SubscriberMaxCountReached:
			err = s.DropSubscription(models.SubscriptionDropReason_MaxSubscribersReached,
				errors.New("Max subscribers reached"), nil)
		default:
			conn, _ := s.getConnection()
			err = s.DropSubscription(models.SubscriptionDropReason(*dto.Reason), nil, conn)
		}
		if err != nil {
			break
		}
		result = models.NewInspectionResult(models.InspectionDecision_EndOperation, "SubscriptionDropped", nil, nil)
	}
	ok = result != nil
	return
}

func (s *connectToPersistentSubscription) createSubscriptionObject(
	lastCommitPosition int64,
	lastEventNumber *int,
) (*models.EventStoreSubscription, error) {
	subscription := models.NewPersistentEventStoreSubscription(s, s.streamId, lastCommitPosition, lastEventNumber)
	return subscription.EventStoreSubscription, nil
}

func (s *connectToPersistentSubscription) NotifyEventsProcessed(processedEvents []uuid.UUID) error {
	if processedEvents == nil {
		panic("processedEvents is nil")
	}

	processedEventIds := make([][]byte, len(processedEvents))
	for i, e := range processedEvents {
		processedEventIds[i] = e.Bytes()
	}

	dto := &protobuf.PersistentSubscriptionAckEvents{
		SubscriptionId: &s.subscriptionId,
		ProcessedEventIds: processedEventIds,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return err
	}

	var flags byte
	if s.userCredentials != nil {
		flags = models.FlagsAuthenticated
	}

	pkg := models.NewTcpPackage(models.Command_PersistentSubscriptionAckEvents, flags, s.correlationId, data,
		s.userCredentials)

	return s.enqueueSend(pkg)
}

func (s *connectToPersistentSubscription) NotifyEventsFailed(
	processedEvents []uuid.UUID,
	action models.PersistentSubscriptionNakEventAction,
	reason string,
) error {
	if processedEvents == nil {
		panic("processedEvents is nil")
	}

	processedEventIds := make([][]byte, len(processedEvents))
	for i, e := range processedEvents {
		processedEventIds[i] = e.Bytes()
	}

	nakAction := protobuf.PersistentSubscriptionNakEvents_NakAction(action)
	dto := &protobuf.PersistentSubscriptionNakEvents{
		SubscriptionId: &s.subscriptionId,
		ProcessedEventIds: processedEventIds,
		Action: &nakAction,
		Message: &reason,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return err
	}

	var flags byte
	if s.userCredentials != nil {
		flags = models.FlagsAuthenticated
	}

	pkg := models.NewTcpPackage(models.Command_PersistentSubscriptionAckEvents, flags, s.correlationId, data,
		s.userCredentials)

	return s.enqueueSend(pkg)
}

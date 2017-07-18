package subscriptions

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
	"github.com/satori/go.uuid"
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
	userCredentials *client.UserCredentials,
	eventAppeared client.EventAppearedHandler,
	subscriptionDropped client.SubscriptionDroppedHandler,
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

func (s *connectToPersistentSubscription) createSubscriptionPackage() (*client.Package, error) {
	bufferSize := int32(s.bufferSize)
	dto := &messages.ConnectToPersistentSubscription{
		SubscriptionId:          &s.groupName,
		EventStreamId:           &s.streamId,
		AllowedInFlightMessages: &bufferSize,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return nil, err
	}
	var flags byte
	if s.userCredentials != nil {
		flags = client.FlagsAuthenticated
	}
	return client.NewTcpPackage(client.Command_ConnectToPersistentSubscription, flags, s.correlationId, data,
		s.userCredentials), nil
}

func (s *connectToPersistentSubscription) inspectPackage(
	p *client.Package,
) (ok bool, result *client.InspectionResult, err error) {
	switch p.Command() {
	case client.Command_PersistentSubscriptionConfirmation:
		dto := &messages.PersistentSubscriptionConfirmation{}
		if err = proto.Unmarshal(p.Data(), dto); err != nil {
			break
		}
		lastEventNumber := int(dto.GetLastEventNumber())
		if err = s.confirmSubscription(dto.GetLastCommitPosition(), &lastEventNumber); err != nil {
			break
		}
		s.subscriptionId = *dto.SubscriptionId
		result = client.NewInspectionResult(client.InspectionDecision_DoNothing, "SubscriptionConfirmation", nil, nil)
	case client.Command_PersistentSubscriptionStreamEventAppeared:
		dto := &messages.PersistentSubscriptionStreamEventAppeared{}
		if err = proto.Unmarshal(p.Data(), dto); err != nil {
			break
		}
		if err = s.eventAppeared(client.NewResolvedEvent(dto.Event)); err != nil {
			break
		}
		result = client.NewInspectionResult(client.InspectionDecision_DoNothing, "StreamEventAppeard", nil, nil)
	case client.Command_SubscriptionDropped:
		dto := &messages.SubscriptionDropped{}
		if err = proto.Unmarshal(p.Data(), dto); err != nil {
			break
		}
		switch dto.GetReason() {
		case messages.SubscriptionDropped_AccessDenied:
			err = s.DropSubscription(client.SubscriptionDropReason_AccessDenied,
				errors.New("You do not have access to the stream."), nil)
		case messages.SubscriptionDropped_NotFound:
			err = s.DropSubscription(client.SubscriptionDropReason_NotFound,
				errors.New("Subscription not found"), nil)
		case messages.SubscriptionDropped_PersistentSubscriptionDeleted:
			err = s.DropSubscription(client.SubscriptionDropReason_PersistentSubscriptionDeleted,
				errors.New("Persistent subscription deleted"), nil)
		case messages.SubscriptionDropped_SubscriberMaxCountReached:
			err = s.DropSubscription(client.SubscriptionDropReason_MaxSubscribersReached,
				errors.New("Max subscribers reached"), nil)
		default:
			conn, _ := s.getConnection()
			err = s.DropSubscription(client.SubscriptionDropReason(*dto.Reason), nil, conn)
		}
		if err != nil {
			break
		}
		result = client.NewInspectionResult(client.InspectionDecision_EndOperation, "SubscriptionDropped", nil, nil)
	}
	ok = result != nil
	return
}

func (s *connectToPersistentSubscription) createSubscriptionObject(
	lastCommitPosition int64,
	lastEventNumber *int,
) (*client.EventStoreSubscription, error) {
	subscription := client.NewPersistentEventStoreSubscription(s, s.streamId, lastCommitPosition, lastEventNumber)
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

	dto := &messages.PersistentSubscriptionAckEvents{
		SubscriptionId:    &s.subscriptionId,
		ProcessedEventIds: processedEventIds,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return err
	}

	var flags byte
	if s.userCredentials != nil {
		flags = client.FlagsAuthenticated
	}

	pkg := client.NewTcpPackage(client.Command_PersistentSubscriptionAckEvents, flags, s.correlationId, data,
		s.userCredentials)

	return s.enqueueSend(pkg)
}

func (s *connectToPersistentSubscription) NotifyEventsFailed(
	processedEvents []uuid.UUID,
	action client.PersistentSubscriptionNakEventAction,
	reason string,
) error {
	if processedEvents == nil {
		panic("processedEvents is nil")
	}

	processedEventIds := make([][]byte, len(processedEvents))
	for i, e := range processedEvents {
		processedEventIds[i] = e.Bytes()
	}

	nakAction := messages.PersistentSubscriptionNakEvents_NakAction(action)
	dto := &messages.PersistentSubscriptionNakEvents{
		SubscriptionId:    &s.subscriptionId,
		ProcessedEventIds: processedEventIds,
		Action:            &nakAction,
		Message:           &reason,
	}
	data, err := proto.Marshal(dto)
	if err != nil {
		return err
	}

	var flags byte
	if s.userCredentials != nil {
		flags = client.FlagsAuthenticated
	}

	pkg := client.NewTcpPackage(client.Command_PersistentSubscriptionAckEvents, flags, s.correlationId, data,
		s.userCredentials)

	return s.enqueueSend(pkg)
}

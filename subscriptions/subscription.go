package subscriptions

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	log "github.com/jdextraze/go-gesclient/logger"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/jdextraze/go-gesclient/tasks"
	"github.com/satori/go.uuid"
	"net"
	"sync/atomic"
)

type Subscription interface {
	DropSubscription(reason client.SubscriptionDropReason, err error, connection *client.PackageConnection) error
	ConnectionClosed() error
	InspectPackage(p *client.Package) (*client.InspectionResult, error)
	Subscribe(correlationId uuid.UUID, connection *client.PackageConnection) (bool, error)
}

type GetConnectionHandler func() (*client.PackageConnection, error)
type CreateSubscriptionPackageHandler func() (*client.Package, error)
type InspectPackageHandler func(p *client.Package) (bool, *client.InspectionResult, error)
type CreateSubscriptionObjectHandler func(lastCommitPosition int64, lastEventNumber *int) (
	interface{}, client.EventStoreSubscription, error)
type ActionHandler func() error

type subscriptionBase struct {
	source              *tasks.CompletionSource
	streamId            string
	resolveLinkTos      bool
	userCredentials     *client.UserCredentials
	_eventAppeared      client.EventAppearedHandler
	subscriptionDropped client.SubscriptionDroppedHandler
	verboseLogging      bool
	getConnection       GetConnectionHandler
	maxQueueSize        int
	actionQueue         chan ActionHandler
	actionExecuting     int32
	subscription        client.EventStoreSubscription
	unsubscribed        int32
	correlationId       uuid.UUID

	createSubscriptionPackage CreateSubscriptionPackageHandler
	inspectPackage            InspectPackageHandler
	createSubscriptionObject  CreateSubscriptionObjectHandler
}

func newSubscriptionBase(
	source *tasks.CompletionSource,
	streamId string,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	eventAppeared client.EventAppearedHandler,
	subscriptionDropped client.SubscriptionDroppedHandler,
	verboseLogging bool,
	getConnection GetConnectionHandler,
	createSubscriptionPackage CreateSubscriptionPackageHandler,
	inspectPackage InspectPackageHandler,
	createSubscriptionObject CreateSubscriptionObjectHandler,
) *subscriptionBase {
	if source == nil {
		panic("source is nil")
	}
	if eventAppeared == nil {
		panic("eventAppeared is nil")
	}
	if getConnection == nil {
		panic("getConnection is nil")
	}
	if subscriptionDropped == nil {
		subscriptionDropped = client.SubscriptionDroppedHandler(
			func(s client.EventStoreSubscription, dr client.SubscriptionDropReason, err error) error { return nil })
	}
	return &subscriptionBase{
		source:                    source,
		streamId:                  streamId,
		resolveLinkTos:            resolveLinkTos,
		userCredentials:           userCredentials,
		_eventAppeared:            eventAppeared,
		subscriptionDropped:       subscriptionDropped,
		verboseLogging:            verboseLogging,
		getConnection:             getConnection,
		maxQueueSize:              2000,
		actionQueue:               make(chan ActionHandler, 2000),
		createSubscriptionPackage: createSubscriptionPackage,
		inspectPackage:            inspectPackage,
		createSubscriptionObject:  createSubscriptionObject,
	}
}

func (s *subscriptionBase) enqueueSend(p *client.Package) error {
	conn, err := s.getConnection()
	if err != nil {
		return err
	}
	return conn.EnqueueSend(p)
}

func (s *subscriptionBase) Subscribe(correlationId uuid.UUID, connection *client.PackageConnection) (bool, error) {
	if connection == nil {
		panic("connection is nil")
	}

	if s.subscription != nil || s.unsubscribed != 0 {
		return false, nil
	}

	s.correlationId = correlationId
	if pkg, err := s.createSubscriptionPackage(); err != nil {
		return false, err
	} else if err := connection.EnqueueSend(pkg); err != nil {
		return false, err
	}
	return true, nil
}

func (s *subscriptionBase) Unsubscribe() error {
	conn, err := s.getConnection()
	if err != nil {
		return err
	}
	return s.DropSubscription(client.SubscriptionDropReason_UserInitiated, nil, conn)
}

func (s *subscriptionBase) createUnsubscriptionPackage() *client.Package {
	data, _ := proto.Marshal(&messages.UnsubscribeFromStream{})
	return client.NewTcpPackage(client.Command_UnsubscribeFromStream, client.FlagsNone, s.correlationId, data, nil)
}

func (s *subscriptionBase) InspectPackage(p *client.Package) (*client.InspectionResult, error) {
	ok, result, err := s.inspectPackage(p)
	if ok {
		return result, nil
	} else if err == nil {
		switch p.Command() {
		case client.Command_StreamEventAppeared:
			dto := &messages.StreamEventAppeared{}
			if err = proto.Unmarshal(p.Data(), dto); err != nil {
				break
			}
			if err = s.eventAppeared(client.NewResolvedEventFrom(dto.Event)); err != nil {
				break
			}
			return client.NewInspectionResult(client.InspectionDecision_DoNothing, "StreamEventAppeared", nil, nil), nil
		case client.Command_SubscriptionDropped:
			dto := &messages.SubscriptionDropped{}
			if err = proto.Unmarshal(p.Data(), dto); err != nil {
				break
			}
			switch dto.GetReason() {
			case messages.SubscriptionDropped_Unsubscribed:
				err = s.DropSubscription(client.SubscriptionDropReason_UserInitiated, nil, nil)
			case messages.SubscriptionDropped_AccessDenied:
				err = s.DropSubscription(client.SubscriptionDropReason_AccessDenied,
					fmt.Errorf("%s failed due to access denied.", s.String()), nil)
			case messages.SubscriptionDropped_NotFound:
				err = s.DropSubscription(client.SubscriptionDropReason_NotFound,
					fmt.Errorf("%s failed due to not found.", s.String()), nil)
			default:
				if s.verboseLogging {
					log.Debugf("Subscription dropped by server. Reason: %s", dto.Reason)
				}
				err = s.DropSubscription(client.SubscriptionDropReason_Unknown,
					fmt.Errorf("Unsubscribe reason: %s.", dto.Reason), nil)
			}
			if err != nil {
				break
			}
			return client.NewInspectionResult(client.InspectionDecision_EndOperation,
				fmt.Sprintf("SubscriptionDropped: %s", dto.Reason), nil, nil), nil
		case client.Command_NotAuthenticated:
			message := string(p.Data())
			if message == "" {
				message = "Authentication error"
			}
			if err = s.DropSubscription(client.SubscriptionDropReason_NotAuthenticated, errors.New(message), nil); err != nil {
				break
			}
			return client.NewInspectionResult(client.InspectionDecision_EndOperation, "NotAuthenticated", nil, nil), nil
		case client.Command_BadRequest:
			message := string(p.Data())
			if message == "" {
				message = "<no message>"
			}
			if err = s.DropSubscription(client.SubscriptionDropReason_ServerError, errors.New(message), nil); err != nil {
				break
			}
			return client.NewInspectionResult(client.InspectionDecision_EndOperation, "NotAuthenticated", nil, nil), nil
		case client.Command_NotHandled:
			if s.subscription != nil {
				err = errors.New("NotHandled command appeared while we are already subscribed.")
				break
			}
			dto := &messages.NotHandled{}
			if err = proto.Unmarshal(p.Data(), dto); err != nil {
				break
			}
			switch dto.GetReason() {
			case messages.NotHandled_NotReady:
				return client.NewInspectionResult(client.InspectionDecision_Retry, "NotHandled - NotReady", nil, nil), nil
			case messages.NotHandled_TooBusy:
				return client.NewInspectionResult(client.InspectionDecision_Retry, "NotHandled - TooBusy", nil, nil), nil
			case messages.NotHandled_NotMaster:
				masterInfo := &messages.NotHandled_MasterInfo{}
				if err = proto.Unmarshal(dto.AdditionalInfo, masterInfo); err != nil {
					break
				}
				var tcpEndpoint *net.TCPAddr
				var secureTcpEndpoint *net.TCPAddr
				tcpEndpoint, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", masterInfo.ExternalTcpAddress,
					masterInfo.ExternalTcpPort))
				secureTcpEndpoint, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d",
					masterInfo.ExternalSecureTcpAddress, masterInfo.ExternalSecureTcpPort))
				return client.NewInspectionResult(client.InspectionDecision_Retry, "NotHandled - NotMaster",
					tcpEndpoint, secureTcpEndpoint), nil
			default:
				log.Errorf("Unknown NotHandledReason: %s.", dto.Reason)
				return client.NewInspectionResult(client.InspectionDecision_Retry, "NotHandler - <unknown>", nil, nil), nil
			}
		default:
			s.DropSubscription(client.SubscriptionDropReason_ServerError, fmt.Errorf("Command not expected: %s",
				p.Command()), nil)
			return client.NewInspectionResult(client.InspectionDecision_EndOperation, p.Command().String(), nil, nil), nil
		}
	}
	if err := s.DropSubscription(client.SubscriptionDropReason_Unknown, err, nil); err != nil {
		return nil, err
	}
	return client.NewInspectionResult(client.InspectionDecision_EndOperation, fmt.Sprintf("Exception - %v", err), nil,
		nil), nil
}

func (s *subscriptionBase) ConnectionClosed() error {
	return s.DropSubscription(client.SubscriptionDropReason_ConnectionClosed, errors.New("Connection was closed"), nil)
}

func (s *subscriptionBase) timeoutSubscription() (bool, error) {
	if s.subscription != nil {
		return false, nil
	}
	if err := s.DropSubscription(client.SubscriptionDropReason_SubscribingError, nil, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (s *subscriptionBase) DropSubscription(
	reason client.SubscriptionDropReason,
	err error,
	connection *client.PackageConnection,
) error {
	if atomic.CompareAndSwapInt32(&s.unsubscribed, 0, 1) {
		if s.verboseLogging {
			log.Debugf("%s (%s): closing subscription, reason: %s, error: %s", s.String(), s.correlationId, reason, err)
		}

		if reason != client.SubscriptionDropReason_UserInitiated {
			if err == nil {
				return fmt.Errorf("No error provided for subscription drop reason '%s'", reason)
			}
			s.source.TrySetError(err)
		}

		if reason == client.SubscriptionDropReason_UserInitiated && s.subscription != nil && connection != nil {
			connection.EnqueueSend(s.createUnsubscriptionPackage())
		}

		if s.subscription != nil {
			s.executeActionAsync(func() error {
				return s.subscriptionDropped(s.subscription, reason, err)
			})
		}
	}
	return nil
}

func (s *subscriptionBase) confirmSubscription(lastCommitPosition int64, lastEventNumber *int) error {
	if lastCommitPosition < -1 {
		return fmt.Errorf("lastCommitPosition %d is out of range", lastCommitPosition)
	}
	if s.subscription != nil {
		return errors.New("Double confirmation of subscription")
	}

	if s.verboseLogging {
		log.Debugf("%s (%s): subscribed at CommitPosition: %d, EventNumber: %v", s.String(), s.correlationId,
			lastCommitPosition, lastEventNumber)
	}
	sub, ess, err := s.createSubscriptionObject(lastCommitPosition, lastEventNumber)
	if err != nil {
		return err
	}
	s.subscription = ess
	s.source.SetResult(sub)
	return nil
}

func (s *subscriptionBase) eventAppeared(event *client.ResolvedEvent) error {
	if s.unsubscribed != 0 {
		return nil
	}

	if s.subscription == nil {
		return errors.New("Subscription not confirmed, but event appeared!")
	}

	if s.verboseLogging {
		log.Debugf("%s (%s): event appeared (%s, %d, %s @ %d)", s.String(), s.correlationId, event.OriginalStreamId(),
			event.OriginalEventNumber(), event.OriginalEvent().EventType(), event.OriginalPosition())
	}

	s.executeActionAsync(func() error {
		return s._eventAppeared(s.subscription, event)
	})
	return nil
}

func (s *subscriptionBase) executeActionAsync(action ActionHandler) {
	if len(s.actionQueue) >= cap(s.actionQueue) {
		s.DropSubscription(client.SubscriptionDropReason_UserInitiated, errors.New("client buffer too big"), nil)
	}
	s.actionQueue <- action
	if atomic.CompareAndSwapInt32(&s.actionExecuting, 0, 1) {
		go s.executeActions()
	}
}

func (s *subscriptionBase) executeActions() {
	for {
		for len(s.actionQueue) > 0 {
			action, ok := <-s.actionQueue
			if !ok {
				break
			}
			if err := action(); err != nil {
				log.Errorf("Error during user callback execution: %v", err)
			}
		}
		atomic.SwapInt32(&s.actionExecuting, 0)
		if len(s.actionQueue) > 0 && atomic.CompareAndSwapInt32(&s.actionExecuting, 0, 1) {
			continue
		}
		break
	}
}

func (s *subscriptionBase) String() string {
	var stream string
	if s.streamId == "" {
		stream = "<all>"
	} else {
		stream = s.streamId
	}
	return fmt.Sprintf("Subscription to %s", stream)
}

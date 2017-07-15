package subscriptions

import (
	"github.com/jdextraze/go-gesclient/models"
	"github.com/satori/go.uuid"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/protobuf"
	"errors"
	"sync/atomic"
	"fmt"
	"net"
	"github.com/jdextraze/go-gesclient/tasks"
)

type Subscription interface {
	DropSubscription(reason models.SubscriptionDropReason, err error, connection *models.PackageConnection) error
	ConnectionClosed() error
	InspectPackage(p *models.Package) (*models.InspectionResult, error)
	Subscribe(correlationId uuid.UUID, connection *models.PackageConnection) (bool, error)
}

type GetConnectionHandler func() (*models.PackageConnection, error)
type CreateSubscriptionPackageHandler func() (*models.Package, error)
type InspectPackageHandler func(p *models.Package) (bool, *models.InspectionResult, error)
type CreateSubscriptionObjectHandler func(lastCommitPosition int64, lastEventNumber *int) (
	*models.EventStoreSubscription, error)
type ActionHandler func() error

type subscriptionBase struct {
	source              *tasks.CompletionSource
	streamId            string
	resolveLinkTos      bool
	userCredentials     *models.UserCredentials
	_eventAppeared      models.EventAppearedHandler
	subscriptionDropped models.SubscriptionDroppedHandler
	verboseLogging      bool
	getConnection       GetConnectionHandler
	maxQueueSize        int
	actionQueue         chan ActionHandler
	actionExecuting     int32
	subscription        *models.EventStoreSubscription
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
	userCredentials *models.UserCredentials,
	eventAppeared models.EventAppearedHandler,
	subscriptionDropped models.SubscriptionDroppedHandler,
	verboseLogging bool,
	getConnection GetConnectionHandler,
	createSubscriptionPackage CreateSubscriptionPackageHandler,
	inspectPackage InspectPackageHandler,
	createSubscriptionObject CreateSubscriptionObjectHandler,
) *subscriptionBase {
	if log == nil {
		panic("log is nil")
	}
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
		subscriptionDropped = models.SubscriptionDroppedHandler(
			func(s *models.EventStoreSubscription, dr models.SubscriptionDropReason, err error) error { return nil })
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

func (s *subscriptionBase) enqueueSend(p *models.Package) error {
	conn, err := s.getConnection()
	if err != nil {
		return err
	}
	return conn.EnqueueSend(p)
}

func (s *subscriptionBase) Subscribe(correlationId uuid.UUID, connection *models.PackageConnection) (bool, error) {
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
	return s.DropSubscription(models.SubscriptionDropReason_UserInitiated, nil, conn)
}

func (s *subscriptionBase) createUnsubscriptionPackage() *models.Package {
	data, _ := proto.Marshal(&protobuf.UnsubscribeFromStream{})
	return models.NewTcpPackage(models.Command_UnsubscribeFromStream, models.FlagsNone, s.correlationId, data,
		s.userCredentials)
}

func (s *subscriptionBase) InspectPackage(p *models.Package) (*models.InspectionResult, error) {
	ok, result, err := s.inspectPackage(p)
	if ok {
		return result, nil
	} else if err == nil {
		switch p.Command() {
		case models.Command_StreamEventAppeared:
			dto := &protobuf.StreamEventAppeared{}
			if err = proto.Unmarshal(p.Data(), dto); err != nil {
				break
			}
			if err = s.eventAppeared(models.NewResolvedEventFrom(dto.Event)); err != nil {
				break
			}
			return models.NewInspectionResult(models.InspectionDecision_DoNothing, "StreamEventAppeared", nil, nil), nil
		case models.Command_SubscriptionDropped:
			dto := &protobuf.SubscriptionDropped{}
			if err = proto.Unmarshal(p.Data(), dto); err != nil {
				break
			}
			switch dto.GetReason() {
			case protobuf.SubscriptionDropped_Unsubscribed:
				err = s.DropSubscription(models.SubscriptionDropReason_UserInitiated, nil, nil)
			case protobuf.SubscriptionDropped_AccessDenied:
				err = s.DropSubscription(models.SubscriptionDropReason_AccessDenied,
					fmt.Errorf("Subscription to '%s' failed due to access denied.", s.streamId), nil)
			case protobuf.SubscriptionDropped_NotFound:
				err = s.DropSubscription(models.SubscriptionDropReason_NotFound,
					fmt.Errorf("Subscription to '%s' failed due to not found.", s.streamId), nil)
			default:
				if s.verboseLogging {
					log.Debugf("Subscription dropped by server. Reason: %s", dto.Reason)
				}
				err = s.DropSubscription(models.SubscriptionDropReason_Unknown,
					fmt.Errorf("Unsubscribe reason: %s.", dto.Reason), nil)
			}
			if err != nil {
				break
			}
			return models.NewInspectionResult(models.InspectionDecision_EndOperation,
				fmt.Sprintf("SubscriptionDropped: %s", dto.Reason), nil, nil), nil
		case models.Command_NotAuthenticated:
			message := string(p.Data())
			if message == "" {
				message = "Authentication error"
			}
			if err = s.DropSubscription(models.SubscriptionDropReason_NotAuthenticated, errors.New(message), nil); err != nil {
				break
			}
			return models.NewInspectionResult(models.InspectionDecision_EndOperation, "NotAuthenticated", nil, nil), nil
		case models.Command_BadRequest:
			message := string(p.Data())
			if message == "" {
				message = "<no message>"
			}
			if err = s.DropSubscription(models.SubscriptionDropReason_ServerError, errors.New(message), nil); err != nil {
				break
			}
			return models.NewInspectionResult(models.InspectionDecision_EndOperation, "NotAuthenticated", nil, nil), nil
		case models.Command_NotHandled:
			if s.subscription != nil {
				err = errors.New("NotHandled command appeared while we are already subscribed.")
				break
			}
			dto := &protobuf.NotHandled{}
			if err = proto.Unmarshal(p.Data(), dto); err != nil {
				break
			}
			switch dto.GetReason() {
			case protobuf.NotHandled_NotReady:
				return models.NewInspectionResult(models.InspectionDecision_Retry, "NotHandled - NotReady", nil, nil), nil
			case protobuf.NotHandled_TooBusy:
				return models.NewInspectionResult(models.InspectionDecision_Retry, "NotHandled - TooBusy", nil, nil), nil
			case protobuf.NotHandled_NotMaster:
				masterInfo := &protobuf.NotHandled_MasterInfo{}
				if err = proto.Unmarshal(dto.AdditionalInfo, masterInfo); err != nil {
					break
				}
				var tcpEndpoint *net.TCPAddr
				var secureTcpEndpoint *net.TCPAddr
				tcpEndpoint, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", masterInfo.ExternalTcpAddress,
					masterInfo.ExternalTcpPort))
				secureTcpEndpoint, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d",
					masterInfo.ExternalSecureTcpAddress, masterInfo.ExternalSecureTcpPort))
				return models.NewInspectionResult(models.InspectionDecision_Retry, "NotHandled - NotMaster",
					tcpEndpoint, secureTcpEndpoint), nil
			default:
				log.Errorf("Unknown NotHandledReason: %s.", dto.Reason)
				return models.NewInspectionResult(models.InspectionDecision_Retry, "NotHandler - <unknown>", nil, nil), nil
			}
		default:
			s.DropSubscription(models.SubscriptionDropReason_ServerError, fmt.Errorf("Command not expected: %s",
				p.Command()), nil)
			return models.NewInspectionResult(models.InspectionDecision_EndOperation, p.Command().String(), nil, nil), nil
		}
	}
	if err := s.DropSubscription(models.SubscriptionDropReason_Unknown, err, nil); err != nil {
		return nil, err
	}
	return models.NewInspectionResult(models.InspectionDecision_EndOperation, fmt.Sprintf("Exception - %v", err), nil,
		nil), nil
}

func (s *subscriptionBase) ConnectionClosed() error {
	return s.DropSubscription(models.SubscriptionDropReason_ConnectionClosed, errors.New("Connection was closed"), nil)
}

func (s *subscriptionBase) timeoutSubscription() (bool, error) {
	if s.subscription != nil {
		return false, nil
	}
	if err := s.DropSubscription(models.SubscriptionDropReason_SubscribingError, nil, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (s *subscriptionBase) DropSubscription(
	reason models.SubscriptionDropReason,
	err error,
	connection *models.PackageConnection,
) error {
	if atomic.CompareAndSwapInt32(&s.unsubscribed, 0, 1) {
		if s.verboseLogging {
			log.Debugf("Subscription %s to %s: closing subscription, reason: %s, error: %s",
				s.correlationId, s.streamId, reason, err)
		}

		if reason != models.SubscriptionDropReason_UserInitiated {
			if err == nil {
				return fmt.Errorf("No error provided for subscription drop reason '%s'", reason)
			}
			s.source.TrySetError(err)
		}

		if reason == models.SubscriptionDropReason_UserInitiated && s.subscription != nil && connection != nil {
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
		log.Debugf("Subscription %s to %s: subscribed at CommitPosition: %d, EventNumber: %d",
			s.correlationId, s.streamId, lastCommitPosition, lastEventNumber)
	}
	var err error
	s.subscription, err = s.createSubscriptionObject(lastCommitPosition, lastEventNumber)
	if err != nil {
		return err
	}
	s.source.SetResult(s.subscription)
	return nil
}

func (s *subscriptionBase) eventAppeared(event *models.ResolvedEvent) error {
	if s.unsubscribed != 0 {
		return nil
	}

	if s.subscription == nil {
		return errors.New("Subscription not confirmed, but event appeared!")
	}

	if s.verboseLogging {
		log.Debugf("Subscription %s to %s: event appeared (%s, %d, %s @ %d)",
			s.correlationId, s.streamId, event.OriginalStreamId(), event.OriginalEventNumber(),
			event.OriginalEvent().EventType(), event.OriginalPosition())
	}

	s.executeActionAsync(func() error {
		return s._eventAppeared(s.subscription, event)
	})
	return nil
}

func (s *subscriptionBase) executeActionAsync(action ActionHandler) {
	if len(s.actionQueue) >= cap(s.actionQueue) {
		s.DropSubscription(models.SubscriptionDropReason_UserInitiated, errors.New("client buffer too big"), nil)
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

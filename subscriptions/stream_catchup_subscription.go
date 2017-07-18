package subscriptions

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
)

type StreamCatchUpSubscription struct {
	*catchUpSubscription
	nextReadEventNumber      int
	lastProcessedEventNumber int
	completion               *tasks.CompletionSource
}

func NewStreamCatchUpSubscription(
	connection client.Connection,
	streamId string,
	fromEventNumberExclusive *int,
	userCredentials *client.UserCredentials,
	eventAppeared client.CatchUpEventAppearedHandler,
	liveProcessingStarted client.LiveProcessingStartedHandler,
	subscriptionDropped client.CatchUpSubscriptionDroppedHandler,
	settings *client.CatchUpSubscriptionSettings,
) *StreamCatchUpSubscription {
	if streamId == "" {
		panic("streamId is empty")
	}
	var (
		lastProcessedEventNumber int
		nextReadEventNumber      int
	)
	if fromEventNumberExclusive != nil {
		lastProcessedEventNumber = *fromEventNumberExclusive
		nextReadEventNumber = *fromEventNumberExclusive
	} else {
		lastProcessedEventNumber = -1
		nextReadEventNumber = 0
	}
	obj := &StreamCatchUpSubscription{
		nextReadEventNumber:      nextReadEventNumber,
		lastProcessedEventNumber: lastProcessedEventNumber,
	}
	obj.catchUpSubscription = newCatchUpSubscription(connection, streamId, userCredentials, eventAppeared,
		liveProcessingStarted, subscriptionDropped, settings, obj.readEventsTillAsync, obj.tryProcess)
	return obj
}

func (s *StreamCatchUpSubscription) readEventsTillAsync(
	connection client.Connection,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) *tasks.Task {
	s.completion = tasks.NewCompletionSource()
	s.readEventsInternal(connection, resolveLinkTos, userCredentials, lastCommitPosition, lastEventNumber)
	return s.completion.Task()
}

func (s *StreamCatchUpSubscription) readEventsInternal(
	connection client.Connection,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) {
	task, err := connection.ReadStreamEventsForwardAsync(s.streamId, s.nextReadEventNumber, s.readBatchSize,
		resolveLinkTos, userCredentials)
	if err == nil {
		task.ContinueWith(tasks.ContinueWithCallback(func(t *tasks.Task) error {
			return s.readEventsCallback(t, connection, resolveLinkTos, userCredentials, lastCommitPosition,
				lastEventNumber)
		}))
	} else {
		s.completion.TrySetError(err)
	}
}

func (s *StreamCatchUpSubscription) readEventsCallback(
	task *tasks.Task,
	connection client.Connection,
	resolveLinkTos bool,
	userCredentials *client.UserCredentials,
	lastCommitPosition *int64,
	lastEventNumber *int32,
) error {
	var err error
	if task.IsFaulted() {
		err = task.Wait()
	} else {
		result := &client.StreamEventsSlice{}
		if err = task.Result(result); err == nil {
			if ok, err2 := s.processEvents(lastEventNumber, result); ok && !s.shouldStop {
				s.readEventsInternal(connection, resolveLinkTos, userCredentials, lastCommitPosition, lastEventNumber)
			} else if err2 != nil {
				err = err2
			} else {
				if s.verbose {
					log.Debugf("Catch-up Subscription to %s: finished reading events, nextReadEventNumber = %s.",
						s.streamId, s.nextReadEventNumber)
				}
				res := true
				s.completion.TrySetResult(&res)
			}
		}
	}
	if err != nil {
		s.completion.TrySetError(err)
	}
	return nil
}

func (s *StreamCatchUpSubscription) processEvents(
	lastEventNumber *int32,
	slice *client.StreamEventsSlice,
) (bool, error) {
	var done bool
	switch slice.Status() {
	case client.SliceReadStatus_Success:
		for _, e := range slice.Events() {
			if err := s.tryProcess(e); err != nil {
				return false, err
			}
		}
		s.nextReadEventNumber = slice.NextEventNumber()
		if lastEventNumber == nil {
			done = slice.IsEndOfStream()
		} else {
			done = slice.NextEventNumber() > int(*lastEventNumber)
		}
	case client.SliceReadStatus_StreamNotFound:
		if lastEventNumber != nil && *lastEventNumber != -1 {
			return false, fmt.Errorf("Impossible: stream %s disappeared in the middle of catching up subscription.",
				s.streamId)
		}
		done = true
	case client.SliceReadStatus_StreamDeleted:
		return false, client.StreamDeleted
	default:
		return false, fmt.Errorf("Unexpect StreamEventsSlice.Status: %s", slice.Status())
	}

	if !done && slice.IsEndOfStream() {
		time.Sleep(time.Millisecond)
	}
	return done, nil
}

func (s *StreamCatchUpSubscription) tryProcess(e *client.ResolvedEvent) error {
	processed := false
	if e.OriginalEventNumber() > s.lastProcessedEventNumber {
		if err := s.eventAppeared(s, e); err != nil {
			return err
		}
		s.lastProcessedEventNumber = e.OriginalEventNumber()
		processed = true
	}
	if s.verbose {
		log.Debugf("Catch-up Subscription to %s: %b event (%s, %d, %s @ %s).", s.streamId, processed,
			e.OriginalEvent().EventStreamId(), e.OriginalEvent().EventNumber(), e.OriginalEvent().EventType(),
			e.OriginalPosition())
	}
	return nil
}

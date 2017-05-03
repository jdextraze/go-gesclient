package gesclient

import "github.com/jdextraze/go-gesclient/protobuf"

type EventReadResult struct {
	status      EventReadStatus
	stream      string
	eventNumber int
	event       *ResolvedEvent
	error       error
}

func newEventReadResult(
	status EventReadStatus,
	stream string,
	eventNumber int,
	event *protobuf.ResolvedIndexedEvent,
	err error,
) *EventReadResult {
	resolvedEvent := newResolvedEvent(event)
	return &EventReadResult{
		status:      status,
		stream:      stream,
		eventNumber: eventNumber,
		event:       resolvedEvent,
		error:       err,
	}
}

func (r *EventReadResult) Status() EventReadStatus {
	return r.status
}

func (r *EventReadResult) Stream() string {
	return r.stream
}

func (r *EventReadResult) EventNumber() int {
	return r.eventNumber
}

func (r *EventReadResult) Event() *ResolvedEvent {
	return r.event
}

func (r *EventReadResult) Error() error {
	return r.error
}

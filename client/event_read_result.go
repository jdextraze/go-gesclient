package client

import "github.com/jdextraze/go-gesclient/messages"

type EventReadResult struct {
	status      EventReadStatus
	stream      string
	eventNumber int
	event       *ResolvedEvent
}

func NewEventReadResult(
	status EventReadStatus,
	stream string,
	eventNumber int,
	event *messages.ResolvedIndexedEvent,
) *EventReadResult {
	resolvedEvent := NewResolvedEvent(event)
	return &EventReadResult{
		status:      status,
		stream:      stream,
		eventNumber: eventNumber,
		event:       resolvedEvent,
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

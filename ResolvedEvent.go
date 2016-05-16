package gesclient

import (
	"bitbucket.org/jdextraze/go-gesclient/protobuf"
)

type ResolvedEvent struct {
	event            *RecordedEvent
	link             *RecordedEvent
	originalPosition *Position
}

func newResolvedEventFrom(evt *protobuf.ResolvedEvent) *ResolvedEvent {
	var event *RecordedEvent
	if evt.Event != nil {
		event = newRecordedEvent(evt.Event)
	}
	var link *RecordedEvent
	if evt.Link != nil {
		link = newRecordedEvent(evt.Link)
	}
	// TODO handle error
	position, _ := NewPosition(evt.GetCommitPosition(), evt.GetPreparePosition())
	return &ResolvedEvent{
		event:            event,
		link:             link,
		originalPosition: position,
	}
}

func newResolvedEvent(evt *protobuf.ResolvedIndexedEvent) *ResolvedEvent {
	var event *RecordedEvent
	if evt.Event != nil {
		event = newRecordedEvent(evt.Event)
	}
	var link *RecordedEvent
	if evt.Link != nil {
		link = newRecordedEvent(evt.Link)
	}
	return &ResolvedEvent{
		event:            event,
		link:             link,
		originalPosition: nil,
	}
}

func (e *ResolvedEvent) Event() *RecordedEvent { return e.event }

func (e *ResolvedEvent) Link() *RecordedEvent { return e.link }

func (e *ResolvedEvent) OriginalPosition() *Position { return e.originalPosition }

func (e *ResolvedEvent) OriginalEvent() *RecordedEvent {
	if e.link == nil {
		return e.event
	}
	return e.link
}

func (e *ResolvedEvent) IsResolved() bool {
	return e.link != nil && e.event != nil
}

func (e *ResolvedEvent) OriginalStreamId() string {
	return e.OriginalEvent().EventStreamId()
}

func (e *ResolvedEvent) OriginalEventNumber() int {
	return e.OriginalEvent().EventNumber()
}

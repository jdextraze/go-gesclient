package client

import (
	"fmt"
	"github.com/jdextraze/go-gesclient/messages"
)

type ResolvedEvent struct {
	event            *RecordedEvent
	link             *RecordedEvent
	originalPosition *Position
}

func NewResolvedEventFrom(evt *messages.ResolvedEvent) *ResolvedEvent {
	var event *RecordedEvent
	if evt.Event != nil {
		event = newRecordedEvent(evt.Event)
	}
	var link *RecordedEvent
	if evt.Link != nil {
		link = newRecordedEvent(evt.Link)
	}
	position := NewPosition(evt.GetCommitPosition(), evt.GetPreparePosition())
	return &ResolvedEvent{
		event:            event,
		link:             link,
		originalPosition: position,
	}
}

func NewResolvedEvent(evt *messages.ResolvedIndexedEvent) *ResolvedEvent {
	var event *RecordedEvent
	if evt != nil && evt.Event != nil {
		event = newRecordedEvent(evt.Event)
	}
	var link *RecordedEvent
	if evt != nil && evt.Link != nil {
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

func (e *ResolvedEvent) String() string {
	var originalEvent string
	if e.link == nil {
		originalEvent = "<event>"
	} else {
		originalEvent = "<link>"
	}
	return fmt.Sprintf(
		"&{event:%+v link:%+v originalPosition:%+v originalEvent:%s isResolved:%t originalStreamId:%s originalEventNumber:%d}",
		e.event, e.link, e.originalPosition, originalEvent, e.IsResolved(), e.OriginalStreamId(), e.OriginalEventNumber())
}

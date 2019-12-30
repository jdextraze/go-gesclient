package client

import (
	"bytes"
	"github.com/jdextraze/go-gesclient/guid"
	"github.com/jdextraze/go-gesclient/messages"
	"github.com/satori/go.uuid"
	"testing"
	"time"
)

func TestRecordedEvent(t *testing.T) {
	streamId := "Test"
	number := int32(123)
	eventId := uuid.NewV4().Bytes()
	eventType := "Tested"
	dataContentType := int32(1)
	metadataContentType := int32(0)
	data := []byte{1, 2, 3}
	metadata := []byte{4, 5, 6}
	now := time.Now()
	created := now.UnixNano()/tick + ticksSinceEpoch
	createdEpoch := now.Round(time.Millisecond).UnixNano() / int64(time.Millisecond)

	e := newRecordedEvent(&messages.EventRecord{
		EventStreamId:       &streamId,
		EventNumber:         &number,
		EventId:             eventId,
		EventType:           &eventType,
		DataContentType:     &dataContentType,
		MetadataContentType: &metadataContentType,
		Data:                data,
		Metadata:            metadata,
		Created:             &created,
		CreatedEpoch:        &createdEpoch,
	})

	if e.EventStreamId() != streamId {
		t.Error("EventStreamId")
	}
	if e.EventNumber() != int(number) {
		t.Error("EventNumber")
	}
	if !bytes.Equal(e.EventId().Bytes(), guid.FromBytes(eventId).Bytes()) {
		t.Error("EventId")
	}
	if e.EventType() != eventType {
		t.Error("EventType")
	}
	if !e.IsJson() {
		t.Error("IsJson")
	}
	if e.Created().String() != now.Round(time.Duration(tick)).String() {
		t.Errorf("Created %s != %s", e.Created().String(), now.Round(time.Duration(tick)).String())
	}
	if e.CreatedEpoch().String() != now.Round(time.Millisecond).String() {
		t.Error("CreatedEpoch")
	}
}

package client_test

import (
	"bytes"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/guid"
	"github.com/jdextraze/go-gesclient/pkg/uuid"
	"testing"
)

func TestNewEventData(t *testing.T) {
	eventId := uuid.Must(uuid.NewV4())
	eventType := "type"
	isJson := true
	data := []byte(`{}`)
	metadata := []byte(`{}`)

	d := client.NewEventData(eventId, eventType, isJson, data, metadata)

	if d == nil {
		t.Fatal("NewEventData returned nil")
	}
	if !uuid.Equal(d.EventId(), eventId) {
		t.Errorf("EventId doesn't match: %s != %s", d.EventId().String(), eventId.String())
	}
	if d.Type() != eventType {
		t.Errorf("Type doesn't match: %s != %s", d.Type(), eventType)
	}
	if d.IsJson() != isJson {
		t.Errorf("IsJson doesn't match: %t != %t", d.IsJson(), isJson)
	}
	if bytes.Compare(d.Data(), data) != 0 {
		t.Errorf("Data doesn't match: %v != %v", d.Data(), data)
	}
	if bytes.Compare(d.Metadata(), metadata) != 0 {
		t.Errorf("Metadata doesn't match: %v != %v", d.Metadata(), metadata)
	}
}

func TestEventData_ToNewEvent(t *testing.T) {
	d := client.NewEventData(uuid.Must(uuid.NewV4()), "type", true, []byte(`{}`), []byte(`{}`))
	m := d.ToNewEvent()

	if !bytes.Equal(m.EventId, guid.ToBytes(d.EventId())) {
		t.Errorf("EventId doesn't match: %v != %v", m.EventId, d.EventId().Bytes())
	}
	if *m.EventType != d.Type() {
		t.Errorf("Type doesn't match: %s != %s", *m.EventType, d.Type())
	}
	if *m.DataContentType != 1 {
		t.Error("DataContentType should be 1 for JSON")
	}
	if *m.MetadataContentType != 0 {
		t.Error("MetadataContentType should be 0")
	}
	if bytes.Compare(m.Data, d.Data()) != 0 {
		t.Errorf("Data doesn't match: %v != %v", m.Data, d.Data())
	}
	if bytes.Compare(m.Metadata, d.Metadata()) != 0 {
		t.Errorf("Metadata doesn't match: %v != %v", m.Metadata, d.Metadata())
	}

	d = client.NewEventData(uuid.Must(uuid.NewV4()), "type", false, []byte(`{}`), []byte(`{}`))
	m = d.ToNewEvent()

	if !bytes.Equal(m.EventId, guid.ToBytes(d.EventId())) {
		t.Errorf("EventId doesn't match: %v != %v", m.EventId, d.EventId().Bytes())
	}
	if *m.EventType != d.Type() {
		t.Errorf("Type doesn't match: %s != %s", *m.EventType, d.Type())
	}
	if *m.DataContentType != 0 {
		t.Error("DataContentType should be 0")
	}
	if *m.MetadataContentType != 0 {
		t.Error("MetadataContentType should be 0")
	}
	if bytes.Compare(m.Data, d.Data()) != 0 {
		t.Errorf("Data doesn't match: %v != %v", m.Data, d.Data())
	}
	if bytes.Compare(m.Metadata, d.Metadata()) != 0 {
		t.Errorf("Metadata doesn't match: %v != %v", m.Metadata, d.Metadata())
	}
}

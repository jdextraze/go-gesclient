package client

import (
	"github.com/jdextraze/go-gesclient/guid"
	"github.com/jdextraze/go-gesclient/messages"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"time"
)

var _ = Describe("RecordedEvent", func() {
	var (
		evt                 *RecordedEvent
		streamId                  = "Test"
		id                        = uuid.Must(uuid.NewV4()).Bytes()
		number              int32 = 123
		typ                       = "Type"
		data                      = []byte{1, 2, 3}
		metadata                  = []byte{4, 5, 6}
		dataContentType     int32 = 1
		metadataContentType int32 = 0
		now                       = time.Now()
		created                   = now.UnixNano()/tick + ticksSinceEpoch
		createdEpoch              = now.Round(time.Millisecond).UnixNano() / int64(time.Millisecond)
	)

	BeforeEach(func() {
		evt = newRecordedEvent(&messages.EventRecord{
			EventStreamId:       &streamId,
			EventNumber:         &number,
			EventId:             id,
			EventType:           &typ,
			DataContentType:     &dataContentType,
			MetadataContentType: &metadataContentType,
			Data:                data,
			Metadata:            metadata,
			Created:             &created,
			CreatedEpoch:        &createdEpoch,
		})
	})

	Describe("getting event stream id", func() {
		It("should return mapped value", func() {
			Expect(evt.EventStreamId()).To(Equal(streamId))
		})
	})

	Describe("getting event id", func() {
		It("should return mapped value", func() {
			Expect(evt.EventId()).To(Equal(guid.FromBytes(id)))
		})
	})

	Describe("getting event number", func() {
		It("should return mapped value", func() {
			Expect(evt.EventNumber()).To(Equal(int(number)))
		})
	})

	Describe("getting event type", func() {
		It("should return mapped value", func() {
			Expect(evt.EventType()).To(Equal(typ))
		})
	})

	Describe("getting data", func() {
		It("should return mapped value", func() {
			Expect(evt.Data()).To(Equal(data))
		})
	})

	Describe("getting metadata", func() {
		It("should return mapped value", func() {
			Expect(evt.Metadata()).To(Equal(metadata))
		})
	})

	Describe("getting is json", func() {
		It("should return mapped value", func() {
			Expect(evt.IsJson()).To(BeTrue())
		})
	})

	Describe("getting created", func() {
		It("should return mapped value", func() {
			Expect(evt.Created().String()).To(Equal(now.Round(time.Duration(tick)).String()))
		})
	})

	Describe("getting created epoch", func() {
		It("should return mapped value", func() {
			Expect(evt.CreatedEpoch().String()).To(Equal(now.Round(time.Millisecond).String()))
		})
	})
})

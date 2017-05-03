package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StreamEventsSlice", func() {
	var streamEventsSlice *StreamEventsSlice

	BeforeEach(func() {
		streamEventsSlice = newStreamEventsSlice(
			SliceReadStatus_Success,
			"Test",
			1,
			ReadDirectionForward,
			[]*protobuf.ResolvedIndexedEvent{
				&protobuf.ResolvedIndexedEvent{},
			},
			12,
			123,
			true,
			nil,
		)
	})

	Describe("getting status", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.Status()).To(Equal(SliceReadStatus_Success))
		})
	})

	Describe("getting stream", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.Stream()).To(Equal("Test"))
		})
	})

	Describe("getting from event number", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.FromEventNumber()).To(Equal(1))
		})
	})

	Describe("getting read direction", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.ReadDirection()).To(Equal(ReadDirectionForward))
		})
	})

	Describe("getting events", func() {
		It("should return same number of events", func() {
			Expect(streamEventsSlice.Events()).To(HaveLen(1))
		})
	})

	Describe("getting next event number", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.NextEventNumber()).To(Equal(12))
		})
	})

	Describe("getting last event number", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.LastEventNumber()).To(Equal(123))
		})
	})

	Describe("getting is end of stream", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.IsEndOfStream()).To(BeTrue())
		})
	})

	Describe("getting error", func() {
		It("should return provided value", func() {
			Expect(streamEventsSlice.Error()).To(BeNil())
		})
	})
})

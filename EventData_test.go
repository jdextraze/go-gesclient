package gesclient_test

import (
	. "bitbucket.org/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("EventData", func() {
	id, _ := uuid.FromString("12345678-1234-1234-1234-1234567890AB")
	eventType := "Test"
	isJson := true
	data := []byte("{}")
	metadata := []byte{}

	Describe("getting event id", func() {
		It("should return provided value", func() {
			Expect(NewEventData(id, eventType, isJson, data, metadata).EventId()).To(Equal(id))
		})
	})

	Describe("getting type", func() {
		It("should return provided value", func() {
			Expect(NewEventData(id, eventType, isJson, data, metadata).Type()).To(Equal(eventType))
		})
	})

	Describe("getting is json", func() {
		It("should return provided value", func() {
			Expect(NewEventData(id, eventType, isJson, data, metadata).IsJson()).To(Equal(isJson))
		})
	})

	Describe("getting data", func() {
		It("should return provided value", func() {
			Expect(NewEventData(id, eventType, isJson, data, metadata).Data()).To(Equal(data))
		})
	})

	Describe("getting metadata", func() {
		It("should return provided value", func() {
			Expect(NewEventData(id, eventType, isJson, data, metadata).Metadata()).To(Equal(metadata))
		})
	})

	Describe("to new event", func() {
		Context("when is json", func() {
			It("should return mapped value", func() {
				dataContentType := int32(1)
				metaDataContentType := int32(0)
				Expect(NewEventData(id, eventType, isJson, data, metadata).ToNewEvent()).To(
					Equal(&protobuf.NewEvent{
						EventId:             id.Bytes(),
						EventType:           &eventType,
						DataContentType:     &dataContentType,
						MetadataContentType: &metaDataContentType,
						Data:                data,
						Metadata:            metadata,
					}))
			})
		})

		Context("when is not json", func() {
			It("should return mapped value", func() {
				dataContentType := int32(0)
				metaDataContentType := int32(0)
				Expect(NewEventData(id, eventType, false, data, metadata).ToNewEvent()).To(
					Equal(&protobuf.NewEvent{
						EventId:             id.Bytes(),
						EventType:           &eventType,
						DataContentType:     &dataContentType,
						MetadataContentType: &metaDataContentType,
						Data:                data,
						Metadata:            metadata,
					}))
			})
		})
	})
})

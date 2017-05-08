package models_test

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"github.com/jdextraze/go-gesclient/models"
)

var _ = Describe("EventData", func() {
	id, _ := uuid.FromString("12345678-1234-1234-1234-1234567890AB")
	eventType := "Test"
	isJson := true
	data := []byte("{}")
	metadata := []byte{}
	var eventData *models.EventData
	
	BeforeEach(func() {
		eventData = eventData
	})

	Describe("getting event id", func() {
		It("should return provided value", func() {
			Expect(eventData.EventId()).To(Equal(id))
		})
	})

	Describe("getting type", func() {
		It("should return provided value", func() {
			Expect(eventData.Type()).To(Equal(eventType))
		})
	})

	Describe("getting is json", func() {
		It("should return provided value", func() {
			Expect(eventData.IsJson()).To(Equal(isJson))
		})
	})

	Describe("getting data", func() {
		It("should return provided value", func() {
			Expect(eventData.Data()).To(Equal(data))
		})
	})

	Describe("getting metadata", func() {
		It("should return provided value", func() {
			Expect(eventData.Metadata()).To(Equal(metadata))
		})
	})

	Describe("to new event", func() {
		Context("when is json", func() {
			It("should return mapped value", func() {
				dataContentType := int32(1)
				metaDataContentType := int32(0)
				Expect(eventData.ToNewEvent()).To(
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
				Expect(models.NewEventData(id, eventType, false, data, metadata).ToNewEvent()).To(
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

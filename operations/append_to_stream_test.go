package operations

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppendToStreamOperation", func() {
	var (
		sut    *appendToStream
		result chan *client.WriteResult
	)

	BeforeEach(func() {
		result = make(chan *client.WriteResult, 1)
		sut = NewAppendToStream(
			"Test",
			[]*client.EventData{},
			client.ExpectedVersion_Any,
			client.NewUserCredentials("test", "!test!"),
			result,
		)
	})

	Describe("getting request command", func() {
		It("should return WriteEvents", func() {
			Expect(sut.GetRequestCommand()).To(BeEquivalentTo(client.Command_WriteEvents))
		})
	})

	Describe("getting request message", func() {
		var msg proto.Message

		BeforeEach(func() {
			msg = sut.GetRequestMessage()
		})

		It("should return a WriteEvents message", func() {
			Expect(msg).To(BeAssignableToTypeOf(&protobuf.WriteEvents{}))
		})

		It("should be populated", func() {
			writeEvents := msg.(*protobuf.WriteEvents)
			Expect(writeEvents.GetEventStreamId()).To(Equal("Test"))
			Expect(writeEvents.GetEvents()).To(BeEquivalentTo([]*protobuf.NewEvent{}))
			Expect(writeEvents.GetExpectedVersion()).To(BeEquivalentTo(client.ExpectedVersion_Any))
			Expect(writeEvents.GetRequireMaster()).To(BeFalse())
		})
	})

	Describe("getting user credentials", func() {
		It("should return provided value", func() {
			Expect(sut.UserCredentials()).To(Equal(client.NewUserCredentials("test", "!test!")))
		})
	})

	Describe("fail", func() {
		var res <-chan *client.WriteResult

		BeforeEach(func() {
			res = sut.resultChannel
			sut.Fail(errors.New("error"))
		})

		It("should send writeresult on the channel and close it", func() {
			Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, errors.New("error")))))
			Expect(res).To(BeClosed())
		})

		It("should be completed", func() {
			Expect(sut.IsCompleted()).To(BeTrue())
		})

		It("should not retry", func() {
			Expect(sut.Retry()).To(BeFalse())
		})
	})

	Describe("parsing response", func() {
		var res <-chan *client.WriteResult

		Context("when response is success", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_Success.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should send writeresult on the channel and close it", func() {
				pos, _ := client.NewPosition(0, 0)
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, pos, nil))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is prepare timeout", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_PrepareTimeout.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should not send writeresult on the channel", func() {
				Expect(res).ToNot(Receive())
			})

			It("should not close the channel", func() {
				Expect(res).ToNot(BeClosed())
			})

			It("should not be completed", func() {
				Expect(sut.IsCompleted()).To(BeFalse())
			})

			It("should set retry", func() {
				Expect(sut.Retry()).To(BeTrue())
			})
		})

		Context("when response is forward timeout", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_PrepareTimeout.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should not send writeresult on the channel", func() {
				Expect(res).ToNot(Receive())
			})

			It("should not close the channel", func() {
				Expect(res).ToNot(BeClosed())
			})

			It("should not be completed", func() {
				Expect(sut.IsCompleted()).To(BeFalse())
			})

			It("should set retry", func() {
				Expect(sut.Retry()).To(BeTrue())
			})
		})

		Context("when response is commit timeout", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_PrepareTimeout.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should not send writeresult on the channel", func() {
				Expect(res).ToNot(Receive())
			})

			It("should not close the channel", func() {
				Expect(res).ToNot(BeClosed())
			})

			It("should not be completed", func() {
				Expect(sut.IsCompleted()).To(BeFalse())
			})

			It("should set retry", func() {
				Expect(sut.Retry()).To(BeTrue())
			})
		})

		Context("when response is wrong expected version", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_WrongExpectedVersion.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, client.WrongExpectedVersion))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is stream deleted", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_StreamDeleted.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, client.StreamDeleted))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is invalid transaction", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_InvalidTransaction.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, client.InvalidTransaction))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is access denied", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero32 := int32(0)
				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.WriteEventsCompleted{
					Result:           protobuf.OperationResult_AccessDenied.Enum(),
					CommitPosition:   &zero64,
					PreparePosition:  &zero64,
					LastEventNumber:  &zero32,
					FirstEventNumber: &zero32,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    payload,
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, client.AccessDenied))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response protobuf unmarshal fails", func() {
			BeforeEach(func() {
				res = sut.resultChannel
				sut.ParseResponse(&client.Package{
					Command: client.Command_WriteEventsCompleted,
					Data:    []byte{0},
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil,
					errors.New("proto: protobuf.WriteEventsCompleted: illegal tag 0 (wire type 0)")))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is not authenticated", func() {
			BeforeEach(func() {
				res = sut.resultChannel
				sut.ParseResponse(&client.Package{
					Command: client.Command_NotAuthenticated,
					Data:    []byte{},
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, client.AuthenticationError))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is bad request", func() {
			BeforeEach(func() {
				res = sut.resultChannel
				sut.ParseResponse(&client.Package{
					Command: client.Command_BadRequest,
					Data:    []byte{},
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, client.BadRequest))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is not handled - not ready", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				payload, _ := proto.Marshal(&protobuf.NotHandled{
					Reason: protobuf.NotHandled_NotReady.Enum(),
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_NotHandled,
					Data:    payload,
				})
			})

			It("should not send writeresult on the channel", func() {
				Expect(res).ToNot(Receive())
			})

			It("should not close the channel", func() {
				Expect(res).ToNot(BeClosed())
			})

			It("should not be completed", func() {
				Expect(sut.IsCompleted()).To(BeFalse())
			})

			It("should retry", func() {
				Expect(sut.Retry()).To(BeTrue())
			})
		})

		Context("when response is not handled - too busy", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				payload, _ := proto.Marshal(&protobuf.NotHandled{
					Reason: protobuf.NotHandled_TooBusy.Enum(),
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_NotHandled,
					Data:    payload,
				})
			})

			It("should not send writeresult on the channel", func() {
				Expect(res).ToNot(Receive())
			})

			It("should not close the channel", func() {
				Expect(res).ToNot(BeClosed())
			})

			It("should not be completed", func() {
				Expect(sut.IsCompleted()).To(BeFalse())
			})

			It("should retry", func() {
				Expect(sut.Retry()).To(BeTrue())
			})
		})

		Context("when response is not handled - not master", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				payload, _ := proto.Marshal(&protobuf.NotHandled{
					Reason: protobuf.NotHandled_NotMaster.Enum(),
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_NotHandled,
					Data:    payload,
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil,
					errors.New("NotHandled - NotMaster not supported")))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})

		Context("when response is not handled - unknown", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				reason := protobuf.NotHandled_NotHandledReason(-1)
				payload, _ := proto.Marshal(&protobuf.NotHandled{
					Reason: &reason,
				})

				sut.ParseResponse(&client.Package{
					Command: client.Command_NotHandled,
					Data:    payload,
				})
			})

			It("should not send writeresult on the channel", func() {
				Expect(res).ToNot(Receive())
			})

			It("should not close the channel", func() {
				Expect(res).ToNot(BeClosed())
			})

			It("should not be completed", func() {
				Expect(sut.IsCompleted()).To(BeFalse())
			})

			It("should retry", func() {
				Expect(sut.Retry()).To(BeTrue())
			})
		})

		Context("when response is unexpected command", func() {
			BeforeEach(func() {
				res = sut.resultChannel
				sut.ParseResponse(&client.Package{
					Command: client.Command_SubscriptionConfirmation,
					Data:    []byte{},
				})
			})

			It("should send writeresult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(client.NewWriteResult(0, nil, errors.New(
					"Command not expected. Expected: WriteEventsCompleted, Actual: SubscriptionConfirmation")))))
				Expect(res).To(BeClosed())
			})

			It("should be completed", func() {
				Expect(sut.IsCompleted()).To(BeTrue())
			})

			It("should not retry", func() {
				Expect(sut.Retry()).To(BeFalse())
			})
		})
	})
})

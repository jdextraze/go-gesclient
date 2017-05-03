package gesclient

import (
	"github.com/jdextraze/go-gesclient/protobuf"
	"errors"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteStreamOperation", func() {
	var (
		sut *deleteStreamOperation
	)

	BeforeEach(func() {
		sut = newDeleteStreamOperation(
			"Test",
			ExpectedVersion_Any,
			false,
			NewUserCredentials("test", "!test!"),
		)
	})

	Describe("getting request command", func() {
		It("should return DeleteStream", func() {
			Expect(sut.GetRequestCommand()).To(BeEquivalentTo(tcpCommand_DeleteStream))
		})
	})

	Describe("getting request message", func() {
		var msg proto.Message

		BeforeEach(func() {
			msg = sut.GetRequestMessage()
		})

		It("should return a DeleteStream message", func() {
			Expect(msg).To(BeAssignableToTypeOf(&protobuf.DeleteStream{}))
		})

		It("should be populated", func() {
			deleteStream := msg.(*protobuf.DeleteStream)
			Expect(deleteStream.GetEventStreamId()).To(Equal("Test"))
			Expect(deleteStream.GetExpectedVersion()).To(BeEquivalentTo(ExpectedVersion_Any))
			Expect(deleteStream.GetHardDelete()).To(BeFalse())
			Expect(deleteStream.GetRequireMaster()).To(BeFalse())
		})
	})

	Describe("getting user credentials", func() {
		It("should return provided value", func() {
			Expect(sut.UserCredentials()).To(Equal(NewUserCredentials("test", "!test!")))
		})
	})

	Describe("fail", func() {
		var res <-chan *DeleteResult

		BeforeEach(func() {
			res = sut.resultChannel
			sut.Fail(errors.New("error"))
		})

		It("should send DeleteResult on the channel and close it", func() {
			Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, errors.New("error")))))
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
		var res <-chan *DeleteResult

		Context("when response is success", func() {
			BeforeEach(func() {
				res = sut.resultChannel

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_Success.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				pos, _ := NewPosition(0, 0)
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(pos, nil))))
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

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_PrepareTimeout.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should not send DeleteResult on the channel", func() {
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

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_PrepareTimeout.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should not send DeleteResult on the channel", func() {
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

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_PrepareTimeout.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should not send DeleteResult on the channel", func() {
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

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_WrongExpectedVersion.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, WrongExpectedVersion))))
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

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_StreamDeleted.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, StreamDeleted))))
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

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_InvalidTransaction.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, InvalidTransaction))))
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

				zero64 := int64(0)
				payload, _ := proto.Marshal(&protobuf.DeleteStreamCompleted{
					Result:          protobuf.OperationResult_AccessDenied.Enum(),
					CommitPosition:  &zero64,
					PreparePosition: &zero64,
				})

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: payload,
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, AccessDenied))))
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
				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_DeleteStreamCompleted,
					Payload: []byte{0},
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil,
					errors.New("proto: protobuf.DeleteStreamCompleted: illegal tag 0 (wire type 0)")))))
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
				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_NotAuthenticated,
					Payload: []byte{},
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, AuthenticationError))))
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
				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_BadRequest,
					Payload: []byte{},
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, BadRequest))))
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

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_NotHandled,
					Payload: payload,
				})
			})

			It("should not send DeleteResult on the channel", func() {
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

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_NotHandled,
					Payload: payload,
				})
			})

			It("should not send DeleteResult on the channel", func() {
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

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_NotHandled,
					Payload: payload,
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil,
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

				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_NotHandled,
					Payload: payload,
				})
			})

			It("should not send DeleteResult on the channel", func() {
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
				sut.ParseResponse(&tcpPacket{
					Command: tcpCommand_SubscriptionConfirmation,
					Payload: []byte{},
				})
			})

			It("should send DeleteResult on the channel and close it", func() {
				Expect(res).To(Receive(BeEquivalentTo(NewDeleteResult(nil, errors.New(
					"Command not expected. Expected: DeleteStreamCompleted, Actual: SubscriptionConfirmation")))))
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

package gesclient

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("tcpPacket", func() {
	correlationId, _ := uuid.FromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	payload := []byte{1, 2, 3, 4, 5}

	Context("when no flags are set", func() {
		flags := tcpFlagsNone

		Describe("getting bytes", func() {
			It("should return byte representation", func() {
				Expect(newTcpPacket(
					tcpCommand_BadRequest,
					flags,
					correlationId,
					payload,
					nil,
				).Bytes()).To(Equal([]byte{
					tcpCommand_BadRequest,
					flags,
					1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
					1, 2, 3, 4, 5,
				}))
			})
		})

		Describe("getting size", func() {
			It("should return length", func() {
				Expect(newTcpPacket(
					tcpCommand_BadRequest,
					flags,
					correlationId,
					payload,
					nil,
				).Size()).To(Equal(int32(1 + 1 + 16 + 5)))
			})
		})
	})

	Context("when authenticated flags is set", func() {
		flags := tcpFlagsAuthenticated

		Describe("getting bytes", func() {
			It("should return byte representation", func() {
				Expect(newTcpPacket(
					tcpCommand_BadRequest,
					flags,
					correlationId,
					payload,
					NewUserCredentials("test", "!test!"),
				).Bytes()).To(Equal([]byte{
					tcpCommand_BadRequest,
					flags,
					1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
					4, 116, 101, 115, 116, 6, 33, 116, 101, 115, 116, 33,
					1, 2, 3, 4, 5,
				}))
			})
		})

		Describe("getting size", func() {
			It("should return length", func() {
				Expect(newTcpPacket(
					tcpCommand_BadRequest,
					flags,
					correlationId,
					payload,
					NewUserCredentials("test", "!test!"),
				).Size()).To(Equal(int32(1 + 1 + 16 + 5 + 7 + 5)))
			})
		})
	})
})

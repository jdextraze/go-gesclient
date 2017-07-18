package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Package", func() {
	correlationId, _ := uuid.FromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	payload := []byte{1, 2, 3, 4, 5}

	Context("when no flags are set", func() {
		flags := byte(0)

		Describe("getting bytes", func() {
			It("should return byte representation", func() {
				Expect(client.NewTcpPackage(
					client.Command_BadRequest,
					flags,
					correlationId,
					payload,
					nil,
				).Bytes()).To(Equal([]byte{
					client.Command_BadRequest,
					flags,
					1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
					1, 2, 3, 4, 5,
				}))
			})
		})

		Describe("getting size", func() {
			It("should return length", func() {
				Expect(client.NewTcpPackage(
					client.Command_BadRequest,
					flags,
					correlationId,
					payload,
					nil,
				).Size()).To(Equal(int32(1 + 1 + 16 + 5)))
			})
		})

		Describe("build from bytes", func() {
			It("should return TcpPackage", func() {
				tcpPackage, _ := client.TcpPacketFromBytes([]byte{
					client.Command_BadRequest,
					flags,
					1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
					1, 2, 3, 4, 5,
				})
				Expect(tcpPackage.Command()).To(Equal(client.Command_BadRequest))
				Expect(tcpPackage.Flags()).To(Equal(flags))
			})
		})
	})

	Context("when authenticated flags is set", func() {
		flags := byte(1)

		Describe("getting bytes", func() {
			It("should return byte representation", func() {
				Expect(client.NewTcpPackage(
					client.Command_BadRequest,
					flags,
					correlationId,
					payload,
					client.NewUserCredentials("test", "!test!"),
				).Bytes()).To(Equal([]byte{
					client.Command_BadRequest,
					flags,
					1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
					4, 116, 101, 115, 116, 6, 33, 116, 101, 115, 116, 33,
					1, 2, 3, 4, 5,
				}))
			})
		})

		Describe("getting size", func() {
			It("should return length", func() {
				Expect(client.NewTcpPackage(
					client.Command_BadRequest,
					flags,
					correlationId,
					payload,
					client.NewUserCredentials("test", "!test!"),
				).Size()).To(Equal(int32(1 + 1 + 16 + 5 + 7 + 5)))
			})
		})
	})
})

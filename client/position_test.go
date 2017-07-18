package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Position", func() {
	Describe("new instance", func() {
		Context("when commit is less than prepare", func() {
			It("should error", func() {
				Expect(func() { client.NewPosition(10, 11) }).To(Panic())
			})
		})

		Context("when commit is not less than prepare", func() {
			It("should not error", func() {
				Expect(func() { client.NewPosition(10, 10) }).To(Panic())
			})
		})
	})

	Describe("getting commit position", func() {
		It("should return provided value", func() {
			Expect(client.NewPosition(10, 9).CommitPosition()).To(Equal(int64(10)))
		})
	})

	Describe("getting prepare position", func() {
		It("should return provided value", func() {
			Expect(client.NewPosition(10, 9).PreparePosition()).To(Equal(int64(9)))
		})
	})

	Describe("getting string", func() {
		It("should return string representation", func() {
			Expect(client.NewPosition(10, 10).String()).To(Equal("10/10"))
		})
	})
})

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
				pos, err := client.NewPosition(10, 11)
				Expect(pos).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when commit is not less than prepare", func() {
			It("should not error", func() {
				pos, err := client.NewPosition(10, 10)
				Expect(pos).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("getting commit position", func() {
		It("should return provided value", func() {
			pos, _ := client.NewPosition(10, 9)
			Expect(pos.CommitPosition()).To(Equal(int64(10)))
		})
	})

	Describe("getting prepare position", func() {
		It("should return provided value", func() {
			pos, _ := client.NewPosition(10, 9)
			Expect(pos.PreparePosition()).To(Equal(int64(9)))
		})
	})

	Describe("getting string", func() {
		It("should return string representation", func() {
			pos, _ := client.NewPosition(10, 10)
			Expect(pos.String()).To(Equal("10/10"))
		})
	})
})

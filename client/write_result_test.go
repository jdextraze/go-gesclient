package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WriteResult", func() {
	Describe("getting next expected version", func() {
		It("should return provided value", func() {
			Expect(client.NewWriteResult(123, nil).NextExpectedVersion()).To(Equal(123))
		})
	})

	Describe("getting log position", func() {
		It("should return provided value", func() {
			Expect(client.NewWriteResult(0, client.Position_Start).LogPosition()).To(Equal(client.Position_Start))
		})
	})
})

package client_test

import (
	"github.com/jdextraze/go-gesclient/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserCredentials", func() {
	Describe("getting username", func() {
		It("should return provided value", func() {
			Expect(client.NewUserCredentials("test", "!test!").Username()).To(Equal("test"))
		})
	})

	Describe("getting password", func() {
		It("should return provided value", func() {
			Expect(client.NewUserCredentials("test", "!test!").Password()).To(Equal("!test!"))
		})
	})
})

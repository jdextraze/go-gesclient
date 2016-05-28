package gesclient_test

import (
	. "bitbucket.org/jdextraze/go-gesclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserCredentials", func() {
	Describe("getting username", func() {
		It("should return provided value", func() {
			Expect(NewUserCredentials("test", "!test!").Username()).To(Equal("test"))
		})
	})

	Describe("getting password", func() {
		It("should return provided value", func() {
			Expect(NewUserCredentials("test", "!test!").Password()).To(Equal("!test!"))
		})
	})
})

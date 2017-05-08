package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jdextraze/go-gesclient/models"
)

var _ = Describe("UserCredentials", func() {
	Describe("getting username", func() {
		It("should return provided value", func() {
			Expect(models.NewUserCredentials("test", "!test!").Username()).To(Equal("test"))
		})
	})

	Describe("getting password", func() {
		It("should return provided value", func() {
			Expect(models.NewUserCredentials("test", "!test!").Password()).To(Equal("!test!"))
		})
	})
})

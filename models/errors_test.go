package models_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jdextraze/go-gesclient/models"
)

var _ = Describe("errors", func() {
	Describe("Getting error message of WrongExpectedVersion", func() {
		It("should return Wrong expected version", func() {
			Expect(models.WrongExpectedVersion.Error()).Should(Equal("Wrong expected version"))
		})
	})

	Describe("Getting error message of StreamDeleted", func() {
		It("should return Stream deleted", func() {
			Expect(models.StreamDeleted.Error()).Should(Equal("Stream deleted"))
		})
	})

	Describe("Getting error message of InvalidTransaction", func() {
		It("should return Invalid transaction", func() {
			Expect(models.InvalidTransaction.Error()).Should(Equal("Invalid transaction"))
		})
	})

	Describe("Getting error message of AccessDenied", func() {
		It("should return Access denied", func() {
			Expect(models.AccessDenied.Error()).Should(Equal("Access denied"))
		})
	})

	Describe("Getting error message of AuthenticationError", func() {
		It("should return Authentication error", func() {
			Expect(models.AuthenticationError.Error()).Should(Equal("Authentication error"))
		})
	})

	Describe("Getting error message of BadRequest", func() {
		It("should return Bad request", func() {
			Expect(models.BadRequest.Error()).Should(Equal("Bad request"))
		})
	})

	Describe("Getting error message of ServerError", func() {
		Context("when message is empty", func() {
			It("should return error message", func() {
				Expect(models.NewServerError("").Error()).To(Equal("Unexpected error on server: <no message>"))
			})
		})

		Context("when message it not empty", func() {
			It("should return error message", func() {
				Expect(models.NewServerError("message").Error()).To(Equal("Unexpected error on server: message"))
			})
		})
	})

	Describe("Getting error message of NotModified", func() {
		It("should return error message", func() {
			Expect(models.NewNotModified("test").Error()).Should(Equal("Stream not modified: test"))
		})
	})
})

package models_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jdextraze/go-gesclient/models"
)

var _ = Describe("WriteResult", func() {
	Describe("getting next expected version", func() {
		It("should return provided value", func() {
			Expect(models.NewWriteResult(123, nil, nil).NextExpectedVersion()).To(Equal(123))
		})
	})

	Describe("getting log position", func() {
		It("should return provided value", func() {
			pos, _ := models.NewPosition(0, 0)
			Expect(models.NewWriteResult(0, pos, nil).LogPosition()).To(Equal(pos))
		})
	})

	Describe("getting error", func() {
		It("should return provided value", func() {
			err := errors.New("error")
			Expect(models.NewWriteResult(0, nil, err).Error()).To(Equal(err))
		})
	})
})

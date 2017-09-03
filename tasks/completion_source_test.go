package tasks_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jdextraze/go-gesclient/tasks"
	"errors"
)

var completionSourceError = errors.New("error")

var _ = Describe("CompletionSource", func() {
	var completionSource *tasks.CompletionSource

	BeforeEach(func() {
		completionSource = tasks.NewCompletionSource()
	})

	Describe("SetError", func() {
		It("should set error on task", func() {
			completionSource.SetError(completionSourceError)
			Expect(completionSource.Task().Wait()).To(MatchError(completionSourceError))
		})

		It("should panic if already set", func() {
			completionSource.SetError(completionSourceError)
			Expect(completionSource.SetError(completionSourceError)).To(MatchError(tasks.AlreadyCompleted))
		})
	})

	Describe("TrySetError", func() {
		It("should set error on task", func() {
			Expect(completionSource.TrySetError(completionSourceError)).To(BeTrue())
			Expect(completionSource.Task().Wait()).To(MatchError(completionSourceError))
		})

		It("should not set error on task if already set", func() {
			completionSource.TrySetError(completionSourceError)
			Expect(completionSource.TrySetError(errors.New("other error"))).To(BeFalse())
			Expect(completionSource.Task().Wait()).To(MatchError(completionSourceError))
		})
	})
})

package tasks_test

import (
	"errors"
	"github.com/jdextraze/go-gesclient/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
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

		It("should return error if already set", func() {
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

	Describe("SetResult", func() {
		It("should set result on task", func() {
			completionSource.SetResult(true)
			Expect(completionSource.Task().Result()).To(BeTrue())
		})

		It("should return error if already set", func() {
			completionSource.SetResult(true)
			Expect(completionSource.SetResult(true)).To(MatchError(tasks.AlreadyCompleted))
		})
	})

	Describe("TrySetResult", func() {
		It("should set result on task", func() {
			Expect(completionSource.TrySetResult(true)).To(BeTrue())
			Expect(completionSource.Task().Result()).To(BeTrue())
		})

		It("should not set result on task if already set", func() {
			completionSource.TrySetResult(true)
			Expect(completionSource.TrySetResult(false)).To(BeFalse())
			Expect(completionSource.Task().Result()).To(BeTrue())
		})
	})

	Describe("Task", func() {
		It("should return a task", func() {
			Expect(completionSource.Task()).To(BeAssignableToTypeOf(&tasks.Task{}))
		})

		It("should block task until result is set", func() {
			go func() {
				time.Sleep(100 * time.Millisecond)
				completionSource.SetResult(true)
			}()
			start := time.Now()
			completionSource.Task().Wait()
			Expect(time.Now().Sub(start) >= 100*time.Millisecond).To(BeTrue())
		})

		It("should block task until error is set", func() {
			go func() {
				time.Sleep(100 * time.Millisecond)
				completionSource.SetError(completionSourceError)
			}()
			start := time.Now()
			completionSource.Task().Wait()
			Expect(time.Now().Sub(start) >= 100*time.Millisecond).To(BeTrue())
		})
	})
})

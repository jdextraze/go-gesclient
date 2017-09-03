package tasks_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jdextraze/go-gesclient/tasks"
	"time"
	"errors"
	"sync"
	"sync/atomic"
)

var taskError = errors.New("error")

var _ = Describe("Task", func() {
	var task *tasks.Task

	Describe("Result", func() {
		It("should wait for task to complete", func() {
			task = tasks.New(func() (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return true, nil
			})
			start := time.Now()
			task.Result()
			Expect(time.Now().Sub(start) >= 100 * time.Millisecond).To(BeTrue())
		})

		It("should return result from callback", func() {
			task = tasks.New(func() (interface{}, error) {
				return true, nil
			})
			Expect(task.Result()).To(BeTrue())
		})
	})

	Describe("Error", func() {
		It("should wait for task to complete", func() {
			task = tasks.New(func() (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return nil, nil
			})
			start := time.Now()
			task.Error()
			Expect(time.Now().Sub(start) >= 100 * time.Millisecond).To(BeTrue())
		})

		It("should return error from callback", func() {
			task = tasks.New(func() (interface{}, error) {
				return nil, taskError
			})
			Expect(task.Error()).To(MatchError(taskError))
		})
	})

	Describe("IsCompleted", func() {
		It("should return false before completion", func() {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			task = tasks.NewStarted(func() (interface{}, error) {
				wg.Wait()
				return nil, nil
			})
			Expect(task.IsCompleted()).To(BeFalse())
			wg.Done()
		})

		It("should return true after completion", func() {
			task = tasks.NewStarted(func() (interface{}, error) {
				return nil, nil
			})
			task.Wait()
			Expect(task.IsCompleted()).To(BeTrue())
		})
	})

	Describe("IsFaulted", func() {
		It("should return false before completion", func() {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			task = tasks.NewStarted(func() (interface{}, error) {
				wg.Wait()
				return nil, nil
			})
			Expect(task.IsFaulted()).To(BeFalse())
			wg.Done()
		})

		It("should return false after completion without error", func() {
			task = tasks.NewStarted(func() (interface{}, error) {
				return nil, nil
			})
			task.Wait()
			Expect(task.IsFaulted()).To(BeFalse())
		})

		It("should return true after completion with error", func() {
			task = tasks.NewStarted(func() (interface{}, error) {
				return nil, taskError
			})
			task.Wait()
			Expect(task.IsFaulted()).To(BeTrue())
		})
	})

	Describe("ContinueWith", func() {
		It("should start source task", func() {
			started := int32(0)
			tasks.New(func() (interface{}, error) {
				atomic.AddInt32(&started, 1)
				return nil, nil
			}).ContinueWith(func(t *tasks.Task) (interface{}, error) {
				atomic.AddInt32(&started, 1)
				return nil, nil
			})
			time.Sleep(10 * time.Millisecond)
			Expect(started).To(Equal(int32(2)))
		})

		It("should receive original task", func() {
			task = tasks.New(func() (interface{}, error) {
				return true, nil
			}).ContinueWith(func(t *tasks.Task) (interface{}, error) {
				Expect(t.Result()).To(BeTrue())
				return nil, nil
			})
			task.Wait()
		})
	})

	Describe("Start", func() {
		It("should start task", func() {
			started := int32(0)
			task = tasks.New(func() (interface{}, error) {
				atomic.StoreInt32(&started, 1)
				return nil, nil
			})
			task.Start()
			time.Sleep(10 * time.Millisecond)
			Expect(started).To(Equal(int32(1)))
		})

		It("should return error if already started", func() {
			task = tasks.New(func() (interface{}, error) {
				return nil, nil
			})
			task.Start()
			Expect(task.Start()).To(MatchError(tasks.AlreadyRunning))
		})
	})

	Describe("Wait", func() {
		It("should wait for task to complete", func() {
			task = tasks.New(func() (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return nil, nil
			})
			start := time.Now()
			task.Wait()
			Expect(time.Now().Sub(start) >= 100 * time.Millisecond).To(BeTrue())
		})

		It("should return callback error", func() {
			task = tasks.New(func() (interface{}, error) {
				return nil, taskError
			})
			Expect(task.Wait()).To(MatchError(taskError))
		})

		It("should return immediately if already completed", func() {
			task = tasks.New(func() (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return nil, nil
			})
			start := time.Now()
			task.Wait()
			task.Wait()
			Expect(time.Now().Sub(start) >= 100 * time.Millisecond).To(BeTrue())
		})
	})
})

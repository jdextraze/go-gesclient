package tasks

import (
	"errors"
	"sync"
	"sync/atomic"
)

var AlreadyCompleted = errors.New("Already completed")

type CompletionSource struct {
	waitGroup *sync.WaitGroup
	task      *Task
	completed int32
	result    interface{}
	err       error
}

func NewCompletionSource() *CompletionSource {
	obj := &CompletionSource{
		waitGroup: &sync.WaitGroup{},
	}
	obj.waitGroup.Add(1)
	obj.task = NewStarted(obj.run)
	return obj
}

func (s *CompletionSource) run() (interface{}, error) {
	s.waitGroup.Wait()
	return s.result, s.err
}

func (s *CompletionSource) SetError(err error) error {
	if atomic.CompareAndSwapInt32(&s.completed, 0, 1) {
		s.err = err
		s.waitGroup.Done()
		return nil
	}
	return AlreadyCompleted
}

func (s *CompletionSource) TrySetError(err error) bool {
	return s.SetError(err) == nil
}

func (s *CompletionSource) SetResult(result interface{}) error {
	if atomic.CompareAndSwapInt32(&s.completed, 0, 1) {
		s.result = result
		s.waitGroup.Done()
		return nil
	}
	return AlreadyCompleted
}

func (s *CompletionSource) TrySetResult(result interface{}) bool {
	return s.SetResult(result) == nil
}

func (s *CompletionSource) Task() *Task {
	return s.task
}

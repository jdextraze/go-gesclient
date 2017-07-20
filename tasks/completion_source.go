package tasks

import (
	"sync"
)

type CompletionSource struct {
	waitGroup *sync.WaitGroup
	task      *Task
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

func (s *CompletionSource) SetError(err error) {
	s.err = err
	s.waitGroup.Done()
}

func (s *CompletionSource) TrySetError(err error) {
	defer func() {
		recover()
	}()
	s.SetError(err)
}

func (s *CompletionSource) SetResult(result interface{}) {
	s.result = result
	s.waitGroup.Done()
}

func (s *CompletionSource) TrySetResult(result interface{}) {
	defer func() {
		recover()
	}()
	s.SetResult(result)
}

func (s *CompletionSource) Task() *Task {
	return s.task
}

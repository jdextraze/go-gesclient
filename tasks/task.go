package tasks

import (
	"errors"
	"sync"
	"sync/atomic"
)

type TaskCallback func() (interface{}, error)
type ContinueWithCallback func(*Task) (interface{}, error)

type Task struct {
	fn        TaskCallback
	result    interface{}
	err       error
	started   int32
	completed int32
	waitGroup *sync.WaitGroup
}

func New(cb TaskCallback) *Task {
	return &Task{
		fn:        cb,
		waitGroup: &sync.WaitGroup{},
	}
}

func NewStarted(cb TaskCallback) *Task {
	t := New(cb)
	t.Start()
	return t
}

func (t *Task) Result() interface{} {
	t.Wait()
	return t.result
}

func (t *Task) Error() error {
	return t.Wait()
}

func (t *Task) IsCompleted() bool {
	return atomic.LoadInt32(&t.completed) == 1
}

func (t *Task) IsFaulted() bool {
	return t.err != nil
}

func (t *Task) ContinueWith(cb ContinueWithCallback) *Task {
	return NewStarted(func() (interface{}, error) {
		t.Wait()
		return cb(t)
	})
}

func (t *Task) Start() error {
	if atomic.CompareAndSwapInt32(&t.started, 0, 1) {
		t.waitGroup.Add(1)
		go func() {
			t.result, t.err = t.fn()
			atomic.StoreInt32(&t.completed, 1)
			t.waitGroup.Done()
		}()
		return nil
	}
	return errors.New("Already running")
}

func (t *Task) Wait() error {
	t.Start()
	t.waitGroup.Wait()
	return t.err
}

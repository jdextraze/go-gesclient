package tasks

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
)

type TaskCallback func() (interface{}, error)
type ContinueWithCallback func(*Task) (interface{}, error)

type Task struct {
	fn        TaskCallback
	result    interface{}
	err       error
	running   int32
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

func (t *Task) Result(res interface{}) error {
	t.Start()
	t.Wait()
	result := reflect.ValueOf(t.result)
	if result.IsValid() && !result.IsNil() {
		resValue := reflect.ValueOf(res).Elem()
		resultValue := result.Elem()
		firstField := resultValue.Type().Field(0)
		if firstField.Anonymous && firstField.Type.AssignableTo(reflect.TypeOf(res)) {
			resValue.Set(resultValue.Field(0).Elem())
		} else {
			resValue.Set(resultValue)
		}
	}
	return t.err
}

func (t *Task) Error() error {
	return t.err
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
	if atomic.CompareAndSwapInt32(&t.running, 0, 1) {
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
	t.waitGroup.Wait()
	return t.err
}

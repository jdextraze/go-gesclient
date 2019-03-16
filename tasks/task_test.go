package tasks_test

import (
	"errors"
	"github.com/jdextraze/go-gesclient/tasks"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var task *tasks.Task
	task = tasks.New(func() (interface{}, error) { return nil, nil })
	if task == nil {
		t.Fail()
	}

	defer func() {
		recover()
	}()
	task = tasks.New(nil)
	t.Fail()
}

func TestNewStarted(t *testing.T) {
	var task *tasks.Task
	task = tasks.NewStarted(func() (interface{}, error) { return nil, nil })
	if task == nil {
		t.FailNow()
	}
	if task.Start() != tasks.AlreadyRunning {
		t.Fail()
	}

	defer func() {
		recover()
	}()
	task = tasks.NewStarted(nil)
	t.Fail()
}

func TestTask_ContinueWith(t *testing.T) {
	task := tasks.New(func() (interface{}, error) {
		return 1, nil
	}).ContinueWith(func(t *tasks.Task) (interface{}, error) {
		return 2, nil
	})
	if task == nil {
		t.FailNow()
	}
	if task.Result() != 2 {
		t.Fail()
	}

	task = tasks.New(func() (interface{}, error) {
		return nil, errors.New(":(")
	}).ContinueWith(func(t *tasks.Task) (interface{}, error) {
		return 2, nil
	})
	if task == nil {
		t.FailNow()
	}
	if task.Result() != 2 {
		t.Fail()
	}
}

func TestTask_Error(t *testing.T) {
	if tasks.New(func () (interface{}, error) { return nil, errors.New(":(") }).Error().Error() != ":(" {
		t.Fail()
	}
}

func TestTask_IsCompleted(t *testing.T) {
	task := tasks.NewStarted(func() (interface{}, error) { return nil, nil })
	if task.IsCompleted() {
		t.Fail()
	}
	time.Sleep(10 * time.Millisecond)
	if !task.IsCompleted() {
		t.Fail()
	}
}

func TestTask_IsFaulted(t *testing.T) {
	task := tasks.NewStarted(func() (interface{}, error) { return nil, errors.New(":(") })
	time.Sleep(10 * time.Millisecond)
	if !task.IsFaulted() {
		t.Fail()
	}
}

func TestTask_Result(t *testing.T) {
	if tasks.New(func() (interface{}, error) { return 1, nil }).Result() != 1 {
		t.Fail()
	}
}

func TestTask_Start(t *testing.T) {
	task := tasks.New(func() (interface{}, error) { return nil, nil })
	if task.Start() != nil {
		t.Fail()
	}
	if task.Start() == nil {
		t.Fail()
	}
}

func TestTask_Wait(t *testing.T) {
	task := tasks.New(func() (interface{}, error) { return nil, nil })
	if task.Wait() != nil {
		t.Fail()
	}
	if task.Wait() != nil {
		t.Fail()
	}
}

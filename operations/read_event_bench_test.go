package operations_test

import (
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"github.com/satori/go.uuid"
	"testing"
)

func init() {
	ensureConnection()
	t, err := es.AppendToStreamAsync("BenchmarkReadEvent", client.ExpectedVersion_Any, []*client.EventData{
		client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(`{}`)),
	}, nil)
	if err != nil {
		panic(err)
	}
	if err := t.Wait(); err != nil {
		panic(err)
	}
}

func BenchmarkReadEventSync(b *testing.B) {
	for n := 0; n < b.N; n++ {
		task, err := es.ReadEventAsync("BenchmarkReadEvent", 0, true, nil)
		if err != nil {
			b.Error(err)
			return
		}
		if err := task.Wait(); err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkReadEventAsync(b *testing.B) {
	var err error
	_tasks := make([]*tasks.Task, b.N)
	for n := 0; n < b.N; n++ {
		_tasks[n], err = es.ReadEventAsync("BenchmarkReadEvent", 0, true, nil)
		if err != nil {
			b.Error(err)
			return
		}
	}
	for _, task := range _tasks {
		if err := task.Wait(); err != nil {
			b.Error(err)
			return
		}
	}
}

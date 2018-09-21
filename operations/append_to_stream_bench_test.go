package operations_test

import (
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"github.com/satori/go.uuid"
	"testing"
)

func init() {
	ensureConnection()
}

func BenchmarkAppendToStreamSync(b *testing.B) {
	for n := 0; n < b.N; n++ {
		t, err := es.AppendToStreamAsync(
			"BenchmarkAppendToStreamSync",
			client.ExpectedVersion_Any,
			[]*client.EventData{
				client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``)),
			},
			nil,
		)
		if err != nil {
			b.Error(err)
			return
		}
		if err := t.Wait(); err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkAppendToStreamAsync(b *testing.B) {
	var err error
	_tasks := make([]*tasks.Task, b.N)
	for n := 0; n < b.N; n++ {
		_tasks[n], err = es.AppendToStreamAsync(
			"BenchmarkAppendToStreamAsync",
			client.ExpectedVersion_Any,
			[]*client.EventData{
				client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``)),
			},
			nil,
		)
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

func AppendToStreamBatchSync(batchSize int, b *testing.B) {
	events := make([]*client.EventData, batchSize)
	for n := 0; n < b.N; n += batchSize {
		for i := 0; i < batchSize; i++ {
			events[i] = client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``))
		}
		t, err := es.AppendToStreamAsync(
			"AppendToStreamBatchSync",
			client.ExpectedVersion_Any,
			events,
			nil,
		)
		if err != nil {
			b.Error(err)
			return
		}
		if err := t.Wait(); err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkAppendToStreamBatchSync_20(b *testing.B)   { AppendToStreamBatchSync(20, b) }
func BenchmarkAppendToStreamBatchSync_100(b *testing.B)  { AppendToStreamBatchSync(100, b) }
func BenchmarkAppendToStreamBatchSync_200(b *testing.B)  { AppendToStreamBatchSync(200, b) }
func BenchmarkAppendToStreamBatchSync_500(b *testing.B)  { AppendToStreamBatchSync(500, b) }
func BenchmarkAppendToStreamBatchSync_1000(b *testing.B) { AppendToStreamBatchSync(1000, b) }

func AppendToStreamBatchAsync(batchSize int, b *testing.B) {
	var err error
	stream := "AppendToStreamBatchAsync-" + uuid.Must(uuid.NewV4()).String()
	_tasks := make([]*tasks.Task, (b.N/batchSize)+1)
	events := make([]*client.EventData, batchSize)
	c := 0
	for n := 0; n < b.N; n += batchSize {
		for i := 0; i < batchSize; i++ {
			events[i] = client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``))
		}
		_tasks[c], err = es.AppendToStreamAsync(
			stream,
			client.ExpectedVersion_Any,
			events,
			nil,
		)
		c++
		if err != nil {
			b.Error(err)
			return
		}
	}
	for _, task := range _tasks {
		if task == nil {
			break
		}
		if err := task.Wait(); err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkAppendToStreamBatchAsync_20(b *testing.B)   { AppendToStreamBatchAsync(20, b) }
func BenchmarkAppendToStreamBatchAsync_100(b *testing.B)  { AppendToStreamBatchAsync(100, b) }
func BenchmarkAppendToStreamBatchAsync_200(b *testing.B)  { AppendToStreamBatchAsync(200, b) }
func BenchmarkAppendToStreamBatchAsync_500(b *testing.B)  { AppendToStreamBatchAsync(500, b) }
func BenchmarkAppendToStreamBatchAsync_1000(b *testing.B) { AppendToStreamBatchAsync(1000, b) }

func BenchmarkAppendToStreamSyncWithExpectedVersion(b *testing.B) {
	stream := "BenchmarkAppendToStreamSyncWithExpectedVersion-" + uuid.Must(uuid.NewV4()).String()
	for n := 0; n < b.N; n++ {
		t, err := es.AppendToStreamAsync(
			stream,
			n-1,
			[]*client.EventData{
				client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``)),
			},
			nil,
		)
		if err != nil {
			b.Error(err)
			return
		}
		if err := t.Wait(); err != nil {
			b.Error(n)
			return
		}
	}
}

func BenchmarkAppendToStreamAsyncWithExpectedVersion(b *testing.B) {
	stream := "BenchmarkAppendToStreamAsyncWithExpectedVersion-" + uuid.Must(uuid.NewV4()).String()
	var err error
	_tasks := make([]*tasks.Task, b.N)
	for n := 0; n < b.N; n++ {
		_tasks[n], err = es.AppendToStreamAsync(
			stream,
			n-1,
			[]*client.EventData{
				client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``)),
			},
			nil,
		)
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

func AppendToStreamBatchSyncWithExpectedVersion(batchSize int, b *testing.B) {
	stream := "BenchmarkAppendToStreamBatchSyncWithExpectedVersion-" + uuid.Must(uuid.NewV4()).String()
	events := make([]*client.EventData, batchSize)
	for n := 0; n < b.N; n += batchSize {
		for i := 0; i < batchSize; i++ {
			events[i] = client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``))
		}
		t, err := es.AppendToStreamAsync(
			stream,
			n-1,
			events,
			nil,
		)
		if err != nil {
			b.Error(err)
			return
		}
		if err := t.Wait(); err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkAppendToStreamBatchSyncWithExpectedVersion_20(b *testing.B) {
	AppendToStreamBatchSyncWithExpectedVersion(20, b)
}
func BenchmarkAppendToStreamBatchSyncWithExpectedVersion_100(b *testing.B) {
	AppendToStreamBatchSyncWithExpectedVersion(100, b)
}
func BenchmarkAppendToStreamBatchSyncWithExpectedVersion_200(b *testing.B) {
	AppendToStreamBatchSyncWithExpectedVersion(200, b)
}
func BenchmarkAppendToStreamBatchSyncWithExpectedVersion_500(b *testing.B) {
	AppendToStreamBatchSyncWithExpectedVersion(500, b)
}
func BenchmarkAppendToStreamBatchWithExpectedVersionSync_1000(b *testing.B) {
	AppendToStreamBatchSyncWithExpectedVersion(1000, b)
}

func AppendToStreamBatchAsyncWithExpectedVersion(batchSize int, b *testing.B) {
	var err error
	stream := "AppendToStreamBatchAsync-" + uuid.Must(uuid.NewV4()).String()
	_tasks := make([]*tasks.Task, (b.N/batchSize)+1)
	events := make([]*client.EventData, batchSize)
	c := 0
	for n := 0; n < b.N; n += batchSize {
		for i := 0; i < batchSize; i++ {
			events[i] = client.NewEventData(uuid.Must(uuid.NewV4()), "Benchmark", true, []byte(`{}`), []byte(``))
		}
		_tasks[c], err = es.AppendToStreamAsync(
			stream,
			n-1,
			events,
			nil,
		)
		c++
		if err != nil {
			b.Error(err)
			return
		}
	}
	for _, task := range _tasks {
		if task == nil {
			break
		}
		if err := task.Wait(); err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkAppendToStreamBatchAsyncWithExpectedVersion_20(b *testing.B) {
	AppendToStreamBatchAsyncWithExpectedVersion(20, b)
}
func BenchmarkAppendToStreamBatchAsyncWithExpectedVersion_100(b *testing.B) {
	AppendToStreamBatchAsyncWithExpectedVersion(100, b)
}
func BenchmarkAppendToStreamBatchAsyncWithExpectedVersion_200(b *testing.B) {
	AppendToStreamBatchAsyncWithExpectedVersion(200, b)
}
func BenchmarkAppendToStreamBatchAsyncWithExpectedVersion_500(b *testing.B) {
	AppendToStreamBatchAsyncWithExpectedVersion(500, b)
}
func BenchmarkAppendToStreamBatchAsyncWithExpectedVersion_1000(b *testing.B) {
	AppendToStreamBatchAsyncWithExpectedVersion(1000, b)
}

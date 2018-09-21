package subscriptions_test

import (
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"github.com/satori/go.uuid"
	"net/url"
	"testing"
	"time"
)

var es client.Connection

func init() {
	ensureConnection()
}

func ensureConnection() {
	if es != nil {
		return
	}

	var err error

	uri, _ := url.Parse("tcp://127.0.0.1:1113/")
	es, err = gesclient.Create(client.DefaultConnectionSettings, uri, "benchmark")
	if err != nil {
		panic(err)
	}
	es.Disconnected().Add(func(event client.Event) error { panic("disconnected") })
	es.ConnectAsync().Wait()
	time.Sleep(100 * time.Millisecond)
}

func appendToStreamBatchAsync(stream string, batchSize int, size int) error {
	var err error
	_tasks := make([]*tasks.Task, (size/batchSize)+1)
	events := make([]*client.EventData, batchSize)
	c := 0
	for n := 0; n < size; n += batchSize {
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
			return err
		}
	}
	for _, task := range _tasks {
		if task == nil {
			break
		}
		if err := task.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func BenchmarkSubscribeToStream(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stream := "AppendToStreamBatchAsync-" + uuid.Must(uuid.NewV4()).String()
		c := make(chan *client.ResolvedEvent, 100000)
		task, _ := es.SubscribeToStreamAsync(stream, true,
			func(s client.EventStoreSubscription, r *client.ResolvedEvent) error { c <- r; return nil },
			func(s client.EventStoreSubscription, dr client.SubscriptionDropReason, err error) error {
				close(c)
				return nil
			},
			nil,
		)
		s := task.Result().(client.EventStoreSubscription)
		appendToStreamBatchAsync(stream, 1000, 100000)
		for e := range c {
			if e.OriginalEventNumber() == 99999 {
				break
			}
		}
		s.Close()
	}
}

func BenchmarkSubscribeToStreamPerEvent(b *testing.B) {
	stream := "AppendToStreamBatchAsync-" + uuid.Must(uuid.NewV4()).String()
	c := make(chan *client.ResolvedEvent, b.N)
	task, _ := es.SubscribeToStreamAsync(stream, true,
		func(s client.EventStoreSubscription, r *client.ResolvedEvent) error { c <- r; return nil },
		func(s client.EventStoreSubscription, dr client.SubscriptionDropReason, err error) error {
			close(c)
			return nil
		},
		nil,
	)
	s := task.Result().(client.EventStoreSubscription)
	appendToStreamBatchAsync(stream, 1000, b.N)
	for e := range c {
		if e.OriginalEventNumber() == b.N-1 {
			break
		}
	}
	s.Close()
}

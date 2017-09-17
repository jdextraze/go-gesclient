package operations_test

import (
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"net/url"
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

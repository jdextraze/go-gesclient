package main

import (
	"encoding/json"
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/satori/go.uuid"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func main() {
	var debug bool
	var addr string
	var stream string
	var interval int

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.IntVar(&interval, "interval", 1000000, "Publish interval in microseconds")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

	uri, err := url.Parse(addr)
	if err != nil {
		log.Fatalf("Error parsing address: %v", err)
	}
	c, err := gesclient.Create(client.DefaultConnectionSettings, uri, "Publisher")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	for {
		select {
		case <-ch:
			c.Close()
			return
		default:
		}
		data, _ := json.Marshal(&TestEvent{})
		evt := client.NewEventData(uuid.NewV4(), "TestEvent", true, data, nil)
		result := &client.WriteResult{}
		task, err := c.AppendToStreamAsync(stream, client.ExpectedVersion_Any, []*client.EventData{evt}, nil)
		if err != nil {
			log.Printf("Error occured while appending to stream: %v", err)
		} else if err := task.Result(result); err != nil {
			log.Printf("Error occured while waiting for result of appending to stream: %v", err)
		} else {
			log.Printf("AppendToStream result: %v", result)
		}
		<-time.After(time.Duration(interval) * time.Microsecond)
	}
}

type TestEvent struct{}

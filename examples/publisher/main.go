package main

import (
	"encoding/json"
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/flags"
	"github.com/satori/go.uuid"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var stream string
	var interval int

	flags.Init(flag.CommandLine)
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.IntVar(&interval, "interval", 1000000, "Publish interval in microseconds")
	flag.Parse()

	if flags.Debug() {
		gesclient.Debug()
	}

	c, err := flags.CreateConnection("Publisher")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	c.Connected().Add(func(evt client.Event) error { log.Printf("Connected: %+v", evt); return nil })
	c.Disconnected().Add(func(evt client.Event) error { log.Printf("Disconnected: %+v", evt); return nil })
	c.Reconnecting().Add(func(evt client.Event) error { log.Printf("Reconnecting: %+v", evt); return nil })
	c.Closed().Add(func(evt client.Event) error { log.Fatalf("Connection closed: %+v", evt); return nil })
	c.ErrorOccurred().Add(func(evt client.Event) error { log.Printf("Error: %+v", evt); return nil })
	c.AuthenticationFailed().Add(func(evt client.Event) error { log.Printf("Auth failed: %+v", evt); return nil })

	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	for {
		select {
		case <-ch:
			c.Close()
			time.Sleep(10 * time.Millisecond)
			return
		default:
		}
		data, _ := json.Marshal(&TestEvent{})
		evt := client.NewEventData(uuid.Must(uuid.NewV4()), "TestEvent", true, data, nil)
		log.Printf("-> '%s': %+v", stream, evt)
		task, err := c.AppendToStreamAsync(stream, client.ExpectedVersion_Any, []*client.EventData{evt}, nil)
		if err != nil {
			log.Printf("Error occured while appending to stream: %v", err)
		} else if err := task.Error(); err != nil {
			log.Printf("Error occured while waiting for result of appending to stream: %v", err)
		} else {
			result := task.Result().(*client.WriteResult)
			log.Printf("<- %+v", result)
		}
		<-time.After(time.Duration(interval) * time.Microsecond)
	}
}

type TestEvent struct{}

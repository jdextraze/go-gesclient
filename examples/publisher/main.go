package main

import (
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/models"
	"net/url"
	"log"
	"github.com/satori/go.uuid"
	"encoding/json"
	"time"
	"os"
	"os/signal"
)

func main() {
	var addr string
	var stream string

	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.Parse()

	uri, err := url.Parse(addr)
	if err != nil {
		log.Fatalf("Error parsing address: %v", err)
	}
	c, err := gesclient.Create(models.DefaultConnectionSettings, uri, "Publisher")
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
		evt := models.NewEventData(uuid.NewV4(), "TestEvent", true, data, nil)
		result := &models.WriteResult{}
		task, err := c.AppendToStreamAsync(stream, models.ExpectedVersion_Any, []*models.EventData{evt}, nil)
		if err != nil {
			log.Printf("Error occured while appending to stream: %v", err)
		} else if err := task.Result(result); err != nil {
			log.Printf("Error occured while waiting for result of appending to stream: %v", err)
		} else {
			log.Printf("AppendToStream result: %v", result)
		}
		<-time.After(time.Second)
	}
}

type TestEvent struct {}

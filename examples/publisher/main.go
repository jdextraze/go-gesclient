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
	"net"
	"strings"
)

func main() {
	var debug bool
	var addr string
	var stream string
	var interval int
	var verbose bool

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.IntVar(&interval, "interval", 1000000, "Publish interval in microseconds")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging (Requires debug)")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

	c := getConnection(addr, verbose)
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

func getConnection(addr string, verbose bool) client.Connection {
	settingsBuilder := client.CreateConnectionSettings()

	var uri *url.URL
	var err error
	gossipSeeds := strings.Split(addr, ",")
	if len(gossipSeeds) > 0 {
		endpoints := make([]*net.TCPAddr, len(gossipSeeds))
		for i, g := range gossipSeeds {
			uri, err := url.Parse(g)
			if err != nil {
				log.Fatalf("Error parsing address: %v", err)
			}
			endpoints[i], err = net.ResolveTCPAddr("tcp", uri.Host)
			if err != nil {
				log.Fatalf("Error resolving: %v", uri.Host)
			}
		}
		settingsBuilder.SetGossipSeedEndPoints(endpoints)
	} else {
		uri, err = url.Parse(addr)
		if err != nil {
			log.Fatalf("Error parsing address: %v", err)
		}
	}

	if verbose {
		settingsBuilder.EnableVerboseLogging()
	}

	c, err := gesclient.Create(settingsBuilder.Build(), uri, "AllCatchUpSubscriber")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	return c
}

type TestEvent struct{}

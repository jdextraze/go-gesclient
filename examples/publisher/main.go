package main

import (
	"encoding/json"
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/satori/go.uuid"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
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
		task, err := c.AppendToStreamAsync(stream, client.ExpectedVersion_Any, []*client.EventData{evt}, nil)
		if err != nil {
			log.Printf("Error occured while appending to stream: %v", err)
		} else if err := task.Error(); err != nil {
			log.Printf("Error occured while waiting for result of appending to stream: %v", err)
		} else {
			result := task.Result().(*client.WriteResult)
			log.Printf("AppendToStream result: %v", result)
		}
		<-time.After(time.Duration(interval) * time.Microsecond)
	}
}

func getConnection(addr string, verbose bool) client.Connection {
	settingsBuilder := client.CreateConnectionSettings()

	var uri *url.URL
	var err error
	if !strings.Contains(addr, "://") {
		gossipSeeds := strings.Split(addr, ",")
		endpoints := make([]*net.TCPAddr, len(gossipSeeds))
		for i, gossipSeed := range gossipSeeds {
			endpoints[i], err = net.ResolveTCPAddr("tcp", gossipSeed)
			if err != nil {
				log.Fatalf("Error resolving: %v", gossipSeed)
			}
		}
		settingsBuilder.SetGossipSeedEndPoints(endpoints)
	} else {
		uri, err = url.Parse(addr)
		if err != nil {
			log.Fatalf("Error parsing address: %v", err)
		}

		if uri.User != nil {
			username := uri.User.Username()
			password, _ := uri.User.Password()
			settingsBuilder.SetDefaultUserCredentials(client.NewUserCredentials(username, password))
		}
	}

	if verbose {
		settingsBuilder.EnableVerboseLogging()
	}

	c, err := gesclient.Create(settingsBuilder.Build(), uri, "AllCatchUpSubscriber")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	c.Connected().Add(func(evt client.Event) error { log.Printf("Connected: %v", evt); return nil })
	c.Disconnected().Add(func(evt client.Event) error { log.Printf("Disconnected: %v", evt); return nil })
	c.Reconnecting().Add(func(evt client.Event) error { log.Printf("Reconnecting: %v", evt); return nil })
	c.Closed().Add(func(evt client.Event) error { panic("Connection closed") })
	c.ErrorOccurred().Add(func(evt client.Event) error { log.Printf("Error: %v", evt); return nil })
	c.AuthenticationFailed().Add(func(evt client.Event) error { log.Printf("Auth failed: %v", evt); return nil })

	return c
}

type TestEvent struct{}

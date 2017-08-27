package main

import (
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
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
	var lastCheckpoint int
	var verbose bool

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.IntVar(&lastCheckpoint, "lastCheckpoint", -1, "Last checkpoint")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging (Requires debug)")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

	c := getConnection(addr, verbose)
	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	settings := client.NewCatchUpSubscriptionSettings(client.CatchUpDefaultMaxPushQueueSize,
		client.CatchUpDefaultReadBatchSize, verbose, true)
	var fromEventNumber *int
	if lastCheckpoint >= 0 {
		fromEventNumber = &lastCheckpoint
	}
	sub, err := c.SubscribeToStreamFrom(stream, fromEventNumber, settings, eventAppeared, liveProcessingStarted,
		subscriptionDropped, nil)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else {
		log.Printf("SubscribeToStreamFrom result: %v", sub)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch

		sub.Stop()
	}

	c.Close()
	time.Sleep(10 * time.Millisecond)
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

func eventAppeared(s client.CatchUpSubscription, e *client.ResolvedEvent) error {
	log.Printf("event appeared: %d | %s", e.OriginalEventNumber(), string(e.OriginalEvent().Data()))
	return nil
}

func liveProcessingStarted(s client.CatchUpSubscription) error {
	log.Println("Live processing started")
	return nil
}

func subscriptionDropped(s client.CatchUpSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}

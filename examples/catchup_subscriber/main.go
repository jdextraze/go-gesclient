package main

import (
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
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

	uri, err := url.Parse(addr)
	if err != nil {
		log.Fatalf("Error parsing address: %v", err)
	}
	settingsBuilder := client.CreateConnectionSettings()
	if verbose {
		settingsBuilder.EnableVerboseLogging()
	}
	c, err := gesclient.Create(settingsBuilder.Build(), uri, "CatchUpSubscriber")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

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

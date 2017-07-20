package main

import (
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"log"
	"net/url"
	"os"
	"os/signal"
)

func main() {
	var debug bool
	var addr string
	var stream string
	var verbose bool

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
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
	c, err := gesclient.Create(settingsBuilder.Build(), uri, "AllSubscriber")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	user := client.NewUserCredentials("admin", "changeit")
	sub := &client.EventStoreSubscription{}
	task, err := c.SubscribeToAllAsync(true, eventAppeared, subscriptionDropped, user)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Result(sub); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		log.Printf("SubscribeToAll result: %v", sub)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch

		sub.Close()
	}

	c.Close()
}

func eventAppeared(s *client.EventStoreSubscription, e *client.ResolvedEvent) error {
	log.Printf("event appeared: %d | %s", e.OriginalEventNumber(), string(e.OriginalEvent().Data()))
	return nil
}

func subscriptionDropped(s *client.EventStoreSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}

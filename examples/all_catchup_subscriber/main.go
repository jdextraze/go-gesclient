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
	var preparePosition int64
	var commitPosition int64

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.Int64Var(&preparePosition, "prepare", 0, "Prepare position")
	flag.Int64Var(&commitPosition, "commit", 0, "Commit position")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

	uri, err := url.Parse(addr)
	if err != nil {
		log.Fatalf("Error parsing address: %v", err)
	}
	c, err := gesclient.Create(client.DefaultConnectionSettings, uri, "AllCatchUpSubscriber")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	user := client.NewUserCredentials("admin", "changeit")
	sub, err := c.SubscribeToAllFrom(client.NewPosition(commitPosition, preparePosition),
		client.CatchUpSubscriptionSettings_Default, eventAppeared, liveProcessingStarted, subscriptionDropped, user)
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
}

func eventAppeared(s client.CatchUpSubscription, e *client.ResolvedEvent) error {
	log.Printf("event appeared: %v | %s", e.OriginalPosition(), string(e.OriginalEvent().Data()))
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

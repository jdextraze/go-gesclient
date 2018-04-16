package main

import (
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/flags"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var stream string
	var preparePosition int64
	var commitPosition int64

	flags.Init(flag.CommandLine)
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.Int64Var(&preparePosition, "prepare", 0, "Prepare position")
	flag.Int64Var(&commitPosition, "commit", 0, "Commit position")
	flag.Parse()

	if flags.Debug() {
		gesclient.Debug()
	}

	c, err := flags.CreateConnection("AllCatchupSubscriber")
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

	user := client.NewUserCredentials("admin", "changeit")
	settings := client.NewCatchUpSubscriptionSettings(client.CatchUpDefaultMaxPushQueueSize,
		client.CatchUpDefaultReadBatchSize, flags.Verbose(), true)
	sub, err := c.SubscribeToAllFrom(client.NewPosition(commitPosition, preparePosition),
		settings, eventAppeared, liveProcessingStarted, subscriptionDropped, user)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else {
		log.Printf("SubscribeToStreamFrom result: %+v", sub)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch

		sub.Stop()
	}

	c.Close()
	time.Sleep(10 * time.Millisecond)
}

func eventAppeared(_ client.CatchUpSubscription, e *client.ResolvedEvent) error {
	log.Printf("event appeared: %+v | %s", e, string(e.OriginalEvent().Data()))
	return nil
}

func liveProcessingStarted(_ client.CatchUpSubscription) error {
	log.Println("Live processing started")
	return nil
}

func subscriptionDropped(_ client.CatchUpSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}

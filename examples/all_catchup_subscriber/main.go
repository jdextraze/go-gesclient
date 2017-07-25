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
	"strings"
	"net"
)

func main() {
	var debug bool
	var addr string
	var stream string
	var preparePosition int64
	var commitPosition int64
	var verbose bool

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.Int64Var(&preparePosition, "prepare", 0, "Prepare position")
	flag.Int64Var(&commitPosition, "commit", 0, "Commit position")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging (Need debug)")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

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

	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	user := client.NewUserCredentials("admin", "changeit")
	settings := client.NewCatchUpSubscriptionSettings(client.CatchUpDefaultMaxPushQueueSize,
		client.CatchUpDefaultReadBatchSize, verbose, true)
	sub, err := c.SubscribeToAllFrom(client.NewPosition(commitPosition, preparePosition),
		settings, eventAppeared, liveProcessingStarted, subscriptionDropped, user)
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

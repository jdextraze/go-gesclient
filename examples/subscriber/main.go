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
	var verbose bool

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging (Requires debug)")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

	c := getConnection(addr, verbose)
	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	sub := &client.EventStoreSubscription{}
	task, err := c.SubscribeToStreamAsync(stream, true, eventAppeared, subscriptionDropped, nil)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Result(sub); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		log.Printf("SubscribeToStream result: %v", sub)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch

		sub.Close()
		time.Sleep(10 * time.Millisecond)
	}

	c.Close()
	time.Sleep(10 * time.Millisecond)
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


func eventAppeared(s *client.EventStoreSubscription, e *client.ResolvedEvent) error {
	log.Printf("event appeared: %d | %s", e.OriginalEventNumber(), string(e.Event().Data()))
	return nil
}

func subscriptionDropped(s *client.EventStoreSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}

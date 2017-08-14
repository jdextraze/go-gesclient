package main

import (
	"flag"
	"fmt"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/internal"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"net"
	"strings"
)

var debug bool
var addr string
var stream string
var groupName string
var verbose bool

func main() {
	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.StringVar(&groupName, "group", "Default", "Group name")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging (Requires debug)")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

	switch flag.Arg(0) {
	case "create":
		createPersistentSubscription()
	case "delete":
		deletePersistentSubscription()
	case "subscribe":
		subscribe()
	default:
		fmt.Fprintf(os.Stderr, "Usage of %s: [flags] action\nAction: create | delete | subscribe\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func getConnection() client.Connection {
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

	return c
}

func closeConnection(c client.Connection) {
	c.Close()
	time.Sleep(10 * time.Millisecond)
}

func createPersistentSubscription() {
	c := getConnection()
	defer closeConnection(c)
	res := &client.PersistentSubscriptionCreateResult{}
	task, err := c.CreatePersistentSubscriptionAsync(stream, groupName, client.DefaultPersistentSubscriptionSettings,
		nil)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Result(res); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		log.Printf("CreatePersistentSubscriptionAsync result: %v", res)
	}
}

func deletePersistentSubscription() {
	c := getConnection()
	defer closeConnection(c)
	res := &client.PersistentSubscriptionDeleteResult{}
	task, err := c.DeletePersistentSubscriptionAsync(stream, groupName, nil)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Result(res); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		log.Printf("CreatePersistentSubscriptionAsync result: %v", res)
	}
}

func subscribe() {
	c := getConnection()
	defer closeConnection(c)
	sub := internal.NilPersistentSubscription()
	task, err := c.ConnectToPersistentSubscriptionAsync(stream, groupName, eventAppeared, subscriptionDropped,
		nil, 10, true)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Result(sub); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		log.Printf("SubscribeToStream result: %v", sub)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch

		sub.Stop()
	}
}

func eventAppeared(s client.PersistentSubscription, e *client.ResolvedEvent) error {
	log.Printf("event appeared: %d | %s", e.OriginalEventNumber(), string(e.Event().Data()))
	return nil
}

func subscriptionDropped(s client.PersistentSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}

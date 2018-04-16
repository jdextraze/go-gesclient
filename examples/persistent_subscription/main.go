package main

import (
	"flag"
	"fmt"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/flags"
	"log"
	"os"
	"os/signal"
	"time"
)

var stream string
var groupName string

func main() {
	flags.Init(flag.CommandLine)
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.StringVar(&groupName, "group", "Default", "Group name")
	flag.Parse()

	if flags.Debug() {
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
	c, err := flags.CreateConnection("AllCatchupSubscriber")
	if err != nil {
		log.Fatalf("Error creating connection: %v", err)
	}

	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	c.Connected().Add(func(evt client.Event) error { log.Printf("Connected: %+v", evt); return nil })
	c.Disconnected().Add(func(evt client.Event) error { log.Printf("Disconnected: %+v", evt); return nil })
	c.Reconnecting().Add(func(evt client.Event) error { log.Printf("Reconnecting: %+v", evt); return nil })
	c.Closed().Add(func(evt client.Event) error { log.Fatalf("Connection closed: %+v", evt); return nil })
	c.ErrorOccurred().Add(func(evt client.Event) error { log.Printf("Error: %+v", evt); return nil })
	c.AuthenticationFailed().Add(func(evt client.Event) error { log.Printf("Auth failed: %+v", evt); return nil })

	return c
}

func closeConnection(c client.Connection) {
	c.Close()
	time.Sleep(10 * time.Millisecond)
}

func createPersistentSubscription() {
	c := getConnection()
	defer closeConnection(c)
	task, err := c.CreatePersistentSubscriptionAsync(stream, groupName, client.DefaultPersistentSubscriptionSettings,
		nil)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		res := task.Result().(*client.PersistentSubscriptionCreateResult)
		log.Printf("CreatePersistentSubscriptionAsync result: %+v", res)
	}
}

func deletePersistentSubscription() {
	c := getConnection()
	defer closeConnection(c)
	task, err := c.DeletePersistentSubscriptionAsync(stream, groupName, nil)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		res := task.Result().(*client.PersistentSubscriptionDeleteResult)
		log.Printf("CreatePersistentSubscriptionAsync result: %+v", res)
	}
}

func subscribe() {
	c := getConnection()
	defer closeConnection(c)
	task, err := c.ConnectToPersistentSubscriptionAsync(stream, groupName, eventAppeared, subscriptionDropped,
		nil, 10, true)
	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		sub := task.Result().(client.PersistentSubscription)
		log.Printf("SubscribeToStream result: %+v", sub)

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch

		sub.Stop()
	}
}

func eventAppeared(_ client.PersistentSubscription, e *client.ResolvedEvent) error {
	log.Printf("event appeared: %+v | %s", e, string(e.Event().Data()))
	return nil
}

func subscriptionDropped(_ client.PersistentSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}

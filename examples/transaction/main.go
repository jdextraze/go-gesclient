package main

import (
	"encoding/json"
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/flags"
	"github.com/satori/go.uuid"
	"log"
)

func main() {
	var stream string
	var expectedVersion int
	var transactionId int64

	flags.Init(flag.CommandLine)
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.IntVar(&expectedVersion, "expected-version", client.ExpectedVersion_Any, "expected version")
	flag.Int64Var(&transactionId, "continue", -1, "Continue transaction with id")
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

	var t *client.Transaction
	if transactionId < 0 {
		task, err := c.StartTransactionAsync(stream, expectedVersion, nil)
		if err != nil {
			log.Fatalf("Failed starting an async transaction: %v", err)
		} else if err := task.Error(); err != nil {
			log.Fatalf("Failed waiting to start an async transaction: %v", err)
		} else {
			t = task.Result().(*client.Transaction)
		}
	} else {
		t = c.ContinueTransaction(transactionId, nil)
	}

	switch flag.Arg(0) {
	case "write":
		write(t)
	case "commit":
		commit(t)
	case "rollback":
		rollback(t)
	default:
		log.Fatalf("Unknown action. Use write, commit or rollback.")
	}
}

type TestEvent struct{}

func write(t *client.Transaction) {
	log.Printf("Writing to transaction #%d", t.TransactionId())
	data, _ := json.Marshal(&TestEvent{})
	evt := client.NewEventData(uuid.Must(uuid.NewV4()), "TestEvent", true, data, nil)
	task, err := t.WriteAsync([]*client.EventData{evt})
	if err != nil {
		log.Printf("Error occured while writing to transaction: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of writing to transaction: %v", err)
	}
}

func commit(t *client.Transaction) {
	log.Printf("Committing transaction #%d", t.TransactionId())
	task, err := t.CommitAsync()
	if err != nil {
		log.Printf("Error occured while committing transaction: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of committing transaction: %v", err)
	} else {
		result := task.Result().(*client.WriteResult)
		log.Printf("<- %+v", result)
	}
}

func rollback(t *client.Transaction) {
	log.Printf("Rollbacking transaction #%d", t.TransactionId())
	if err := t.Rollback(); err != nil {
		log.Printf("Error occured while rollbacking transaction: %v", err)
	}
}

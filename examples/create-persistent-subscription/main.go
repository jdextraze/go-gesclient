package main

import (
	"os"
	"fmt"
	"github.com/jdextraze/go-gesclient"
	"encoding/json"
	"time"
	"github.com/satori/go.uuid"
	"log"
)

type Tested struct {
	Test string
}

func main() {
	gesclient.Debug()

	if len(os.Args) != 2 {
		fmt.Println("Usage: example [event-store-addr]")
		fmt.Println("  event-store-addr    Event store address (Ex: 192.168.0.100:1113)")
		return
	}

	c := gesclient.NewConnection(os.Args[1])
	c.WaitForConnection()

	streamName := "Test-" + uuid.NewV4().String()
	log.Println(streamName)

	run := true
	done := make(chan struct{}, 1)
	go func() {
		for run {
			data, _ := json.Marshal(&Tested{})
			evt := gesclient.NewEventData(uuid.NewV4(), "Tested", true, data, nil)
			create, err := c.AppendToStreamAsync(streamName, gesclient.ExpectedVersion_Any, []*gesclient.EventData{evt}, nil)
			if err != nil {
				log.Println("AppendToStream failed", err)
				break
			} else {
				log.Println("CreateEvent:", <-create)
			}
			time.Sleep(time.Second)
		}
		done <- struct{}{}
	}()

	time.Sleep(3 * time.Second)

	groupName := "test"
	user := gesclient.NewUserCredentials("admin", "changeit")

	settings := gesclient.NewPersistentSubscriptionSettings(false, -1, false, 30 * time.Second, 500, 500, 10, 20,
		2 * time.Second, 10, 1000, 0, gesclient.SystemConsumerStrategies_RoundRobin)
	ares, _ := c.CreatePersistentSubscriptionAsync(streamName, groupName, settings, user)
	res := <- ares
	log.Println(">>>", res.GetError())

	asub1, _ := c.ConnectToPersistentSubscriptionAsync(streamName, groupName, nil)
	sub1 := <- asub1
	log.Println(">>>", sub1.Error())
	go subscriber(sub1, "First: ")

	asub2, _ := c.ConnectToPersistentSubscriptionAsync(streamName, groupName, nil)
	sub2 := <- asub2
	log.Println(">>>", sub2.Error())
	go subscriber(sub2, "Second: ")

	//c.DeletePersistentSubscriptionAsync(streamName, groupName, nil)

	<- done
}

func subscriber(sub gesclient.PersistentSubscription, prefix string) {
	events := sub.Events()
	dropped := sub.Dropped()
	run := true
	for run == true {
		select {
		case e, ok := <-events:
			if !ok {
				run = false
				break
			}
			log.Println(prefix, "StreamEventAppeared:", e.Event().EventId(), string(e.Event().Data()))
		case d := <-dropped:
			log.Println(prefix, "Subscription dropped", d)
			run = false
		}
	}
	log.Println(prefix, "Subscription ended!")
}
package main

import (
	"bitbucket.org/jdextraze/go-gesclient"
	"encoding/json"
	"github.com/satori/go.uuid"
	"log"
	"time"
)

type Tested struct {
	Test string
}

func main() {
	gesclient.Debug()

	c := gesclient.NewConnection("192.168.22.10:1113")
	c.WaitForConnection()

	streamName := "Test-" + uuid.NewV4().String()
	log.Println(streamName)

	sub, err := c.SubscribeToStream(streamName)
	log.Println("SubscribeToStream:", sub, err)
	go func() {
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
				log.Println("StreamEventAppeared:", e.Event().EventId(), string(e.Event().Data()))
			case <-dropped:
				log.Println("Subscription dropped")
				run = false
			}
		}
		log.Println("Subscription ended!")
	}()

	run := true
	go func() {
		for run {
			data, _ := json.Marshal(&Tested{})
			evt := gesclient.NewEventData(uuid.NewV4(), "Tested", true, data, nil)
			create, err := c.AppendToStreamAsync(streamName, gesclient.ExpectedVersion_Any, []*gesclient.EventData{evt}, nil)
			if err != nil {
				log.Println("AppendToStream failed", err)
			} else {
				log.Println("CreateEvent:", <-create)
			}
			time.Sleep(time.Second)
		}
	}()

	<-time.After(time.Second * 10)
	run = false

	err = sub.Unsubscribe()
	log.Println("Unsubscribe:", err)

	read, err := c.ReadStreamEventsForward(streamName, 0, 1000)
	log.Println("ReadStreamEventsForward:", read, err)

	err = c.Close()
	log.Println(err)
}

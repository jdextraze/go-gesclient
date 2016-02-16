package main

import (
    _ "bitbucket.org/jdextraze/go-gesclient"
    "bitbucket.org/jdextraze/go-gesclient"
    "log"
    "os"
    "bufio"
    "encoding/json"
    "github.com/satori/go.uuid"
)

type Tested struct {
    Test string
}

func main() {
    gesclient.Debug()

    c, err := gesclient.NewConnection("192.168.22.10:1113")
    if err != nil {
        log.Fatalln(err)
    }

    sub, err := c.SubscribeToStream("Test-1")
    log.Println("SubscribeToStream:", *sub, err)
    go func() {
        run := true
        for run == true {
            select {
            case e, ok := <- sub.Events :
                if !ok {
                    run = false
                    break
                }
                evtId, _ := uuid.FromBytes(e.Event.Event.EventId)
                log.Println("StreamEventAppeared:", evtId, string(e.Event.Event.Data))
            }
        }
        log.Println("Subscription ended!")
    }()

    data, _ := json.Marshal(&Tested{})
    create, err := c.CreateEvent("Test-1", "Tested", true, data, nil, gesclient.ExpectedVersion_Any)
    log.Println("CreateEvent:", create, err)

    bufio.NewReader(os.Stdin).ReadString('\n')

    uns, err := sub.Unsubscribe()
    log.Println("Unsubscribe:", *uns, err)

//    read, err := c.ReadStreamEventsForward("Test-1", 0, 1000)
//    log.Println("ReadStreamEventsForward:", read, err)

    log.Println(c.Close())
}

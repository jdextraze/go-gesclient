package main

import (
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"log"
	"github.com/jdextraze/go-gesclient/flags"
)

func main() {
	var settingsJsonBytes []byte

	flags.Init(flag.CommandLine)
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		return
	}
	settingsJsonBytes = []byte(flag.Arg(0))

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

	systemSettings, err := client.SystemSettingsFromJsonBytes(settingsJsonBytes)
	if err != nil {
		log.Fatalf("Invalid metadata: %v", err)
	}

	if t, err := c.SetSystemSettings(systemSettings, nil); err != nil {
		log.Fatalf("Failed seting system settings: %v", err)
	} else if err := t.Error(); err != nil {
		log.Fatalf("Failed getting result for seting system settings: %v", err)
	} else {
		result := t.Result().(*client.WriteResult)
		log.Printf("result: %+v", result)
	}

	c.Close()
}
package main

import (
	"log"
	"github.com/jdextraze/go-gesclient"
	"flag"
	"net/url"
	"net"
	"github.com/jdextraze/go-gesclient/client"
	"strings"
)

func main() {
	var debug bool
	var addr string
	var stream string
	var metadata string
	var verbose bool

	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&addr, "endpoint", "tcp://127.0.0.1:1113", "EventStore address")
	flag.StringVar(&stream, "stream", "Default", "Stream ID")
	flag.StringVar(&metadata, "metadata", "", "Metadata")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging (Requires debug)")
	flag.Parse()

	if debug {
		gesclient.Debug()
	}

	c := getConnection(addr, verbose)
	if err := c.ConnectAsync().Wait(); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	data, err := client.StreamMetadataFromJsonBytes([]byte(metadata))
	if err != nil {
		log.Fatalf("Invalid metadata: %v", err)
	}

	if t, err := c.SetStreamMetadataAsync(stream, client.ExpectedVersion_Any, data, nil); err != nil {
		log.Fatalf("Failed getting stream metadata: %v", err)
	} else if err := t.Error(); err != nil {
		log.Fatalf("Failed getting stream metadata result: %v", err)
	} else {
		result := t.Result().(*client.WriteResult)
		log.Printf("result: %v", result)
	}

	c.Close()
}

func getConnection(addr string, verbose bool) client.Connection {
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

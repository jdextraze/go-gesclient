package flags

import (
	"flag"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"log"
	"net"
	"net/url"
	"strings"
)

var (
	debug         bool
	endpoint      string
	sslHost       string
	sslSkipVerify bool
	verbose       bool
)

func Init(fs *flag.FlagSet) {
	fs.BoolVar(&debug, "debug", false, "Debug")
	fs.StringVar(&endpoint, "endpoint", "tcp://admin:changeit@127.0.0.1:1113", "EventStore address")
	fs.StringVar(&sslHost, "ssl-host", "", "SSL Host")
	fs.BoolVar(&sslSkipVerify, "ssl-skip-verify", false, "SSL skip unsecure verify")
	fs.BoolVar(&verbose, "verbose", false, "Verbose logging (Requires debug)")
}

func Debug() bool { return debug }

func Endpoint() string { return endpoint }

func SslHost() string { return sslHost }

func SslSkipVerify() bool { return sslSkipVerify }

func Verbose() bool { return verbose }

func CreateConnection(name string) (client.Connection, error) {
	settingsBuilder := client.CreateConnectionSettings()

	var uri *url.URL
	var err error
	if !strings.Contains(endpoint, "://") {
		gossipSeeds := strings.Split(endpoint, ",")
		endpoints := make([]*net.TCPAddr, len(gossipSeeds))
		for i, gossipSeed := range gossipSeeds {
			endpoints[i], err = net.ResolveTCPAddr("tcp", gossipSeed)
			if err != nil {
				log.Fatalf("Error resolving: %v", gossipSeed)
			}
		}
		settingsBuilder.SetGossipSeedEndPoints(endpoints)
	} else {
		uri, err = url.Parse(endpoint)
		if err != nil {
			log.Fatalf("Error parsing address: %v", err)
		}

		if uri.User != nil {
			username := uri.User.Username()
			password, _ := uri.User.Password()
			settingsBuilder.SetDefaultUserCredentials(client.NewUserCredentials(username, password))
		}
	}

	if sslHost != "" {
		settingsBuilder.UseSslConnection(sslHost, !sslSkipVerify)
	}

	if verbose {
		settingsBuilder.EnableVerboseLogging()
	}

	return gesclient.Create(settingsBuilder.Build(), uri, name)
}

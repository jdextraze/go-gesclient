package main

import (
	"flag"
	"fmt"
	"github.com/jdextraze/go-gesclient"
	"github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/projections"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

type Command struct {
	description  string
	flagsHandler func()
	handler      func()
}

var (
	commands        map[string]*Command
	userCredentials *client.UserCredentials
	manager         *projections.Manager
	name            string
)

func init() {
	commands = map[string]*Command{
		"list-all": {"List all projections", nil, listAll},
		"enable":   {"Enable projection", projectionNameFlag, enable},
		"disable":  {"Disable projection", projectionNameFlag, disable},
	}
}

func main() {
	var debug bool
	var endpoint string

	flag.BoolVar(&debug, "debug", false, "Output debug information")
	flag.StringVar(&endpoint, "endpoint", "http://admin:changeit@localhost:2113/", "HTTP endpoint")

	if len(os.Args) <= 1 || commands[os.Args[1]] == nil {
		commandsUsage()
		os.Exit(2)
	}

	command := commands[os.Args[1]]
	if command.flagsHandler != nil {
		command.flagsHandler()
	}
	flag.CommandLine.Parse(os.Args[2:])

	if debug {
		gesclient.Debug()
	}

	httpUrl, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("Failed parsing endpoint URL: %v", httpUrl)
	}
	if httpUrl.User != nil {
		password, _ := httpUrl.User.Password()
		userCredentials = client.NewUserCredentials(httpUrl.User.Username(), password)
	}

	addr, err := net.ResolveTCPAddr("tcp", httpUrl.Host)
	if err != nil {
		log.Fatalf("Failed resolving tcp address: %v", err)
	}

	manager = projections.NewManager(addr, time.Second*3)

	command.handler()
}

func commandsUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [command] [flags]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "The commands are:\n")
	for k, v := range commands {
		fmt.Fprintf(os.Stderr, "  %s%s%s\n", k, strings.Repeat(" ", 12-len(k)), v.description)
	}
}

func listAll() {
	task := manager.ListAllAsync(userCredentials)
	if err := task.Error(); err != nil {
		log.Fatalf("Failed listing all projections: %v", err)
	}
	for i, projection := range task.Result().([]*projections.ProjectionDetails) {
		fmt.Fprintf(os.Stdout, "%d: %v\n", i, projection)
	}
}

func projectionNameFlag() {
	flag.StringVar(&name, "name", "", "Projection name")
}

func enable() {
	if name == "" {
		commandsUsage()
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if err := manager.EnableAsync(name, userCredentials).Error(); err != nil {
		log.Fatalf("Failed enabling projection %s: %v", name, err)
	}
	log.Println("Success")
}

func disable() {
	if name == "" {
		commandsUsage()
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	if err := manager.DisableAsync(name, userCredentials).Error(); err != nil {
		log.Fatalf("Failed disabling projection %s: %v", name, err)
	}
	log.Println("Success")
}

# Event Store Go Client

A [Go](https://golang.org/) port of the .Net client for [Event Store](https://eventstore.org/).

**This library is not compatible with the new gRPC protocol.**

## Learning

If you want to learn more about EventSourcing/CQRS/EventModeling, you can join the virtual workshop offered by my employer Adaptech Group, see info at [https://www.adaptechgroup.com/virtual-workshop/](https://www.adaptechgroup.com/virtual-workshop/).

## License

Ported code is released under the [MIT](https://github.com/jdextraze/go-gesclient/blob/master/LICENSE) license.

Original code is released under the [Event Store License](https://github.com/EventStore/EventStore/blob/master/LICENSE.md).

## Status

This project is considered ready for production.

### Implemented

* Writing to a stream
* Reading a single event
* Reading a stream forwards
* Reading a stream backwards
* Reading all events forwards
* Reading all events backwards
* Volatile subscriptions
* Persistent subscription
* Deleting stream
* Cluster connection
* Global authentication
* Get/Set stream metadata
* Set system settings
* Transaction
* SSL connection
* Projections Management

### Missing

* 64bit event number support

### Need Improvements

* Documentation
* Unit and integration tests
* Benchmarks

### Known issues

* None at the moment

## Getting started

### Requirements

- Go 1.4+

### Install

1. Install using `go get -u github.com/jdextraze/go-gesclient`.
2. Install [govendor](https://github.com/kardianos/govendor) with `go get -u github.com/kardianos/govendor`.
3. Install vendor packages with `govendor sync`

### Optional Tools

* [Robo](https://github.com/tj/robo) (`go get github.com/tj/robo`)
* [Docker](https://www.docker.com/get-docker)

### Running EventStore on local machine

See https://developers.eventstore.com/server/5.0.9/server/installation/

### Running EventStore with docker

`docker run --rm --name eventstore -d -p 1113:1113 -p 2113:2113 eventstore/eventstore:release-5.0.8`

or if you installed robo

* Start a node: `robo start_es 5.0.8`
* Start a node with SSL: `robo start_es 5.0.8 ssl`
* Stop the node: `robo stop_es`
* Start a cluster: `robo start_es_cluster 5.0.8`
* Stop the cluster: `robo stop_es_cluster`

### Examples

For examples, look into the `examples` folder. All examples connect to `tcp://localhost:1113` by default.

#### Building

To build all examples, use `robo build`.

#### Using SSL

To use SSL in examples, add `-ssl-host localhost -ssl-skip-verify` to the command line.

You can also add the example certificates to your local store by running `robo install_certificate`.
Then you don't need to add `-ssl-skip-verify`.

## Other languages client

* [.Net](https://github.com/EventStore/EventStore) (Official)
* [Java / Scala](https://github.com/EventStore/EventStore.JVM) (Official)
* [Node.JS](https://github.com/nicdex/node-eventstore-client)

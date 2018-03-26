# Event Store Go Client

A [Go](https://golang.org/) port of the .Net client for [Event Store](https://eventstore.org/).

## License

Ported code is released under the [MIT](https://github.com/jdextraze/go-gesclient/blob/master/LICENSE) license.

Original code is released under the [Event Store License](https://github.com/EventStore/EventStore/blob/master/LICENSE.md).

## Status

This project is considered at a beta stage. *Use in production at your own risk.*

### Warning

API is still under development and could change.

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

### Missing

* Projections Management
* 64bit event number support

### Need Improvements

* Documentation
* Unit and integration tests
* Benchmarks

### Known issues

* SSL connection are not working

## Getting started

### Requirements

- Go 1.4+

### Install

Install using `go get github.com/jdextraze/go-gesclient`.

### Optional Tools

* [Robo](https://github.com/tj/robo) (`go get github.com/tj/robo`)
* [Docker](https://www.docker.com/get-docker)
* [Dep](https://github.com/golang/dep)

### Running EventStore on local machine

See https://eventstore.org/docs/introduction/4.1.0/

### Running EventStore with docker

`docker run --rm --name eventstore -d -p 1113:1113 -p 2113:2113 eventstore/eventstore:release-4.1.0`

or if you installed robo

`robo start_es 4.1.0`

## Examples

For examples, look into `examples`. You will need an instance of event store to be running to try it.
I suggest using [Docker](https://docker.com/) with [Event Store Docker Container](https://hub.docker.com/r/eventstore/eventstore/).

## Other languages client

* [.Net](https://github.com/EventStore/EventStore) (Official)
* [Java / Scala](https://github.com/EventStore/EventStore.JVM) (Official)
* [Node.JS](https://github.com/nicdex/node-eventstore-client)

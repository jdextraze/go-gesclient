# Event Store Go Client

A [Go](https://golang.org/) port of the .Net client for [Event Store](https://geteventstore.com/).

## Status

This project is considered at an alpha stage and shouldn't be used in production.

## License

MIT. See [LICENSE](https://github.com/jdextraze/go-gesclient/blob/master/LICENSE).

## Examples

For examples, look into `examples`. You will need an instance of event store to be running to try it.
I suggest using [Docker](https://docker.com/) with [Event Store Docker Container](https://hub.docker.com/r/eventstore/eventstore/).

## Implemented

* Writing to a stream
* Reading a single event
* Reading a stream forwards
* Reading a stream backwards
* Reading all events forwards
* Reading all events backwards
* Volatile subscriptions
* Deleting stream

## TODO

* Persistent subscription
* Complete unit and integration tests
* Transaction
* Scavenging database
* Global authentication
* Cluster connection

# Event Store Go Client

A [Go](https://golang.org/) client for [Event Store](https://geteventstore.com/).

## Status

This project is considered at an alpha stage and shouldn't be used in production.

## Example

For an example, look into `example`. You will need an instance of event store to be running to try it.
I suggest using [Docker](https://docker.com/) with [Event Store Docker Container](https://hub.docker.com/r/eventstore/eventstore/).

## Implemented

* Writing to a stream (No transaction)
* Reading a specific stream forwards
* Volatile subscriptions

## TODO

* Unit and integration tests

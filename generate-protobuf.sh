#!/bin/bash

protoc -I$GOPATH/src --go_out=$GOPATH/src $GOPATH/src/github.com/jdextraze/go-gesclient/protobuf/ClientMessageDtos.proto

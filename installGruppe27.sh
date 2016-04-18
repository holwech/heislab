#!/bin/bash

mkdir $HOME/Desktop/gospace
export GOPATH=$HOME/Desktop/gospace
go get github.com/holwech/heislab
go get github.com/satori/go.uuid
go run $GOPATH/src/github.com/holwech/heislab/main.go

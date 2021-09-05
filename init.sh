#!/bin/sh

# init

if [ ! -f go.mod ]; then
	go mod init
fi

go mod tidy
go mod vendor



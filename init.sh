#!/bin/sh

# init

if [ ! -f go.mod ]; then
	go mod init
fi

go mod tidy

if [ ! -d vendor ]; then
	go mod vendor
fi


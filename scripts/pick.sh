#!/bin/bash

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

curPath=`pwd`
rootPath=$(dirname "$curPath")
export CGO_ENABLED=0 GOOS=linux GOARCH=amd64
cd $rootPath && go build imail.go
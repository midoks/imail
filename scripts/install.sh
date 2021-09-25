#!/bin/bash

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

_os=`uname`
_os=$(echo $_os | tr '[A-Z]' '[a-z]')
_ver="0.0.4"

which go
if [ "0" != $? ]; then
	echo "missing go running environment!"
	exit
fi

echo "package main
import (
	\"fmt\"
	\"runtime\"
)
func main() { fmt.Println(runtime.GOARCH) }" > /tmp/t.go

_go=`cd /tmp && go run t.go`


url="https://github.com/midoks/imail/releases/download/0.0.4/imail-{$_os}-{$_go}.tar.gz"

echo $_os
echo $_go

TAGRT_DIR=/usr/local/imail
mkdir -p $TAGRT_DIR
cd $TAGRT_DIR

wget -o "imail-{$_os}-{$_go}.tar.gz" $url





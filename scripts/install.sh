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

_arch=`cd /tmp && go run t.go`


url="https://github.com/midoks/imail/releases/download/0.0.4/imail-$_os-$_arch.tar.gz"

echo $_os
echo $_arch

TAGRT_DIR=/usr/local/imail
mkdir -p $TAGRT_DIR
cd $TAGRT_DIR

wget -O "imail-$_os-$_arch.tar.gz" $url
tar zxvf "imail-$_os-$_arch.tar.gz"
rm -rf "imail-$_os-$_arch.tar.gz"

cd $TAGRT_DIR/scripts
sh make.sh

systemctl daemon-reload
service imail restart

cd $TAGRT_DIR && ./imail -v



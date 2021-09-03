#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin



cd /usr/local

git clone https://github.com/midoks/imail

cd imail



go mod tidy

if [ ! -d vendor ]; then
	go mod vendor
fi

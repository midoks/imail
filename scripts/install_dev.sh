#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin


cd /usr/local


if [ ! -d /usr/local/imail ]; then
	git clone https://github.com/midoks/imail
	cd imail
else
	cd imail
	git pull https://github.com/midoks/imail
fi

go mod tidy

if [ ! -d vendor ]; then
	go mod vendor
fi

go build ./

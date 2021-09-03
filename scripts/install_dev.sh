#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin


TAGRT_DIR=/usr/local
cd $TAGRT_DIR


if [ ! -d $TAGRT_DIR/imail ]; then
	git clone https://github.com/midoks/imail
	cd $TAGRT_DIR/imail
else
	cd $TAGRT_DIR/imail
	git pull https://github.com/midoks/imail
fi

go mod tidy

if [ ! -d vendor ]; then
	go mod vendor
fi

if [ ! -f imail ];then
	rm -rf imail
	go build ./
fi

cd $TAGRT_DIR/imail/scripts

sh make.sh

#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin


TAGRT_DIR=/usr/local/imail_dev
mkdir -p $TAGRT_DIR
cd $TAGRT_DIR


if [ ! -d $TAGRT_DIR/imail ]; then
	git clone https://github.com/midoks/imail
	cd $TAGRT_DIR/imail
else
	cd $TAGRT_DIR/imail
	git pull https://github.com/midoks/imail
fi

go mod tidy
go mod vendor


rm -rf imail
go build ./


cd $TAGRT_DIR/imail/scripts

sh make.sh

systemctl daemon-reload

service imail restart

cd $TAGRT_DIR/imail && ./imail -v



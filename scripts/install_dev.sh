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
	go build ./
fi


_os=`uname`
_path=`pwd`
_dir=`dirname $_path`

IMAIL_PATH=$_dir/imail

echo $IMAIL_PATH

sed "s:{APP_PATH}/:${IMAIL_PATH}:g" $TAGRT_DIR/imail/scripts/init.d/imail.tpl > $TAGRT_DIR/imail/scripts/init.d/imail
chmod +x $TAGRT_DIR/imail/scripts/init.d/imail


if [ -d /etc/init.d ];then
	cp $TAGRT_DIR/imail/scripts/init.d/imail /etc/init.d/imail
	chmod +x /etc/init.d/imail
fi

echo `dirname $_path`

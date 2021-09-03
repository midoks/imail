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

sed "s:{APP_PATH}:${_dir}:g" $_dir/script/init.d/imail.tpl > $_dir/script/init.d/imail
chmod +x $_dir/script/init.d/imail


if [ -d /etc/init.d ];then
	cp $_dir/script/init.d/imail /etc/init.d/imail
	chmod +x /etc/init.d/imail
fi

echo `dirname $_path`

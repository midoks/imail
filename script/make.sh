#!/bin/bash

_os=`uname`
_path=`pwd`
_dir=`dirname $_path`

if [ "$_os" == "Darwin" ] ; then
	echo "macosx not need!"
else
	# go build imail.go
	echo $_dir

fi


if [ ! -d /etc/init.d/imail ];then
	cat $_dir/script/init.d/imail.tpl >  /etc/init.d/imail
	chmod +x /etc/init.d/imail
fi

echo `dirname $_path`
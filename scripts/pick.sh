#!/bin/bash

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

curPath=`pwd`
rootPath=$(dirname "$curPath")

PACK_NAME=imail

# go tool dist list
mkdir -p $rootPath/tmp/build
mkdir -p $rootPath/tmp/package


build_app(){
	echo "build_app" $1 $2

	echo "export CGO_ENABLED=0 GOOS=$1 GOARCH=$2"
	echo "cd $rootPath && go build imail.go"

	export CGO_ENABLED=0 GOOS=$1 GOARCH=$2
	# export CGO_ENABLED=0 GOOS=linux GOARCH=amd64

	if [ $1 == "windows" ];then
		cd $rootPath/internal/win32 && go build imail.go
	else
		cd $rootPath && go build imail.go
		# cd $rootPath && go build imail.go && /usr/local/bin/strip imail
	fi
	
	
	cp -r $rootPath/conf $rootPath/tmp/build
	cd $rootPath/tmp/build/conf/dkim && rm -rf ./* && echo "#dkim" > ./README.md
	cd $rootPath/tmp/build/conf/ && rm -rf ./app.conf
	mkdir -p $rootPath/tmp/build/data && echo "#data" > $rootPath/tmp/build/data/README.md


	cp -r $rootPath/scripts $rootPath/tmp/build
	cp -r $rootPath/logs $rootPath/tmp/build
	cp -r $rootPath/hook $rootPath/tmp/build

	cd $rootPath/tmp/build/logs && rm -rf ./*.log
	cd $rootPath/tmp/build && xattr -c * && rm -rf ./*/.DS_Store && rm -rf ./*/*/.DS_Store


	if [ $1 == "windows" ];then
		cp $rootPath/internal/win32/imail.exe $rootPath/tmp/build
		rm -rf $rootPath/tmp/build/imail
	else
		rm -rf $rootPath/imail.exe
		rm -rf $rootPath/tmp/build/imail.exe
		cp $rootPath/imail $rootPath/tmp/build
	fi

	# zip
	cd $rootPath/tmp/build && zip -r -q -o ${PACK_NAME}-$1-$2.zip  ./ && mv ${PACK_NAME}-$1-$2.zip $rootPath/tmp/package
	# tar.gz
	cd $rootPath/tmp/build && tar -zcvf ${PACK_NAME}-$1-$2.tar.gz ./ && mv ${PACK_NAME}-$1-$2.tar.gz $rootPath/tmp/package
	# bz
	#cd $rootPath/tmp/build && tar -jcvf ${PACK_NAME}-$1-$2.tar.bz2 ./ && mv ${PACK_NAME}-$1-$2.tar.bz2 $rootPath/tmp/package
}

golist=`go tool dist list`

echo $golist

build_app linux amd64
build_app linux 386
build_app linux armv7
build_app darwin amd64
build_app windows 386
build_app windows amd64


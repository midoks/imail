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
	cd $rootPath && go build imail.go
	# cd $rootPath && go build imail.go && /usr/local/bin/strip imail
	
	cp -r $rootPath/conf $rootPath/tmp/build
	cp -r $rootPath/scripts $rootPath/tmp/build
	cp -r $rootPath/logs $rootPath/tmp/build
	cp -r $rootPath/hook $rootPath/tmp/build

	cd $rootPath/tmp/build/logs && rm -rf ./*.log

	if [ $1 == "windows" ];then
		cp $rootPath/imail.exe $rootPath/tmp/build
		rm -rf $rootPath/tmp/build/imail
	else
		rm -rf $rootPath/imail.exe
		rm -rf $rootPath/tmp/build/imail.exe
		cp $rootPath/imail $rootPath/tmp/build
	fi

	cd $rootPath/tmp/build && zip -r -q -o ${PACK_NAME}-$1-$2.zip  ./ && mv ${PACK_NAME}-$1-$2.zip $rootPath/tmp/package
}

golist=`go tool dist list`

echo $golist

build_app linux amd64
# build_app linux 386
# build_app darwin amd64
# build_app windows amd64


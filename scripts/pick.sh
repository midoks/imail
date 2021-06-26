#!/bin/bash

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

curPath=`pwd`
rootPath=$(dirname "$curPath")

PACK_NAME=imail


mkdir -p $curPath/tmp
mkdir -p $curPath/package


export CGO_ENABLED=0 GOOS=linux GOARCH=amd64
cd $rootPath && go build imail.go

cp $rootPath/imail $curPath/tmp
cp -r $rootPath/conf $curPath/tmp


cd $curPath/tmp && zip -r -q -o ${PACK_NAME}-linux-amd64.zip  ./ && mv ${PACK_NAME}-linux-amd64.zip $curPath/package



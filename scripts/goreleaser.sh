#!/bin/bash

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

curPath=`pwd`
rootPath=$(dirname "$curPath")

cd  $rootPath/conf/dkim && rm -rf ./* && echo "#dkim" > ./README.md
cd  $rootPath/conf/ && rm -rf ./app.conf

cd $rootPath

goreleaser --snapshot --skip-publish --rm-dist


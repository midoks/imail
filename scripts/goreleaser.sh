#!/bin/bash

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

curPath=`pwd`
rootPath=$(dirname "$curPath")

cd  $rootPath/conf/dkim && rm -rf ./* && echo "#dkim" > ./README.md
cd  $rootPath/conf/ && rm -rf ./app.conf

cd $rootPath

# brew install FiloSottile/musl-cross/musl-cross
# brew install mingw-w64

#goreleaser --snapshot --skip-publish --rm-dist

#git tag -a 0.0.5 -m "release 0.0.5"
#goreleaser release --rm-dist

#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

# https://github.com/FiloSottile/homebrew-musl-cross
# brew install FiloSottile/musl-cross/musl-cross --without-x86_64 --with-i486 --with-aarch64 --with-arm

# brew install mingw-w64
# sudo port install mingw-w64

VERSION=0.0.12
curPath=`pwd`
rootPath=$(dirname "$curPath")

PACK_NAME=imail

# go tool dist list
mkdir -p $rootPath/tmp/build
mkdir -p $rootPath/tmp/package

source ~/.bash_profile

cd $rootPath
LDFLAGS="-X \"github.com/midoks/imail/internal/conf.BuildTime=$(date -u '+%Y-%m-%d %I:%M:%S %Z')\""
LDFLAGS="${LDFLAGS} -X \"github.com/midoks/imail/internal/conf.BuildCommit=$(git rev-parse HEAD)\""


echo $LDFLAGS
build_app(){

	if [ -f $rootPath/tmp/build/imail ]; then
		rm -rf $rootPath/tmp/build/imail
		rm -rf $rootPath/imail
	fi

	if [ -f $rootPath/tmp/build/imail.exe ]; then
		rm -rf $rootPath/tmp/build/imail.exe
		rm -rf $rootPath/imail.exe
	fi

	echo "build_app" $1 $2

	echo "export CGO_ENABLED=1 GOOS=$1 GOARCH=$2"
	echo "cd $rootPath && go build imail.go"

	# export CGO_ENABLED=1 GOOS=linux GOARCH=amd64

	if [ $1 != "darwin" ];then
		export CGO_ENABLED=1 GOOS=$1 GOARCH=$2
		export CGO_LDFLAGS="-static"
	fi

	cd $rootPath && go generate internal/assets/conf/conf.go
	cd $rootPath && go generate internal/assets/templates/templates.go
	cd $rootPath && go generate internal/assets/public/public.go

	cd $rootPath && go generate internal/assets/conf/conf.go
	cd $rootPath && go generate internal/assets/templates/templates.go
	cd $rootPath && go generate internal/assets/public/public.go


	if [ $1 == "windows" ];then
		
		if [ $2 == "amd64" ]; then
			export CC=x86_64-w64-mingw32-gcc
			export CXX=x86_64-w64-mingw32-g++
		else
			export CC=i686-w64-mingw32-gcc
			export CXX=i686-w64-mingw32-g++
		fi

		cd $rootPath && go build -o imail.exe -ldflags "${LDFLAGS}" imail.go

		# -ldflags="-s -w"
		# cd $rootPath && go build imail.go && /usr/local/bin/strip imail
	fi

	if [ $1 == "linux" ]; then
		export CC=x86_64-linux-musl-gcc
		if [ $2 == "amd64" ]; then
			export CC=x86_64-linux-musl-gcc

		fi

		if [ $2 == "386" ]; then
			export CC=i486-linux-musl-gcc
		fi

		if [ $2 == "arm64" ]; then
			export CC=aarch64-linux-musl-gcc
		fi

		if [ $2 == "arm" ]; then
			export CC=arm-linux-musleabi-gcc
		fi

		cd $rootPath && go build -ldflags "${LDFLAGS}"  imail.go 
	fi

	if [ $1 == "darwin" ]; then
		echo "cd $rootPath && go build -v -ldflags '${LDFLAGS}'"
		cd $rootPath && go build -v -ldflags "${LDFLAGS}"
		
		cp $rootPath/imail $rootPath/tmp/build
	fi
	

	cp -r $rootPath/scripts $rootPath/tmp/build
	cp -r $rootPath/LICENSE $rootPath/tmp/build
	cp -r $rootPath/README.md $rootPath/tmp/build

	cd $rootPath/tmp/build && xattr -c * && rm -rf ./*/.DS_Store && rm -rf ./*/*/.DS_Store


	if [ $1 == "windows" ];then
		cp $rootPath/imail.exe $rootPath/tmp/build
	else
		cp $rootPath/imail $rootPath/tmp/build
	fi

	# zip
	#cd $rootPath/tmp/build && zip -r -q -o ${PACK_NAME}_${VERSION}_$1_$2.zip  ./ && mv ${PACK_NAME}_${VERSION}_$1_$2.zip $rootPath/tmp/package
	# tar.gz
	cd $rootPath/tmp/build && tar -zcvf ${PACK_NAME}_${VERSION}_$1_$2.tar.gz ./ && mv ${PACK_NAME}_${VERSION}_$1_$2.tar.gz $rootPath/tmp/package
	# bz
	#cd $rootPath/tmp/build && tar -jcvf ${PACK_NAME}_${VERSION}_$1_$2.tar.bz2 ./ && mv ${PACK_NAME}_${VERSION}_$1_$2.tar.bz2 $rootPath/tmp/package

}

golist=`go tool dist list`
echo $golist

build_app linux amd64
build_app linux 386
build_app linux arm64
build_app linux arm
build_app darwin amd64
build_app windows 386
build_app windows amd64


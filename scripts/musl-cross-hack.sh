#!/bin/bash

TMP_DIR=/usr/local/Cellar/musl-cross3/0.0.1/.brew

mkdir -p $TMP_DIR

if [ ! -f $TMP_DIR/musl-cross3.rb ];then
	wget -O $TMP_DIR/musl-cross3.rb  https://raw.githubusercontent.com/FiloSottile/homebrew-musl-cross/master/musl-cross.rb
	cd $TMP_DIR
	sed -i '_bak' 's/MuslCross/MuslCross3/g' musl-cross3.rb
fi


brew install musl-cross3.rb
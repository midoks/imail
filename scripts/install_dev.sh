#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin


TAGRT_DIR=/usr/local/imail_dev
mkdir -p $TAGRT_DIR
cd $TAGRT_DIR

export GIT_COMMIT=$(git rev-parse HEAD)
export BUILD_TIME=$(date -u '+%Y-%m-%d %I:%M:%S %Z')


if [ ! -d $TAGRT_DIR/imail ]; then
	git clone https://github.com/midoks/imail
	cd $TAGRT_DIR/imail
else
	cd $TAGRT_DIR/imail
	git pull https://github.com/midoks/imail
fi

go mod tidy
go mod vendor


rm -rf imail
go build  -ldflags "-X \"github.com/midoks/imail/internal/conf.BuildTime=${BUILD_TIME}\" -X \"github.com/midoks/imail/internal/conf.BuildCommit=${GIT_COMMIT}\"" ./


cd $TAGRT_DIR/imail/scripts

sh make.sh

systemctl daemon-reload

service imail restart

cd $TAGRT_DIR/imail && ./imail -v



#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin


# curl -fsSL  https://raw.githubusercontent.com/midoks/imail/master/scripts/install_dev.sh | sh

# Linux 手动安装
# wget https://go.dev/dl/go1.19.1.linux-amd64.tar.gz
# sudo tar -C /usr/local -xzf go1.19.1.linux-amd64.tar.gz
# sudo ln -s /usr/local/go/bin/* /usr/bin/


TAGRT_DIR=/usr/local/imail_dev
mkdir -p $TAGRT_DIR
cd $TAGRT_DIR

export GIT_COMMIT=$(git rev-parse HEAD)
export BUILD_TIME=$(date -u '+%Y-%m-%d %I:%M:%S %Z')

go install github.com/midoks/zzz@latest

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


rm -rf /usr/local/imail_dev/imail/custom
rm -rf /usr/local/imail_dev/imail/data

service imail restart

cd $TAGRT_DIR/imail && ./imail -v

# Debug Now
export PATH=$PATH:/root/go/bin
export GOPATH=/root/go
service imail stop
cd /usr/local/imail_dev/imail && bash zzz.sh


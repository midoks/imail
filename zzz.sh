#!/bin/sh
curPath=`pwd`

go env -w GOSUMDB=off

export GIT_COMMIT=$(git rev-parse HEAD)
export BUILD_TIME=$(date -u '+%Y-%m-%d %I:%M:%S %Z')

# zzz run -ldflags "-w -s -X \"github.com/midoks/imail/internal/conf.BuildCommit=$(git rev-parse HEAD)\" -X \"github.com/midoks/imail/internal/conf.BuildTime=$(date -u '+%Y-%m-%d %I:%M:%S %Z')\""

echo zzz run -ldflags "-X \"github.com/midoks/imail/internal/conf.BuildTime=${BUILD_TIME}\" -X \"github.com/midoks/imail/internal/conf.BuildCommit=${GIT_COMMIT}\""
zzz run --ld "-X \"github.com/midoks/imail/internal/conf.BuildTime=${BUILD_TIME}\" -X \"github.com/midoks/imail/internal/conf.BuildCommit=${GIT_COMMIT}\""
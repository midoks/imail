#!/bin/sh
curPath=`pwd`


# init

if [ ! -f go.mod ]; then
	go mod init
fi

go mod tidy
go mod vendor


# test cover
cd $curPath/internal/imap
go test -coverprofile=cov.out -coverpkg ./...
go tool cover -html cov.out -o index.html

cd $curPath/internal/pop3
go test -coverprofile=cov.out -coverpkg ./...
go tool cover -html cov.out -o index.html

cd $curPath/internal/smtpd
go test -coverprofile=cov.out -coverpkg ./...
go tool cover -html cov.out -o index.html


# goreleaser --snapshot --skip-publish --rm-dist

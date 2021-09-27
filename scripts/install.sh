#!/bin/bash

check_go_environment() {
	if test ! -x "$(command -v go)"; then
		printf "\e[1;31mmissing go running environment\e[0m\n"
		exit 1
	fi
}

load_vars() {
	OS=$(uname | tr '[:upper:]' '[:lower:]')

	VERSION=$(get_latest_release "midoks/imail")

	TARGET_DIR="/usr/local/imail"
}

get_latest_release() {
    curl -sL "https://api.github.com/repos/$1/releases/latest" | grep '"tag_name":' | cut -d'"' -f4
}

get_arch() {
	echo "package main
import (
	\"fmt\"
	\"runtime\"
)
func main() { fmt.Println(runtime.GOARCH) }" > /tmp/go_arch.go

	ARCH=$(go run /tmp/go_arch.go)
}

get_download_url() {
	DOWNLOAD_URL="https://github.com/midoks/imail/releases/download/$VERSION/imail_${VERSION}_${OS}_${ARCH}.tar.gz"
}

# download file
download_file() {
    url="${1}"
    destination="${2}"

    printf "Fetching ${url} \n\n"

    if test -x "$(command -v curl)"; then
        code=$(curl --connect-timeout 15 -w '%{http_code}' -L "${url}" -o "${destination}")
    elif test -x "$(command -v wget)"; then
        code=$(wget -t2 -T15 -O "${destination}" --server-response "${url}" 2>&1 | awk '/^  HTTP/{print $2}' | tail -1)
    else
        printf "\e[1;31mNeither curl nor wget was available to perform http requests.\e[0m\n"
        exit 1
    fi

    if [ "${code}" != 200 ]; then
        printf "\e[1;31mRequest failed with code %s\e[0m\n" $code
        exit 1
    else 
	    printf "\n\e[1;33mDownload succeeded\e[0m\n"
    fi
}


main() {
	check_go_environment

	load_vars

	get_arch

	get_download_url

	DOWNLOAD_FILE="$(mktemp).tar.gz"
	download_file $DOWNLOAD_URL $DOWNLOAD_FILE

	if [ ! -d "$TARGET_DIR" ]; then
		mkdir -p "$TARGET_DIR"
	fi

	tar -C "$TARGET_DIR" -zxf $DOWNLOAD_FILE
	rm -rf $DOWNLOAD_FILE

	pushd "$TARGET_DIR/scripts" >/dev/null 2>&1
	bash make.sh

	systemctl daemon-reload
	service imail restart

	cd .. && ./imail -v	
	popd >/dev/null 2>&1
}

main "$@" || exit 1

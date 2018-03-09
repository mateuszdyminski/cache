#!/bin/bash

usage() {
	cat <<EOF
Usage: $(basename $0) <command>

Wrappers around core binaries:
    build                  Builds the cache.
EOF
	exit 1
}

CMD="$1"
shift
case "$CMD" in
	build)
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cache -a -tags netgo .
	;;
	*)
		usage
	;;
esac
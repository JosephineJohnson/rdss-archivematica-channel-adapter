#!/usr/bin/env bash

# This script is used to compile the protocol buffers.

set -o errexit
set -o pipefail
set -o nounset

if ! which protoc > /dev/null 2>&1 ; then
	echo >&2 "Cannot find protoc. Install with \"go get -u github.com/golang/protobuf/protoc-gen-go\""
	echo >&2 "See https://developers.google.com/protocol-buffers/ for more instructions.";
	echo >&2 "Aborting.";
	exit 1;
fi

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly __root="$(cd "$(dirname "${__dir}")" && pwd)"
readonly __gopath="$(cd "$(dirname "${__root}/../../../")" && pwd)"

readonly GOPATH="${GOPATH:-${__gopath}}"

echo "Compiling..."
cd ${__root}

protoc \
	-I/usr/local/include \
	-I. \
	-I${GOPATH}/src \
	--gogo_out=plugins=grpc:${GOPATH}/src/github.com/JiscRDSS/rdss-archivematica-channel-adapter \
	publisher/pb/rpc.proto

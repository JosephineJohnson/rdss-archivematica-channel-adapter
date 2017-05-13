#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly __root="$(cd "$(dirname "${__dir}")" && pwd)"

readonly workdir="$(mktemp -d)"
readonly profile="$workdir/cover.out"
readonly mode=count

generate_cover_data() {
	for pkg in "$@"; do
		f="$workdir/$(echo $pkg | tr / -).cover"
		go test -covermode="$mode" -coverprofile="$f" "$pkg"
	done

	echo "mode: $mode" >"$profile"
	grep -h -v "^mode:" "$workdir"/*.cover >> "$profile"
}

show_cover_report() {
	go tool cover -${1}="$profile"
}

cleanup() {
	rm -rf ${workdir};
}

cd ${__root}
generate_cover_data $(go list ./... | grep -v '/vendor/')
show_cover_report func

if [ -n "${1+set}" ]; then
	case "$1" in
	"")
		;;
	--html)
		show_cover_report html ;;
	*)
		echo >&2 "error: invalid option: $1"; exit 1 ;;
	esac
fi

cleanup

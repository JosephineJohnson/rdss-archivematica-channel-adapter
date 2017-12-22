#!/usr/bin/env bash

# This script is used to download the JSON Schema files.
# It should not be needed once the repository is open sourced.

set -o errexit
set -o pipefail
set -o nounset

readonly __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly __token="${1:-}"
readonly __owner="JiscRDSS"
readonly __repo="rdss-message-api-docs"

function error() {
	echo "ERR"
	echo "ERR ${2}"
	echo "ERR"
	exit ${1}
}

function download_file() {
	file="${1}"
	path="${__dir}/schemas/${file}"
	dir="$(dirname "${path}")"
	url="https://api.github.com/repos/${__owner}/${__repo}/contents/${file}"

	echo "Downloading ${file}..."
	mkdir -p ${dir}
	curl \
		--silent \
		--fail \
		--header "Authorization: token ${__token}" \
		--header "Accept: application/vnd.github.v3.raw" \
		--location "${url}" \
		> ${path} || true
	rc=$?
	if [ $rc != 0 ]; then
		error 1 "The schema file could not be downloaded. Have you checked that the token works?"
	fi
}

if [ -z "${__token}" ]; then
    error 0 "GitHub API token argument is missing or empty, skipping."
fi

download_file schemas/enumeration.json
download_file schemas/intellectual_asset.json
download_file schemas/material_asset.json
download_file schemas/research_object.json
download_file schemas/types.json
download_file messages/body/metadata/create/request.json
download_file messages/body/metadata/create/request_schema.json
download_file messages/body/metadata/delete/request.json
download_file messages/body/metadata/delete/request_schema.json
download_file messages/body/metadata/read/request.json
download_file messages/body/metadata/read/request_schema.json
download_file messages/body/metadata/read/response.json
download_file messages/body/metadata/read/response_schema.json
download_file messages/body/metadata/update/request.json
download_file messages/body/metadata/update/request_schema.json
download_file messages/body/vocabulary/patch/request.json
download_file messages/body/vocabulary/patch/request_schema.json
download_file messages/body/vocabulary/read/request.json
download_file messages/body/vocabulary/read/request_schema.json
download_file messages/body/vocabulary/read/response.json
download_file messages/body/vocabulary/read/response_schema.json
download_file messages/header/header.json
download_file messages/header/header_schema.json
download_file messages/message.json
download_file messages/message_schema.json

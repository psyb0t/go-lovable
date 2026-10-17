#!/bin/bash
# Downloads the Lovable OpenAPI spec into api/. The spec is not committed;
# `go generate` fetches it as its first step.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

readonly SPEC_URL="https://api.lovable.dev/v1/openapi.yaml"
readonly DOWNLOAD_TIMEOUT_SECONDS=60
readonly SPEC_NAME="openapi.yaml"
SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

main() {
	mkdir -p "$SPEC_DIR"
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.${SPEC_NAME}.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp now, the local is gone when EXIT fires
	trap "rm -f -- '${tmp}'" EXIT

	log INFO "downloading spec=${SPEC_NAME}"
	wget -q \
		-T "$DOWNLOAD_TIMEOUT_SECONDS" \
		-O "$tmp" \
		"$SPEC_URL"

	# The spec is YAML and the image has no YAML parser, so check for the
	# top-level openapi key instead of parsing it.
	if ! grep -qE '^openapi: ' "$tmp"; then
		log ERROR "downloaded file is not an OpenAPI document spec=${SPEC_NAME}"
		exit 1
	fi

	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/${SPEC_NAME}"
	log INFO "saved spec=${SPEC_NAME} openapi=$(grep -m1 -E '^openapi: ' "${SPEC_DIR}/${SPEC_NAME}" | cut -d' ' -f2)"
}

main "$@"

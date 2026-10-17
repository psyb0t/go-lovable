#!/bin/bash
# Converts OpenAPI 3.1 to 3.0.3 with the local YAML-aware patch-spec tool.
# Nullable type lists become a concrete type plus nullable: true. Null enum
# entries are removed because oapi-codegen emits them as the string "<nil>";
# nullability remains on the schema. Schema examples arrays become one example
# because OpenAPI 3.0 accepts only the singular keyword there. Operations,
# references and example payloads are preserved.
# Documentation uses "directory" for the provider's directory permission grants.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR
readonly SPEC_OUT="${SPEC_DIR}/openapi.fixed.yaml"

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
	cd "${SPEC_DIR}/.."
	go tool patch-spec
	log INFO "converted spec out=${SPEC_OUT##*/}"
}

main "$@"

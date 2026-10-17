# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.0, 2026-10-17

- Added `lovable`, a REST client generated with oapi-codegen from Lovable's OpenAPI 3.1.0 document, spec version `2026-09-11`, at https://api.lovable.dev/v1/openapi.yaml. It covers all 24 operations across 20 paths: projects, publishing and deployment status, preview embed URLs, analytics, personal data findings, security scans, workspaces, groups and members.
- Number fields use `float64`. Required nullable fields preserve JSON null with `nullable.Nullable[T]`.
- `make generate` downloads the spec, converts nullable type lists, null enum entries and schema example arrays into OpenAPI 3.0.3 form, and regenerates the client. The spec is not committed.
- Added HTTP contract tests, the compiled README example, vendored dependencies, development tooling, CI workflows and the `go-lovable` agent skill.

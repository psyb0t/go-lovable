# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.4, 2026-10-28

- Updated Semgrep in the development image to 1.182.1 and refreshed its locked dependencies. certifi 2026.10.23 drops the emSign root CAs whose intermediate keys were compromised (GHSA-6q48-6c7w-9gw2).
- Dependabot no longer opens pull requests for `mcp` or the OpenTelemetry packages in `requirements-semgrep.txt`. Semgrep pins their versions itself, so bumping one of them on its own cannot install.
- Updated kin-openapi, which the code generator uses, from 0.142.0 to 0.149.0. The newer version fixes several panics on malformed specs reported in GitHub security advisories.
- The security scan merges its two gitleaks runs into one SARIF run. GitHub code scanning rejects a report with two runs from the same tool, so no scan results had been uploaded.

## v0.1.3, 2026-10-27

- The development image is based on Debian (`golang:1.26.10-trixie`) instead of Alpine. Alpine keeps only the latest build of each package, so the exact pins broke the image build whenever a package was updated. System packages now come from the snapshot.debian.org date the base image was built from, which keeps them fixed for a given base image digest. jq 1.8.2, shellcheck 0.11.0 and shfmt 3.13.1 keep their versions, and Semgrep runs in its own virtual environment.
- Raised the minimum Go version in `go.mod` from 1.26.9 to 1.26.10, which fixes the `crypto/x509` vulnerability GO-2026-6809 found by govulncheck.

## v0.1.2, 2026-10-20

- Documentation only. The `WorkspaceInsightsProject.MatchedReason` comment no longer says the value can be missing from `review_priority_explanation`.
- The `ListSecurityInsightProjectsParams.FindingTypeId` comment now says finding filters use the same category rules as `project_count`, and that a project matching a more specific related type is excluded from the broader type.

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.0, 2026-10-17

- Added `lovable`, a REST client generated with oapi-codegen from Lovable's OpenAPI 3.1.0 document, spec version `2026-09-11`, at https://api.lovable.dev/v1/openapi.yaml. It covers all 24 operations across 20 paths: projects, publishing and deployment status, preview embed URLs, analytics, personal data findings, security scans, workspaces, groups and members.
- Number fields use `float64`. Required nullable fields preserve JSON null with `nullable.Nullable[T]`.
- `make generate` downloads the spec, converts nullable type lists, null enum entries and schema example arrays into OpenAPI 3.0.3 form, and regenerates the client. The spec is not committed.
- Added HTTP contract tests, the compiled README example, vendored dependencies, development tooling, CI workflows and the `go-lovable` agent skill.

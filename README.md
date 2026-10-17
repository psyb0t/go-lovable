# go-lovable

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-lovable.svg)](https://pkg.go.dev/github.com/psyb0t/go-lovable)
[![CI](https://github.com/psyb0t/go-lovable/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-lovable/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-lovable/badges/version.svg)](https://github.com/psyb0t/go-lovable/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-lovable/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-lovable/badges/importers.svg)](https://github.com/psyb0t/go-lovable/blob/badges/importers.md)

Unofficial Go client for the [Lovable API](https://docs.lovable.dev/integrations/api-ref-intro), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the OpenAPI document Lovable publishes. I am not affiliated with Lovable.

The package is `lovable`, in the repository root. Everything in `client.gen.go` is generated. Don't edit it by hand, regenerate it.

## Install

```sh
go get github.com/psyb0t/go-lovable
```

## Usage

The client covers projects, publishing, deployments, preview embed URLs, analytics, personal data findings, security scans, workspaces, groups and members. Every endpoint in the spec has a typed method on `ClientWithResponses`. Responses decode into `JSON200` and the other status fields.

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-lovable"
)

func main() {
	auth := lovable.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Lovable-API-Key", os.Getenv("LOVABLE_API_KEY"))

			return nil
		},
	)

	client, err := lovable.NewClientWithResponses(
		"https://api.lovable.dev",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.ListWorkspacesWithResponse(
		context.Background(), &lovable.ListWorkspacesParams{},
	)
	if err != nil {
		panic(err)
	}

	if res.JSON200 == nil {
		panic(res.Status())
	}

	_ = res.JSON200.Data
}
```

The same code is in `example_test.go`, so `make test` checks that it compiles.

Required nullable fields use `nullable.Nullable[T]`: `IsNull()` checks JSON null, and `Get()` returns a value and an error. Optional request fields are pointers. HTTP errors do not produce a Go error by themselves: check `StatusCode()`, the typed status field and `Body`. Use `WithHTTPClient` to set timeouts.

## Auth

Create a key at `lovable.dev/settings/api-keys`, export it as `LOVABLE_API_KEY`, and set `Lovable-API-Key` in a request editor as above. Keys belong to one workspace. Scopes are `projects:read`, `projects:write`, `workspaces:read` and `workspaces:write`; write scopes include their matching read scope. Another workspace's resources return 404, and insufficient scopes return 403.

The spec also accepts bearer authentication. Replace the editor with:

```go
auth := lovable.WithRequestEditorFn(
	func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+os.Getenv("LOVABLE_API_KEY"))

		return nil
	},
)
```

The base URL is `https://api.lovable.dev`. Generated paths contain `/v1`, so leave it out of the base URL. Set the optional `LovableVersion` operation parameter or `Lovable-Version` header to `2026-09-11` to select the spec's stable version explicitly. List operations use `limit` and an opaque `cursor`. Pass `Pagination.NextCursor` as `Cursor` until `Pagination.HasMore` is false, keeping the same filters. `ListProjects` requires a `WorkspaceId` slice. The client does not paginate or retry automatically. Honor `Retry-After` on 429 responses. Publishing returns 202; poll `GetDeploymentWithResponse` with backoff and a bounded timeout.

## Regenerating

The spec is not committed. `make generate` downloads the current OpenAPI document into the gitignored `api/` directory and regenerates the client:

```sh
make generate
```

`gen.go` declares the steps, so `make generate` is just `go generate ./...`:

- `scripts/fetch-spec.sh` downloads `https://api.lovable.dev/v1/openapi.yaml` to `api/openapi.yaml`.
- `scripts/fix-nullable-types.sh` runs the local YAML-aware `patch-spec` Go tool, writing `api/openapi.fixed.yaml`. It changes OpenAPI 3.1 nullable type lists into a concrete type plus `nullable: true`, removes null enum entries that the generator otherwise emits as the string `"<nil>"`, converts schema example arrays into a singular example and sets the input version to 3.0.3. All operations and example payloads are preserved. Generated descriptions use the repository's directory terminology.
- oapi-codegen runs with `config.yaml`. Numbers map to `float64`, and nullable fields retain JSON null through `nullable.Nullable[T]`.

`.github/workflows/regen.yml` regenerates daily. Changed clients go through `make lint`, `make test` and `make sec`, up to three repair attempts if needed, and a review. A passing change gets a CHANGELOG entry and a release tag.

## Agent integrations

The [skill](.agents/skills/go-lovable) teaches a coding agent how to use go-lovable when it writes Go. It works in any agent that reads `.agents/skills/`, and installs natively in the clients below.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-lovable@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-lovable@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

Tag pushes publish the skill to ClawHub:

```bash
openclaw skills install @psyb0t/go-lovable
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs only Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the spec and regenerate
```

Run `make dep` after regeneration when imports change. `vendor/` is part of the source distribution. Security scans cover Git history when available; source archives have no history to scan.

## License

[MIT](LICENSE). The generated code comes from Lovable's published API specification. The specification itself is not redistributed here. It names [Lovable's terms](https://lovable.dev/terms) and has no license declaration.

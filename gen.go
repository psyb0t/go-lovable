// Package lovable is a Go client for the Lovable API,
// generated with oapi-codegen from the OpenAPI document Lovable publishes.
package lovable

//go:generate bash scripts/fetch-spec.sh
//go:generate bash scripts/fix-nullable-types.sh
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.fixed.yaml

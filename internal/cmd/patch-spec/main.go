// Command patch-spec converts Lovable's OpenAPI 3.1 schemas for oapi-codegen.
package main

import (
	"os"
	"strings"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"gopkg.in/yaml.v3"
)

const (
	sourcePath         = "api/openapi.yaml"
	outputPath         = "api/openapi.fixed.yaml"
	outputMode         = 0o600
	typeKeyword        = "type"
	nullableKeyword    = "nullable"
	enumKeyword        = "enum"
	exampleKeyword     = "example"
	examplesKeyword    = "examples"
	descriptionKeyword = "description"
	webhooksKeyword    = "webhooks"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return ctxerrors.Wrap(err, "read source spec")
	}

	var spec map[string]any

	if err := yaml.Unmarshal(data, &spec); err != nil {
		return ctxerrors.Wrap(err, "decode source spec")
	}

	if err := convert(spec); err != nil {
		return ctxerrors.Wrap(err, "convert OpenAPI schemas")
	}

	data, err = yaml.Marshal(spec)
	if err != nil {
		return ctxerrors.Wrap(err, "encode patched spec")
	}

	if err := os.WriteFile(outputPath, data, outputMode); err != nil {
		return ctxerrors.Wrap(err, "write patched spec")
	}

	return nil
}

func convert(spec map[string]any) error {
	if err := patch(spec); err != nil {
		return err
	}

	// oapi-codegen emits webhook request body aliases without the schemas
	// they point to, and the client never calls webhooks.
	delete(spec, webhooksKeyword)

	spec["openapi"] = "3.0.3"

	return nil
}

func patch(value any) error {
	switch node := value.(type) {
	case map[string]any:
		if err := patchSchema(node); err != nil {
			return ctxerrors.Wrap(err, "convert schema keywords")
		}

		for key, child := range node {
			// Payloads can contain arbitrary property names.
			switch key {
			case exampleKeyword, examplesKeyword, "default", enumKeyword:
				continue
			}

			if err := patch(child); err != nil {
				return ctxerrors.Wrapf(err, "convert key %s", key)
			}
		}
	case []any:
		for _, child := range node {
			if err := patch(child); err != nil {
				return ctxerrors.Wrap(err, "convert array item")
			}
		}
	}

	return nil
}

func patchSchema(node map[string]any) error {
	if types, ok := node[typeKeyword].([]any); ok {
		if err := patchType(node, types); err != nil {
			return ctxerrors.Wrap(err, "convert nullable type list")
		}
	}

	patchEnum(node)
	patchExamples(node)

	// Keep generated documentation consistent with repository terminology.
	if text, ok := node[descriptionKeyword].(string); ok {
		node[descriptionKeyword] = strings.ReplaceAll(
			text, "folder", "directory",
		)
	}

	return nil
}

func patchEnum(node map[string]any) {
	if node[nullableKeyword] != true {
		return
	}

	values, ok := node[enumKeyword].([]any)
	if !ok {
		return
	}

	filtered := make([]any, 0, len(values))

	for _, value := range values {
		if value == nil {
			continue
		}

		filtered = append(filtered, value)
	}

	node[enumKeyword] = filtered
}

func patchExamples(node map[string]any) {
	examples, ok := node[examplesKeyword].([]any)
	if !ok {
		return
	}

	if len(examples) > 0 {
		if _, exists := node[exampleKeyword]; !exists {
			node[exampleKeyword] = examples[0]
		}
	}

	delete(node, examplesKeyword)
}

func patchType(node map[string]any, types []any) error {
	const nullableTypeCount = 2

	if len(types) != nullableTypeCount {
		return ctxerrors.Wrap(commerr.ErrValidationFailed,
			"expected one concrete type and null")
	}

	var concrete string

	hasNull := false

	for _, value := range types {
		name, ok := value.(string)
		if !ok {
			return ctxerrors.Wrap(commerr.ErrValidationFailed,
				"type list contains a non-string")
		}

		if name == "null" {
			hasNull = true

			continue
		}

		concrete = name
	}

	if !hasNull || concrete == "" {
		return ctxerrors.Wrap(commerr.ErrValidationFailed,
			"unsupported type union")
	}

	node[typeKeyword] = concrete
	node[nullableKeyword] = true

	return nil
}

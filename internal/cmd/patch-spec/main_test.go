package main

import (
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPatch(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name: "nullable enum",
			input: `type: [string, 'null']
enum: [public, null]
`,
			want: `type: string
nullable: true
enum: [public]
`,
		},
		{
			name:  "reversed nullable number",
			input: "type: ['null', number]\nformat: double\n",
			want:  "type: number\nnullable: true\nformat: double\n",
		},
		{
			name: "example payload preserved",
			input: `type: object
examples:
  - type: [string, 'null']
    enum: [null]
`,
			want: `type: object
example:
  type: [string, 'null']
  enum: [null]
`,
		},
		{
			name: "parameter examples preserved",
			input: `examples:
  named:
    value: {type: [string, 'null']}
`,
			want: `examples:
  named:
    value: {type: [string, 'null']}
`,
		},
		{
			name:  "documentation terminology",
			input: "description: folder grants\n",
			want:  "description: directory grants\n",
		},
		{
			name:    "multiple concrete types",
			input:   "type: [string, number]\n",
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "unquoted YAML null type",
			input:   "type: [string, null]\n",
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "unknown union shape",
			input:   "type: [string, number, 'null']\n",
			wantErr: commerr.ErrValidationFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var input map[string]any

			require.NoError(t, yaml.Unmarshal([]byte(tc.input), &input))
			err := patch(input)
			require.ErrorIs(t, err, tc.wantErr)

			if tc.wantErr != nil {
				return
			}

			var want map[string]any

			require.NoError(t, yaml.Unmarshal([]byte(tc.want), &want))
			assert.Equal(t, want, input)
		})
	}
}

func TestConvertDropsWebhooks(t *testing.T) {
	t.Parallel()

	spec := map[string]any{
		"openapi":  "3.1.0",
		"paths":    map[string]any{},
		"webhooks": map[string]any{"project.created.v1": map[string]any{}},
	}

	require.NoError(t, convert(spec))
	assert.Equal(t, map[string]any{
		"openapi": "3.0.3",
		"paths":   map[string]any{},
	}, spec)
}

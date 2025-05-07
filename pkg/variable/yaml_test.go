package variable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestProcessNode(t *testing.T) {
	testcases := []struct {
		name     string
		yamlStr  string
		expected map[string]any
	}{
		{
			name: "scalar value of map",
			yamlStr: `
name: alice
age: 30
`,
			expected: map[string]any{
				"name": "alice",
				"age":  int64(30),
			},
		},
		{
			name: "map value of map",
			yamlStr: `
user:
  name: alice
  age: 30
`,
			expected: map[string]any{
				"user": map[string]any{
					"name": "alice",
					"age":  int64(30),
				},
			},
		},
		{
			name: "scalar value of sequence",
			yamlStr: `
user:
  - alice
  - carol
`,
			expected: map[string]any{
				"user": []any{"alice", "carol"},
			},
		},
		{
			name: "map value of sequence",
			yamlStr: `
user:
  - name: carol
`,
			expected: map[string]any{
				"user": []any{map[string]any{"name": "carol"}},
			},
		},
		{
			name: "sequence of sequences",
			yamlStr: `
matrix:
  - [1, 2, 3]
  - [4, 5, 6]
`,
			expected: map[string]any{
				"matrix": []any{
					[]any{int64(1), int64(2), int64(3)},
					[]any{int64(4), int64(5), int64(6)},
				},
			},
		},
		{
			name: "deeply nested map and sequence",
			yamlStr: `
app:
  name: myapp
  env:
    - dev
    - staging
    - prod
  config:
    ports:
      - 80
      - 443
`,
			expected: map[string]any{
				"app": map[string]any{
					"name": "myapp",
					"env":  []any{"dev", "staging", "prod"},
					"config": map[string]any{
						"ports": []any{int64(80), int64(443)},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tc.yamlStr), &node)
			if err != nil {
				t.Fatalf("failed to unmarshal YAML: %v", err)
			}

			if len(node.Content) == 0 {
				t.Fatalf("empty YAML content")
			}

			ctx := make(map[string]any)
			if err = processNode(ctx, node.Content[0]); err != nil {
				t.Fatalf("processNode failed: %v", err)
			}

			assert.Equal(t, tc.expected, ctx)
		})
	}
}

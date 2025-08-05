package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalWhen(t *testing.T) {
	testcases := []struct {
		name    string
		content string
		except  []string
	}{
		{
			name: "test single string",
			content: `
when: .a | eq "b"`,
			except: []string{
				".a | eq \"b\"",
			},
		},
		{
			name: "test multi string",
			content: `
when: 
  - .a | eq "b"
  - .b | ne "c"`,
			except: []string{
				".a | eq \"b\"",
				".b | ne \"c\"",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var when Conditional
			err := yaml.Unmarshal([]byte(tc.content), &when)
			assert.NoError(t, err)
			assert.Equal(t, tc.except, when.When.Data)
		})
	}
}

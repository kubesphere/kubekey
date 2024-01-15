/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tmpl

import (
	"testing"

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		ctx    pongo2.Context
		except string
	}{
		{
			name:  "default",
			input: "{{ os.release.Name | default_if_none:false }}",
			ctx: map[string]any{
				"os": map[string]any{
					"release": map[string]any{
						"ID": "a",
					},
				},
			},
			except: "False",
		},
		{
			name:  "default_if_none",
			input: "{{ os.release.Name | default_if_none:'b' }}",
			ctx: map[string]any{
				"os": map[string]any{
					"release": map[string]any{
						"ID": "a",
					},
				},
			},
			except: "b",
		},
		{
			name:  "defined",
			input: "{{ test | defined }}",
			ctx: map[string]any{
				"test": "aaa",
			},
			except: "True",
		},
		{
			name:  "version_greater",
			input: "{{ test | version:'>=,v1.19.0'  }}",
			ctx: map[string]any{
				"test": "v1.23.10",
			},
			except: "True",
		},
		{
			name:  "divisibleby",
			input: "{{ not test['a'] | length | divisibleby:2 }}",
			ctx: map[string]any{
				"test": map[string]any{
					"a": "1",
				},
			},
			except: "True",
		},
		{
			name:  "power",
			input: "{{ (test | integer) >= (2 | pow: test2 | integer  ) }}",
			ctx: map[string]any{
				"test":  "12",
				"test2": "3s",
			},
			except: "True",
		},
		{
			name:  "split",
			input: "{{ kernelVersion | split:'-' | first }}",
			ctx: map[string]any{
				"kernelVersion": "5.15.0-89-generic",
			},
			except: "5.15.0",
		},
		{
			name:  "match",
			input: "{{ test | match:regex }}",
			ctx: map[string]any{
				"test":  "abc",
				"regex": "[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\\\\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$",
			},
			except: "True",
		},
	}

	for _, tc := range testcases {
		t.Run("filter: "+tc.name, func(t *testing.T) {
			tql, err := pongo2.FromString(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			result, err := tql.Execute(tc.ctx)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.except, result)
		})
	}
}

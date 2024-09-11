/*
Copyright 2024 The KubeSphere Authors.

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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSetValue(t *testing.T) {
	testcases := []struct {
		name   string
		key    string
		val    any
		except Config
	}{
		{
			name:   "one level",
			key:    "a",
			val:    2,
			except: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":2}`)}},
		},
		{
			name:   "two level repeat key",
			key:    "a.b",
			val:    2,
			except: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":{"b":2}}`)}},
		},
		{
			name:   "two level no-repeat key",
			key:    "b.c",
			val:    2,
			except: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":1,"b":{"c":2}}`)}},
		},
	}

	for _, tc := range testcases {
		in := Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":1}`)}}
		t.Run(tc.name, func(t *testing.T) {
			err := in.SetValue(tc.key, tc.val)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.except, in)
		})
	}
}

func TestGetValue(t *testing.T) {
	testcases := []struct {
		name   string
		key    string
		config Config
		except any
	}{
		{
			name:   "all value",
			key:    "",
			config: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":1}`)}},
			except: map[string]any{
				"a": int64(1),
			},
		},
		{
			name:   "none value",
			key:    "b",
			config: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":1}`)}},
			except: nil,
		},
		{
			name:   "none multi value",
			key:    "b.c",
			config: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":1}`)}},
			except: nil,
		},
		{
			name:   "find one value",
			key:    "a",
			config: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":1}`)}},
			except: int64(1),
		},
		{
			name:   "find mulit value",
			key:    "a.b",
			config: Config{Spec: runtime.RawExtension{Raw: []byte(`{"a":{"b":1}}`)}},
			except: int64(1),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			value, _ := tc.config.GetValue(tc.key)
			assert.Equal(t, tc.except, value)
		})
	}
}

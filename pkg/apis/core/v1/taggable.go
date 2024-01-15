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

package v1

import "k8s.io/utils/strings/slices"

type Taggable struct {
	Tags []string `yaml:"tags,omitempty"`
}

// IsEnabled check if the block should be executed
func (t Taggable) IsEnabled(onlyTags []string, skipTags []string) bool {
	shouldRun := true

	if len(onlyTags) > 0 {
		if slices.Contains(t.Tags, "always") {
			shouldRun = true
		} else if slices.Contains(onlyTags, "all") && !slices.Contains(t.Tags, "never") {
			shouldRun = true
		} else if slices.Contains(onlyTags, "tagged") && len(onlyTags) > 0 && !slices.Contains(t.Tags, "never") {
			shouldRun = true
		} else if !isdisjoint(onlyTags, t.Tags) {
			shouldRun = true
		} else {
			shouldRun = false
		}
	}

	if shouldRun && len(skipTags) > 0 {
		if slices.Contains(skipTags, "all") {
			if !slices.Contains(t.Tags, "always") || !slices.Contains(skipTags, "always") {
				shouldRun = false
			}
		} else if !isdisjoint(skipTags, t.Tags) {
			shouldRun = false
		} else if slices.Contains(skipTags, "tagged") && len(skipTags) > 0 {
			shouldRun = false
		}
	}

	return shouldRun
}

// isdisjoint returns true if a and b have no elements in common.
func isdisjoint(a, b []string) bool {
	for _, s := range a {
		if slices.Contains(b, s) {
			return false
		}
	}
	return true
}

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

import "slices"

// the special tags
const (
	// AlwaysTag it always run
	AlwaysTag = "always"
	// NeverTag it never run
	NeverTag = "never"
	// AllTag represent all tags
	AllTag = "all"
	// TaggedTag represent which has tags
	TaggedTag = "tagged"
)

// Taggable if it should executor
type Taggable struct {
	Tags []string `yaml:"tags,omitempty"`
}

// IsEnabled check if the block should be executed
func (t Taggable) IsEnabled(onlyTags []string, skipTags []string) bool {
	shouldRun := true

	if len(onlyTags) > 0 {
		switch {
		case slices.Contains(t.Tags, AlwaysTag):
			shouldRun = true
		case slices.Contains(onlyTags, AllTag) && !slices.Contains(t.Tags, NeverTag):
			shouldRun = true
		case slices.Contains(onlyTags, TaggedTag) && !slices.Contains(t.Tags, NeverTag):
			shouldRun = true
		case !isdisjoint(onlyTags, t.Tags):
			shouldRun = true
		default:
			shouldRun = false
		}
	}

	if shouldRun && len(skipTags) > 0 {
		switch {
		case slices.Contains(skipTags, AllTag) &&
			(!slices.Contains(t.Tags, AlwaysTag) || !slices.Contains(skipTags, AlwaysTag)):
			shouldRun = false
		case !isdisjoint(skipTags, t.Tags):
			shouldRun = false
		case slices.Contains(skipTags, TaggedTag) && len(skipTags) > 0:
			shouldRun = false
		}
	}

	return shouldRun
}

// JoinTag the child block should inherit tag for parent block
func JoinTag(child, parent Taggable) Taggable {
	for _, tag := range parent.Tags {
		if !slices.Contains(child.Tags, tag) {
			child.Tags = append(child.Tags, tag)
		}
	}

	return child
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

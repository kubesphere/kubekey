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

package multipath_conf

import (
	"strings"
	"testing"
)

func TestUpdateMultipathConfigCreatesBlacklistBlock(t *testing.T) {
	t.Parallel()

	got, changed, err := updateMultipathConfig("", []string{`^sd[a-z]`})
	if err != nil {
		t.Fatalf("updateMultipathConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("expected changed = true")
	}
	if !strings.Contains(got, "blacklist {") {
		t.Fatalf("missing blacklist block: %q", got)
	}
	if !strings.Contains(got, `devnode "^sd[a-z]"`) {
		t.Fatalf("missing devnode rule: %q", got)
	}
}

func TestUpdateMultipathConfigAppendsMissingRulesOnly(t *testing.T) {
	t.Parallel()

	input := "defaults {\n    user_friendly_names yes\n}\n\nblacklist {\n    devnode \"^sd[a-z]\"\n}\n"
	got, changed, err := updateMultipathConfig(input, []string{`^sd[a-z]`, `^nvme[0-9]n[0-9]`})
	if err != nil {
		t.Fatalf("updateMultipathConfig() error = %v", err)
	}
	if !changed {
		t.Fatal("expected changed = true")
	}
	if strings.Count(got, `devnode "^sd[a-z]"`) != 1 {
		t.Fatalf("duplicate sd rule: %q", got)
	}
	if !strings.Contains(got, `devnode "^nvme[0-9]n[0-9]"`) {
		t.Fatalf("missing nvme rule: %q", got)
	}
}

func TestUpdateMultipathConfigNoChangeWhenRulesExist(t *testing.T) {
	t.Parallel()

	input := "blacklist {\n    devnode \"^sd[a-z]\"\n}\n"
	got, changed, err := updateMultipathConfig(input, []string{`^sd[a-z]`})
	if err != nil {
		t.Fatalf("updateMultipathConfig() error = %v", err)
	}
	if changed {
		t.Fatalf("expected changed = false, got %q", got)
	}
	if got != input {
		t.Fatalf("content changed unexpectedly:\n%s", got)
	}
}

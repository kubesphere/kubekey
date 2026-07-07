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

import "testing"

func TestParseItemDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := parseItem(map[string]any{})
	if err != nil {
		t.Fatalf("parseItem() error = %v", err)
	}
	if cfg.Path != defaultPath {
		t.Fatalf("Path = %q, want %q", cfg.Path, defaultPath)
	}
	if !cfg.Backup || !cfg.Reload {
		t.Fatalf("Backup = %t Reload = %t, want both true", cfg.Backup, cfg.Reload)
	}
	if len(cfg.DevNodes) != 4 {
		t.Fatalf("DevNodes = %#v, want 4 defaults", cfg.DevNodes)
	}
}

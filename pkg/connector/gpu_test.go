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

package connector

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestGPUVendorConfig(t *testing.T) *GPUVendorConfig {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	src := filepath.Join(repoRoot, "docs", "en", "reference", "gpu_vendors.yaml")

	tmp := t.TempDir()
	dstDir := filepath.Join(tmp, "gather_facts")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("create temp gather_facts dir: %v", err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read sample gpu vendor config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "gpu_vendors.yaml"), data, 0644); err != nil {
		t.Fatalf("write temp gpu vendor config: %v", err)
	}

	cfg, err := LoadGPUVendorConfig(tmp)
	if err != nil {
		t.Fatalf("load gpu vendor config: %v", err)
	}
	return cfg
}

func TestGPUVendorConfigRestrictsAIComputeClasses(t *testing.T) {
	cfg := loadTestGPUVendorConfig(t)

	assert.True(t, cfg.IsGPUClass("0300", "VGA compatible controller"))
	assert.False(t, cfg.IsGPUClass("0380", "Display controller"))
	assert.False(t, cfg.IsGPUClass("", "Display adapter"))
	assert.False(t, cfg.IsGPUClass("", "Graphics card"))
	assert.False(t, cfg.IsGPUClass("", "Co-processor"))
	assert.True(t, cfg.IsGPUClass("0302", "3D controller"))
	assert.True(t, cfg.IsGPUClass("1200", "Processing accelerators"))
	assert.True(t, cfg.IsGPUClass("", "Compute accelerator"))
	assert.True(t, cfg.IsGPUClass("", "Neural Processing Unit"))
}

func TestParseLspciMMRestrictsAIComputeDevices(t *testing.T) {
	cfg := loadTestGPUVendorConfig(t)

	tests := []struct {
		name            string
		line            string
		wantDetected    bool
		wantDriverClass string
	}{
		{
			name:         "huawei ibmc vga is not an AI device",
			line:         `06:00.0 "VGA compatible controller [0300]" "Huawei Technologies Co., Ltd. [19e5]" "Hi171x Series [iBMC Intelligent Management system chip w/VGA support] [1711]" -r01 "Huawei Technologies Co., Ltd. [19e5]" "Hi171x Series [iBMC Intelligent Management system chip w/VGA support] [1711]"`,
			wantDetected: false,
		},
		{
			name:         "display controller class is ignored even for a known vendor",
			line:         `09:00.0 "Display controller [0380]" "NVIDIA Corporation [10de]" "Display Device [1eb8]" -ra1 "" ""`,
			wantDetected: false,
		},
		{
			name:            "nvidia 3d controller remains detected",
			line:            `01:00.0 "3D controller [0302]" "NVIDIA Corporation [10de]" "GA100 [A100 PCIe 40GB] [20f1]" -ra1 "" ""`,
			wantDetected:    true,
			wantDriverClass: "nvidia",
		},
		{
			name:            "hygon processing accelerator remains detected",
			line:            `03:00.0 "Processing accelerators [1200]" "Hygon [1d94]" "DCU Device [1000]" -r00 "" ""`,
			wantDetected:    true,
			wantDriverClass: "hygon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLspciMM(tt.line, cfg)
			if !tt.wantDetected {
				assert.Nil(t, got)
				return
			}
			if assert.NotNil(t, got) {
				assert.Equal(t, tt.wantDriverClass, got.DriverClass)
			}
		})
	}
}

func TestSplitLastBracketedID(t *testing.T) {
	name, id := splitLastBracketedID(`GA100 [A100 PCIe 40GB] [20f1]`)
	assert.Equal(t, "GA100 [A100 PCIe 40GB]", name)
	assert.Equal(t, "20f1", id)
}

func TestExtractQuotedFields(t *testing.T) {
	fields := extractQuotedFields(`01:00.0 "3D controller [0302]" "NVIDIA Corporation [10de]" "GA100 [A100 PCIe 40GB] [20f1]"`)
	assert.Len(t, fields, 3)
	assert.Equal(t, "3D controller [0302]", fields[0])
}

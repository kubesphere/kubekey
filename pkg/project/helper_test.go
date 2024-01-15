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

package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPlaybookBaseFromAbsPlaybook(t *testing.T) {
	testcases := []struct {
		name         string
		basePlaybook string
		playbook     string
		except       string
	}{
		{
			name:         "find from project/playbooks/playbook",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			playbook:     "playbook2.yaml",
			except:       filepath.Join("playbooks", "playbook2.yaml"),
		},
		{
			name:         "find from current_playbook/playbooks/playbook",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			playbook:     "playbook3.yaml",
			except:       filepath.Join("playbooks", "playbooks", "playbook3.yaml"),
		},
		{
			name:         "cannot find",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			playbook:     "playbook4.yaml",
			except:       "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.except, GetPlaybookBaseFromPlaybook(os.DirFS("testdata"), tc.basePlaybook, tc.playbook))
		})
	}
}

func TestGetRoleBaseFromAbsPlaybook(t *testing.T) {
	testcases := []struct {
		name         string
		basePlaybook string
		roleName     string
		except       string
	}{
		{
			name:         "find from project/roles/roleName",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			roleName:     "role1",
			except:       filepath.Join("roles", "role1"),
		},
		{
			name:         "find from current_playbook/roles/roleName",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			roleName:     "role2",
			except:       filepath.Join("playbooks", "roles", "role2"),
		},
		{
			name:         "cannot find",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			roleName:     "role3",
			except:       "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.except, GetRoleBaseFromPlaybook(os.DirFS("testdata"), tc.basePlaybook, tc.roleName))
		})
	}
}

func TestGetFilesFromPlayBook(t *testing.T) {
	testcases := []struct {
		name     string
		pbPath   string
		role     string
		filePath string
		excepted string
	}{
		{
			name:     "absolute filePath",
			filePath: "/tmp",
			excepted: "/tmp",
		},
		{
			name:     "empty role",
			pbPath:   "playbooks/test.yaml",
			filePath: "tmp",
			excepted: "playbooks/files/tmp",
		},
		{
			name:     "not empty role",
			pbPath:   "playbooks/test.yaml",
			role:     "role1",
			filePath: "tmp",
			excepted: "roles/role1/files/tmp",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.excepted, GetFilesFromPlayBook(os.DirFS("testdata"), tc.pbPath, tc.role, tc.filePath))
		})
	}
}

func TestGetTemplatesFromPlayBook(t *testing.T) {
	testcases := []struct {
		name     string
		pbPath   string
		role     string
		filePath string
		excepted string
	}{
		{
			name:     "absolute filePath",
			filePath: "/tmp",
			excepted: "/tmp",
		},
		{
			name:     "empty role",
			pbPath:   "playbooks/test.yaml",
			filePath: "tmp",
			excepted: "playbooks/templates/tmp",
		},
		{
			name:     "not empty role",
			pbPath:   "playbooks/test.yaml",
			role:     "role1",
			filePath: "tmp",
			excepted: "roles/role1/templates/tmp",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.excepted, GetTemplatesFromPlayBook(os.DirFS("testdata"), tc.pbPath, tc.role, tc.filePath))
		})
	}

}

func TestGetYamlFile(t *testing.T) {
	testcases := []struct {
		name   string
		base   string
		except string
	}{
		{
			name:   "get yaml",
			base:   filepath.Join("playbooks", "playbook2"),
			except: filepath.Join("playbooks", "playbook2.yaml"),
		},
		{
			name:   "get yml",
			base:   filepath.Join("playbooks", "playbook3"),
			except: filepath.Join("playbooks", "playbook3.yml"),
		},
		{
			name:   "cannot find",
			base:   filepath.Join("playbooks", "playbook4"),
			except: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.except, GetYamlFile(os.DirFS("testdata"), tc.base))
		})
	}
}

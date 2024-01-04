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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// GetPlaybookBaseFromPlaybook
// find from project/playbooks/playbook if exists.
// find from current_playbook/playbooks/playbook if exists.
// find current_playbook/playbook
func GetPlaybookBaseFromPlaybook(baseFS fs.FS, pbPath string, playbook string) string {
	var find []string
	// find from project/playbooks/playbook
	find = append(find, filepath.Join(filepath.Dir(filepath.Dir(pbPath)), _const.ProjectPlaybooksDir, playbook))
	// find from pbPath dir like: current_playbook/playbooks/playbook
	find = append(find, filepath.Join(filepath.Dir(pbPath), _const.ProjectPlaybooksDir, playbook))
	// find from pbPath dir like: current_playbook/playbook
	find = append(find, filepath.Join(filepath.Dir(pbPath), playbook))
	for _, s := range find {
		if baseFS != nil {
			if _, err := fs.Stat(baseFS, s); err == nil {
				return s
			}
		} else {
			if baseFS != nil {
				if _, err := fs.Stat(baseFS, s); err == nil {
					return s
				}
			} else {
				if _, err := os.Stat(s); err == nil {
					return s
				}
			}
		}
	}

	return ""
}

// GetRoleBaseFromPlaybook
// find from project/roles/roleName if exists.
// find from current_playbook/roles/roleName if exists.
// find current_playbook/playbook
func GetRoleBaseFromPlaybook(baseFS fs.FS, pbPath string, roleName string) string {
	var find []string
	// find from project/roles/roleName
	find = append(find, filepath.Join(filepath.Dir(filepath.Dir(pbPath)), _const.ProjectRolesDir, roleName))
	// find from pbPath dir like: current_playbook/roles/roleName
	find = append(find, filepath.Join(filepath.Dir(pbPath), _const.ProjectRolesDir, roleName))

	for _, s := range find {
		if baseFS != nil {
			if _, err := fs.Stat(baseFS, s); err == nil {
				return s
			}
		} else {
			if _, err := os.Stat(s); err == nil {
				return s
			}
		}
	}

	return ""
}

// GetFilesFromPlayBook
func GetFilesFromPlayBook(baseFS fs.FS, pbPath string, roleName string, filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}

	if roleName != "" {
		return filepath.Join(GetRoleBaseFromPlaybook(baseFS, pbPath, roleName), _const.ProjectRolesFilesDir, filePath)
	} else {
		// find from pbPath dir like: project/playbooks/templates/tmplPath
		return filepath.Join(filepath.Dir(pbPath), _const.ProjectRolesFilesDir, filePath)
	}
}

// GetTemplatesFromPlayBook
func GetTemplatesFromPlayBook(baseFS fs.FS, pbPath string, roleName string, tmplPath string) string {
	if filepath.IsAbs(tmplPath) {
		return tmplPath
	}

	if roleName != "" {
		return filepath.Join(GetRoleBaseFromPlaybook(baseFS, pbPath, roleName), _const.ProjectRolesTemplateDir, tmplPath)
	} else {
		// find from pbPath dir like: project/playbooks/templates/tmplPath
		return filepath.Join(filepath.Dir(pbPath), _const.ProjectRolesTemplateDir, tmplPath)
	}
}

// GetYamlFile
// return *.yaml if exists
// return  *.yml if exists.
func GetYamlFile(baseFS fs.FS, base string) string {
	var find []string
	find = append(find,
		fmt.Sprintf("%s.yaml", base),
		fmt.Sprintf("%s.yml", base))

	for _, s := range find {
		if baseFS != nil {
			if _, err := fs.Stat(baseFS, s); err == nil {
				return s
			}
		} else {
			if _, err := os.Stat(s); err == nil {
				return s
			}
		}
	}

	return ""
}

package project

import (
	"path/filepath"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

const (
	// PathFormatImportPlaybook defines the directory structure for importing playbooks
	// The structure supports three formats:
	// 1. Import playbook in same directory as base playbook
	// 2. Import playbook from playbooks/ subdirectory
	// 3. Import playbook from project root playbooks/ directory
	PathFormatImportPlaybook = `
|-- base-playbook.yaml
|-- [import_playbook]

|-- playbook.yaml
|-- playbooks/
|   |-- [import_playbook]

|-- [projectDir]
|-- playbooks/
|   |-- [import_playbook]
`

	// PathFormatVarsFile defines the directory structure for vars files
	// Vars files are expected to be in the same directory as the base playbook
	PathFormatVarsFile = `
|-- base-playbook.yaml
|-- [VarsFile]	
`
	// PathFormatRole defines the directory structure for roles
	// Roles can be found in three locations:
	// 1. roles/ directory at project root
	// 2. roles/ directory next to playbook
	// 3. roles/ directory at project root
	PathFormatRole = `
|-- playbooks/
|   |-- playbook.yaml
|-- roles/
|   |-- [role]/

|-- playbook.yaml
|-- roles/
|   |-- [role]/	

|-- [projectDir]
|-- roles/
|   |-- [role]/
`

	// PathFormatRoleMeta defines the directory structure for role meta information.
	// The meta/main.yaml (or main.yml) file can contain dependencies on other roles via the "dependencies" key.
	PathFormatRoleMeta = `
|-- baseRole/
|   |-- meta/
|   |   |-- main.yaml

|-- baseRole/
|   |-- meta/
|   |   |-- main.yml
`

	// PathFormatRoleTask defines the directory structure for role tasks
	// Role tasks are expected to be in tasks/main.yaml or tasks/main.yml under the role directory
	PathFormatRoleTask = `
|-- baseRole/
|   |-- tasks/
|   |   |-- main.yaml

|-- baseRole/
|   |-- tasks/
|   |   |-- main.yml
`

	// PathFormatIncludeTask defines the directory structure for included tasks
	// Tasks can be included from:
	// 1. Direct path under source role
	// 2. tasks/ directory under source role
	// 3. Direct path under top level role
	// 4. tasks/ directory under top level role
	PathFormatIncludeTask = `
|-- [source_task]
|   |-- tasks/
|   |   |-- [include_tasks]

|-- [source_task]
|   |-- [include_tasks]

|-- [top_source_task]
|   |-- tasks/
|   |   |-- [include_tasks]

|-- [top_source_task]
|   |-- [include_tasks]
`
)

// GetProjectPath from basePlaybook. the Project directory structure can be either:
/*
|-- projectFS/
|   |-- basePlaybook.yaml

|-- projectFS/
|   |-- playbooks/
|   |   |-- basePlaybook.yaml
*/
func GetProjectPath(basePlaybook string) string {
	if filepath.Base(filepath.Dir(basePlaybook)) == _const.ProjectPlaybooksDir {
		return filepath.Dir(filepath.Dir(basePlaybook))
	}

	return filepath.Dir(basePlaybook)
}

// GetImportPlaybookRelPath returns possible relative paths for an imported playbook based on the base playbook location
// The format follows PathFormatImportPlaybook structure
func GetImportPlaybookRelPath(basePlaybook string, includePlaybook string) []string {
	return []string{
		filepath.Join(filepath.Dir(basePlaybook), includePlaybook),
		filepath.Join(filepath.Dir(basePlaybook), _const.ProjectPlaybooksDir, includePlaybook),
		// should support find playbooks from projectDir
		filepath.Join(_const.ProjectPlaybooksDir, includePlaybook),
	}
}

// GetVarsFilesRelPath returns possible relative paths for vars files based on the base playbook location
// The format follows PathFormatVarsFile structure
func GetVarsFilesRelPath(basePlaybook string, varsFile string) []string {
	return []string{
		filepath.Join(filepath.Dir(basePlaybook), varsFile),
	}
}

// GetRoleRelPath returns possible relative paths for a role based on the base playbook location
// The format follows PathFormatRole structure
func GetRoleRelPath(basePlaybook string, role string) []string {
	return []string{
		filepath.Join(filepath.Dir(filepath.Dir(basePlaybook)), _const.ProjectRolesDir, role),
		filepath.Join(filepath.Dir(basePlaybook), _const.ProjectRolesDir, role),
		filepath.Join(_const.ProjectRolesDir, role),
	}
}

// GetRoleMetaRelPath returns possible relative paths for a role's meta file (main.yaml or main.yml) within the meta directory.
// The format follows the standard Ansible role meta directory structure.
func GetRoleMetaRelPath(baseRole string) []string {
	return []string{
		filepath.Join(baseRole, _const.ProjectRolesMetaDir, "main.yaml"),
		filepath.Join(baseRole, _const.ProjectRolesMetaDir, "main.yml"),
	}
}

// GetRoleTaskRelPath returns possible relative paths for a role's main task file
// The format follows PathFormatRoleTask structure
func GetRoleTaskRelPath(baseRole string) []string {
	return []string{
		filepath.Join(baseRole, _const.ProjectRolesTasksDir, "main.yaml"),
		filepath.Join(baseRole, _const.ProjectRolesTasksDir, "main.yml"),
	}
}

// GetRoleDefaultsRelPath returns possible relative paths for a role's defaults file
// The format follows similar structure to role tasks
func GetRoleDefaultsRelPath(baseRole string) []string {
	return []string{
		filepath.Join(baseRole, _const.ProjectRolesDefaultsDir, "main.yaml"),
		filepath.Join(baseRole, _const.ProjectRolesDefaultsDir, "main.yml"),
	}
}

// GetIncludeTaskRelPath returns possible relative paths for included task files
// The format follows PathFormatIncludeTask structure
func GetIncludeTaskRelPath(top string, source string, includeTask string) []string {
	return []string{
		filepath.Join(source, includeTask),
		filepath.Join(source, _const.ProjectRolesTasksDir, includeTask),
		filepath.Join(top, includeTask),
		filepath.Join(top, _const.ProjectRolesTasksDir, includeTask),
	}
}

// Standard project directory structure when playbook does not define a role and directly defines tasks:
/*
|-- projects_dir/
|   |-- project1/
|   |   |-- playbook.yaml
|   |   |-- files/
|   |   |-- templates/

|-- projects_dir/
|   |-- project1/
|   |   |-- playbooks/
|   |   |   |-- playbook.yaml
|   |   |   |-- files/
|   |   |   |-- templates/
*/

// Standard roles directory structure:
/*
|-- roles/
|   |-- defaults/
|   |-- files/
|   |-- tasks/
|   |-- templates/
*/

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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"gopkg.in/yaml.v3"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// marshalPlaybook kkprojectv1.Playbook from a playbook file
func marshalPlaybook(baseFS fs.FS, pbPath string) (*kkprojectv1.Playbook, error) {
	// convert playbook to kkprojectv1.Playbook
	pb := &kkprojectv1.Playbook{}
	if err := loadPlaybook(baseFS, pbPath, pb); err != nil {
		return nil, err
	}
	// convertRoles.
	if err := convertRoles(baseFS, pbPath, pb); err != nil {
		return nil, err
	}
	// convertIncludeTasks
	if err := convertIncludeTasks(baseFS, pbPath, pb); err != nil {
		return nil, err
	}
	// validate playbook
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return pb, nil
}

// loadPlaybook with include_playbook. Join all playbooks into one playbook
func loadPlaybook(baseFS fs.FS, pbPath string, pb *kkprojectv1.Playbook) error {
	// baseDir is the local ansible project dir which playbook belong to
	pbData, err := fs.ReadFile(baseFS, pbPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read playbook %q", pbPath)
	}
	var plays []kkprojectv1.Play
	if err := yaml.Unmarshal(pbData, &plays); err != nil {
		return errors.Wrapf(err, "failed to unmarshal playbook %q", pbPath)
	}

	for _, p := range plays {
		if err := dealImportPlaybook(p, baseFS, pbPath, pb); err != nil {
			return err
		}

		if err := dealVarsFiles(&p, baseFS, pbPath); err != nil {
			return err
		}
		// fill block in roles
		if err := dealRoles(p, baseFS, pbPath); err != nil {
			return err
		}

		pb.Play = append(pb.Play, p)
	}

	return nil
}

// dealImportPlaybook "import_playbook" argument in play
func dealImportPlaybook(p kkprojectv1.Play, baseFS fs.FS, pbPath string, pb *kkprojectv1.Playbook) error {
	if p.ImportPlaybook != "" {
		importPlaybook := getPlaybookBaseFromPlaybook(baseFS, pbPath, p.ImportPlaybook)
		if importPlaybook == "" {
			return errors.Errorf("import_playbook %q path is empty, it's maybe [project-dir/playbooks/import_playbook_file, playbook-dir/playbooks/import_playbook-file, playbook-dir/import_playbook-file]", p.ImportPlaybook)
		}
		if err := loadPlaybook(baseFS, importPlaybook, pb); err != nil {
			return err
		}
	}

	return nil
}

// dealVarsFiles "var_files" argument in play
func dealVarsFiles(p *kkprojectv1.Play, baseFS fs.FS, pbPath string) error {
	for _, file := range p.VarsFiles {
		// load vars from vars_files
		if _, err := fs.Stat(baseFS, filepath.Join(filepath.Dir(pbPath), file)); err != nil {
			return errors.Wrapf(err, "failed to stat file %q", file)
		}
		data, err := fs.ReadFile(baseFS, filepath.Join(filepath.Dir(pbPath), file))
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", filepath.Join(filepath.Dir(pbPath), file))
		}
		var newVars map[string]any
		// Unmarshal the YAML document into a root node.
		if err := yaml.Unmarshal(data, &newVars); err != nil {
			return errors.Wrap(err, "failed to failed to unmarshal YAML")
		}
		// store vars in play. the vars defined in file should not be repeated.
		p.Vars = variable.CombineVariables(newVars, p.Vars)
	}

	return nil
}

// dealRoles "roles" argument in play
func dealRoles(p kkprojectv1.Play, baseFS fs.FS, pbPath string) error {
	for i, r := range p.Roles {
		roleBase := getRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
		if roleBase == "" {
			return errors.Errorf("cannot found Role %q", r.Role)
		}

		mainTask := getYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir, _const.ProjectRolesTasksMainFile))
		if mainTask == "" {
			return errors.Errorf("cannot found main task for Role %q", r.Role)
		}

		rdata, err := fs.ReadFile(baseFS, mainTask)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", mainTask)
		}
		var blocks []kkprojectv1.Block
		if err := yaml.Unmarshal(rdata, &blocks); err != nil {
			return errors.Wrapf(err, "failed to unmarshal yaml file %q", filepath.Join(filepath.Dir(pbPath), mainTask))
		}
		p.Roles[i].Block = blocks
	}

	return nil
}

// convertRoles convert roleName to block
func convertRoles(baseFS fs.FS, pbPath string, pb *kkprojectv1.Playbook) error {
	for i, p := range pb.Play {
		for i, r := range p.Roles {
			roleBase := getRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if roleBase == "" {
				return errors.Errorf("cannot found Role %q in playbook %q", r.Role, pbPath)
			}

			var err error
			if p.Roles[i].Block, err = convertRoleBlocks(baseFS, pbPath, roleBase); err != nil {
				return err
			}

			if err = convertRoleVars(baseFS, roleBase, &p.Roles[i]); err != nil {
				return err
			}
		}
		pb.Play[i] = p
	}

	return nil
}

func convertRoleVars(baseFS fs.FS, roleBase string, role *kkprojectv1.Role) error {
	// load defaults (optional)
	defaultsFile := getYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesDefaultsDir, _const.ProjectRolesDefaultsMainFile))
	if defaultsFile != "" {
		data, err := fs.ReadFile(baseFS, defaultsFile)
		if err != nil {
			return errors.Wrapf(err, "failed to read defaults variable file %q", defaultsFile)
		}

		var newVars map[string]any
		// Unmarshal the YAML document into a root node.
		if err := yaml.Unmarshal(data, &newVars); err != nil {
			return errors.Wrap(err, "failed to unmarshal YAML")
		}
		// store vars in play. the vars defined in file should not be repeated.
		role.Vars = variable.CombineVariables(newVars, role.Vars)
	}

	return nil
}

// convertRoleBlocks roles/task/main.yaml to []kkprojectv1.Block
func convertRoleBlocks(baseFS fs.FS, pbPath string, roleBase string) ([]kkprojectv1.Block, error) {
	mainTask := getYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir, _const.ProjectRolesTasksMainFile))
	if mainTask == "" {
		return nil, errors.Errorf("cannot found main task for Role %q", roleBase)
	}

	rdata, err := fs.ReadFile(baseFS, mainTask)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %q", mainTask)
	}
	var blocks []kkprojectv1.Block
	if err := yaml.Unmarshal(rdata, &blocks); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal yaml file %q", filepath.Join(filepath.Dir(pbPath), mainTask))
	}

	return blocks, nil
}

// convertIncludeTasks from file to blocks
func convertIncludeTasks(baseFS fs.FS, pbPath string, pb *kkprojectv1.Playbook) error {
	var pbBase = filepath.Dir(filepath.Dir(pbPath))
	for _, play := range pb.Play {
		if err := fileToBlock(baseFS, pbBase, play.PreTasks); err != nil {
			return err
		}

		if err := fileToBlock(baseFS, pbBase, play.Tasks); err != nil {
			return err
		}

		if err := fileToBlock(baseFS, pbBase, play.PostTasks); err != nil {
			return err
		}

		for _, r := range play.Roles {
			roleBase := getRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if err := fileToBlock(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir), r.Block); err != nil {
				return err
			}
		}
	}

	return nil
}

func fileToBlock(baseFS fs.FS, baseDir string, blocks []kkprojectv1.Block) error {
	for i, b := range blocks {
		if b.IncludeTasks != "" {
			data, err := fs.ReadFile(baseFS, filepath.Join(baseDir, b.IncludeTasks))
			if err != nil {
				return errors.Wrapf(err, "failed to read includeTask file %q", filepath.Join(baseDir, b.IncludeTasks))
			}
			var bs []kkprojectv1.Block
			if err := yaml.Unmarshal(data, &bs); err != nil {
				return errors.Wrapf(err, "failed to unmarshal includeTask file %q", filepath.Join(baseDir, b.IncludeTasks))
			}

			b.Block = bs
			blocks[i] = b
		}

		if err := fileToBlock(baseFS, baseDir, b.Block); err != nil {
			return err
		}

		if err := fileToBlock(baseFS, baseDir, b.Rescue); err != nil {
			return err
		}

		if err := fileToBlock(baseFS, baseDir, b.Always); err != nil {
			return err
		}
	}

	return nil
}

// getPlaybookBaseFromPlaybook find import_playbook path base on the current_playbook
// find from project/playbooks/playbook if exists.
// find from current_playbook/playbooks/playbook if exists.
// find current_playbook/playbook
func getPlaybookBaseFromPlaybook(baseFS fs.FS, pbPath string, playbook string) string {
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
			if _, err := os.Stat(s); err == nil {
				return s
			}
		}
	}

	return ""
}

// getRoleBaseFromPlaybook
// find from project/roles/roleName if exists.
// find from current_playbook/roles/roleName if exists.
// find current_playbook/playbook
func getRoleBaseFromPlaybook(baseFS fs.FS, pbPath string, roleName string) string {
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

// getYamlFile
// return *.yaml if exists
// return  *.yml if exists.
func getYamlFile(baseFS fs.FS, base string) string {
	var find []string
	find = append(find, base+".yaml", base+".yml")

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

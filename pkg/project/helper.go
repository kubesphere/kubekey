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

	"gopkg.in/yaml.v3"

	projectv1 "github.com/kubesphere/kubekey/v4/pkg/apis/project/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// marshalPlaybook projectv1.Playbook from a playbook file
func marshalPlaybook(baseFS fs.FS, pbPath string) (*projectv1.Playbook, error) {
	// convert playbook to projectv1.Playbook
	pb := &projectv1.Playbook{}
	if err := loadPlaybook(baseFS, pbPath, pb); err != nil {
		return nil, fmt.Errorf("load playbook failed: %w", err)
	}

	// convertRoles
	if err := convertRoles(baseFS, pbPath, pb); err != nil {
		return nil, fmt.Errorf("convert roles failed: %w", err)
	}

	if err := convertIncludeTasks(baseFS, pbPath, pb); err != nil {
		return nil, fmt.Errorf("convert include tasks failed: %w", err)
	}

	if err := pb.Validate(); err != nil {
		return nil, fmt.Errorf("validate playbook failed: %w", err)
	}
	return pb, nil
}

// loadPlaybook with include_playbook. Join all playbooks into one playbook
func loadPlaybook(baseFS fs.FS, pbPath string, pb *projectv1.Playbook) error {
	// baseDir is the local ansible project dir which playbook belong to
	pbData, err := fs.ReadFile(baseFS, pbPath)
	if err != nil {
		return fmt.Errorf("read playbook failed: %w", err)
	}
	var plays []projectv1.Play
	if err := yaml.Unmarshal(pbData, &plays); err != nil {
		return fmt.Errorf("unmarshal playbook failed: %w", err)
	}

	for _, p := range plays {
		if p.ImportPlaybook != "" {
			importPlaybook := getPlaybookBaseFromPlaybook(baseFS, pbPath, p.ImportPlaybook)
			if importPlaybook == "" {
				return fmt.Errorf("import playbook %s failed", p.ImportPlaybook)
			}
			if err := loadPlaybook(baseFS, importPlaybook, pb); err != nil {
				return fmt.Errorf("load playbook failed: %w", err)
			}
		}

		// load var_files (optional)
		for _, file := range p.VarsFiles {
			if _, err := fs.Stat(baseFS, filepath.Join(filepath.Dir(pbPath), file)); err != nil {
				return fmt.Errorf("file %s not exists", file)
			}
			mainData, err := fs.ReadFile(baseFS, filepath.Join(filepath.Dir(pbPath), file))
			if err != nil {
				return fmt.Errorf("read file %s failed: %w", filepath.Join(filepath.Dir(pbPath), file), err)
			}

			var vars map[string]any
			var node yaml.Node // marshal file on defined order
			if err := yaml.Unmarshal(mainData, &node); err != nil {
				return fmt.Errorf("unmarshal yaml file: %s failed: %w", filepath.Join(filepath.Dir(pbPath), file), err)
			}
			if err := node.Decode(&vars); err != nil {
				return fmt.Errorf("unmarshal yaml file: %s failed: %w", filepath.Join(filepath.Dir(pbPath), file), err)
			}

			p.Vars, err = combineMaps(p.Vars, vars)
			if err != nil {
				return fmt.Errorf("combine maps file:%s failed: %w", filepath.Join(filepath.Dir(pbPath), file), err)
			}
		}

		// fill block in roles
		for i, r := range p.Roles {
			roleBase := getRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if roleBase == "" {
				return fmt.Errorf("cannot found Role %s", r.Role)
			}
			mainTask := getYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir, _const.ProjectRolesTasksMainFile))
			if mainTask == "" {
				return fmt.Errorf("cannot found main task for Role %s", r.Role)
			}

			rdata, err := fs.ReadFile(baseFS, mainTask)
			if err != nil {
				return fmt.Errorf("read file %s failed: %w", mainTask, err)
			}
			var blocks []projectv1.Block
			if err := yaml.Unmarshal(rdata, &blocks); err != nil {
				return fmt.Errorf("unmarshal yaml file: %s failed: %w", filepath.Join(filepath.Dir(pbPath), mainTask), err)
			}
			p.Roles[i].Block = blocks
		}
		pb.Play = append(pb.Play, p)
	}

	return nil
}

// convertRoles convert roleName to block
func convertRoles(baseFS fs.FS, pbPath string, pb *projectv1.Playbook) error {
	for i, p := range pb.Play {
		for i, r := range p.Roles {
			roleBase := getRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if roleBase == "" {
				return fmt.Errorf("cannot found Role %s", r.Role)
			}

			// load block
			mainTask := getYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir, _const.ProjectRolesTasksMainFile))
			if mainTask == "" {
				return fmt.Errorf("cannot found main task for Role %s", r.Role)
			}

			rdata, err := fs.ReadFile(baseFS, mainTask)
			if err != nil {
				return fmt.Errorf("read file %s failed: %w", mainTask, err)
			}
			var blocks []projectv1.Block
			if err := yaml.Unmarshal(rdata, &blocks); err != nil {
				return fmt.Errorf("unmarshal yaml file: %s failed: %w", filepath.Join(filepath.Dir(pbPath), mainTask), err)
			}
			p.Roles[i].Block = blocks

			// load defaults (optional)
			mainDefault := getYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesDefaultsDir, _const.ProjectRolesDefaultsMainFile))
			if mainDefault != "" {
				mainData, err := fs.ReadFile(baseFS, mainDefault)
				if err != nil {
					return fmt.Errorf("read defaults variable file %s failed: %w", mainDefault, err)
				}

				var vars map[string]any
				var node yaml.Node // marshal file on defined order
				if err := yaml.Unmarshal(mainData, &node); err != nil {
					return fmt.Errorf("unmarshal defaults variable yaml file: %s failed: %w", mainDefault, err)
				}
				if err := node.Decode(&vars); err != nil {
					return fmt.Errorf("decode defaults variable yaml file: %s failed: %w", mainDefault, err)
				}

				p.Roles[i].Vars, err = combineMaps(p.Roles[i].Vars, vars)
				if err != nil {
					return fmt.Errorf("combine defaults variable failed: %w", err)
				}
			}
		}
		pb.Play[i] = p
	}
	return nil
}

// convertIncludeTasks from file to blocks
func convertIncludeTasks(baseFS fs.FS, pbPath string, pb *projectv1.Playbook) error {
	var pbBase = filepath.Dir(filepath.Dir(pbPath))
	for _, play := range pb.Play {
		if err := fileToBlock(baseFS, pbBase, play.PreTasks); err != nil {
			return fmt.Errorf("convert pre_tasks file %s failed: %w", pbPath, err)
		}
		if err := fileToBlock(baseFS, pbBase, play.Tasks); err != nil {
			return fmt.Errorf("convert tasks file %s failed: %w", pbPath, err)
		}
		if err := fileToBlock(baseFS, pbBase, play.PostTasks); err != nil {
			return fmt.Errorf("convert post_tasks file %s failed: %w", pbPath, err)
		}

		for _, r := range play.Roles {
			roleBase := getRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if err := fileToBlock(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir), r.Block); err != nil {
				return fmt.Errorf("convert role %s failed: %w", filepath.Join(pbPath, r.Role), err)
			}
		}
	}
	return nil
}

func fileToBlock(baseFS fs.FS, baseDir string, blocks []projectv1.Block) error {
	for i, b := range blocks {
		if b.IncludeTasks != "" {
			data, err := fs.ReadFile(baseFS, filepath.Join(baseDir, b.IncludeTasks))
			if err != nil {
				return fmt.Errorf("read includeTask file %s failed: %w", filepath.Join(baseDir, b.IncludeTasks), err)
			}
			var bs []projectv1.Block
			if err := yaml.Unmarshal(data, &bs); err != nil {
				return fmt.Errorf("unmarshal includeTask file %s failed: %w", filepath.Join(baseDir, b.IncludeTasks), err)
			}
			b.Block = bs
			blocks[i] = b
		}
		if err := fileToBlock(baseFS, baseDir, b.Block); err != nil {
			return fmt.Errorf("convert block file %s failed: %w", filepath.Join(baseDir, b.IncludeTasks), err)
		}
		if err := fileToBlock(baseFS, baseDir, b.Rescue); err != nil {
			return fmt.Errorf("convert rescue file %s failed: %w", filepath.Join(baseDir, b.IncludeTasks), err)
		}
		if err := fileToBlock(baseFS, baseDir, b.Always); err != nil {
			return fmt.Errorf("convert always file %s failed: %w", filepath.Join(baseDir, b.IncludeTasks), err)
		}
	}
	return nil
}

// getPlaybookBaseFromPlaybook
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

// combine v2 map to v1 if not repeat.
func combineMaps(v1, v2 map[string]any) (map[string]any, error) {
	if len(v1) == 0 {
		return v2, nil
	}

	mv := make(map[string]any)
	for k, v := range v1 {
		mv[k] = v
	}
	for k, v := range v2 {
		if _, ok := mv[k]; ok {
			return nil, fmt.Errorf("duplicate key: %s", k)
		}
		mv[k] = v
	}
	return mv, nil
}

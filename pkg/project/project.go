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

// Package project provides functionality for managing Ansible-like projects in KubeKey.
// It handles project file operations, playbook parsing, and task execution.
package project

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"gopkg.in/yaml.v3"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// builtinProjectFunc is a function that creates a Project from a built-in playbook
var builtinProjectFunc func(kkcorev1.Playbook) (Project, error)

// Project represent location of actual project.
// get project file should base on it
type Project interface {
	// MarshalPlaybook project file to playbook.
	MarshalPlaybook() (*kkprojectv1.Playbook, error)
	// Stat file or dir in project
	Stat(path string) (os.FileInfo, error)
	// WalkDir dir in project
	WalkDir(path string, f fs.WalkDirFunc) error
	// ReadFile file or dir in project
	ReadFile(path string) ([]byte, error)
	// Rel path file or dir in project
	Rel(root string, path string) (string, error)
}

// New creates a new Project instance based on the provided playbook.
// If project address is git format, it creates a git project.
// If playbook has BuiltinsProjectAnnotation, it uses builtinProjectFunc.
// Otherwise, it creates a local project.
func New(ctx context.Context, playbook kkcorev1.Playbook, update bool) (Project, error) {
	if strings.HasPrefix(playbook.Spec.Project.Addr, "https://") ||
		strings.HasPrefix(playbook.Spec.Project.Addr, "http://") ||
		strings.HasPrefix(playbook.Spec.Project.Addr, "git@") {
		return newGitProject(ctx, playbook, update)
	}

	if _, ok := playbook.Annotations[kkcorev1.BuiltinsProjectAnnotation]; ok {
		return builtinProjectFunc(playbook)
	}

	return newLocalProject(playbook)
}

// project implements the Project interface using an fs.FS
type project struct {
	fs.FS

	basePlaybook string

	*kkprojectv1.Playbook
}

// ReadFile reads and returns the contents of the file at the given path
func (f *project) ReadFile(path string) ([]byte, error) {
	return fs.ReadFile(f.FS, path)
}

// Rel returns a relative path that is lexically equivalent to targpath when joined to basepath
func (f *project) Rel(root string, path string) (string, error) {
	return filepath.Rel(root, path)
}

// Stat returns the FileInfo for the file at the given path
func (f *project) Stat(path string) (os.FileInfo, error) {
	return fs.Stat(f.FS, path)
}

// WalkDir walks the file tree rooted at path, calling fn for each file or directory
func (f *project) WalkDir(path string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(f.FS, path, fn)
}

// MarshalPlaybook converts a playbook file into a kkprojectv1.Playbook
func (f *project) MarshalPlaybook() (*kkprojectv1.Playbook, error) {
	f.Playbook = &kkprojectv1.Playbook{}
	// convert playbook to kkprojectv1.Playbook
	if err := f.loadPlaybook(f.basePlaybook); err != nil {
		return nil, err
	}
	// validate playbook
	if err := f.Playbook.Validate(); err != nil {
		return nil, err
	}

	return f.Playbook, nil
}

// loadPlaybook loads a playbook and all its included playbooks into a single playbook
func (f *project) loadPlaybook(basePlaybook string) error {
	// baseDir is the local ansible project dir which playbook belong to
	pbData, err := fs.ReadFile(f.FS, basePlaybook)
	if err != nil {
		return errors.Wrapf(err, "failed to read playbook %q", basePlaybook)
	}
	var plays []kkprojectv1.Play
	if err := yaml.Unmarshal(pbData, &plays); err != nil {
		return errors.Wrapf(err, "failed to unmarshal playbook %q", basePlaybook)
	}

	for _, p := range plays {
		if err := f.dealImportPlaybook(p, basePlaybook); err != nil {
			return err
		}

		if err := f.dealVarsFiles(&p, basePlaybook); err != nil {
			return err
		}
		// deal "pre_tasks"
		if err := f.dealBlock(filepath.Dir(basePlaybook), filepath.Dir(basePlaybook), p.PreTasks); err != nil {
			return err
		}
		// deal "tasks"
		if err := f.dealBlock(filepath.Dir(basePlaybook), filepath.Dir(basePlaybook), p.Tasks); err != nil {
			return err
		}
		// deal "post_tasks"
		if err := f.dealBlock(filepath.Dir(basePlaybook), filepath.Dir(basePlaybook), p.PostTasks); err != nil {
			return err
		}

		//deal "roles"
		for i := range p.Roles {
			if err := f.dealRole(&p.Roles[i], basePlaybook); err != nil {
				return err
			}
			// deal tasks
			if err := f.dealRoleTask(&p.Roles[i], basePlaybook); err != nil {
				return err
			}
		}

		f.Playbook.Play = append(f.Playbook.Play, p)
	}

	return nil
}

// dealImportPlaybook handles the "import_playbook" argument in a play
func (f *project) dealImportPlaybook(p kkprojectv1.Play, basePlaybook string) error {
	if p.ImportPlaybook != "" {
		importPlaybook := f.getPath(GetImportPlaybookRelPath(basePlaybook, p.ImportPlaybook))
		if importPlaybook == "" {
			return errors.Errorf("failed to find import_playbook %q base on %q. it's should be:\n %s", p.ImportPlaybook, basePlaybook, PathFormatImportPlaybook)
		}
		if err := f.loadPlaybook(importPlaybook); err != nil {
			return err
		}
	}

	return nil
}

// dealVarsFiles handles the "var_files" argument in a play
func (f *project) dealVarsFiles(p *kkprojectv1.Play, basePlaybook string) error {
	for _, varsFile := range p.VarsFiles {
		// load vars from vars_files
		file := f.getPath(GetVarsFilesRelPath(basePlaybook, varsFile))
		if file == "" {
			return errors.Errorf("failed to find vars_files %q base on %q. it's should be:\n %s", varsFile, basePlaybook, PathFormatVarsFile)
		}
		data, err := fs.ReadFile(f.FS, file)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", file)
		}
		var node yaml.Node
		// Unmarshal the YAML document into a root node.
		if err := yaml.Unmarshal(data, &node); err != nil {
			return errors.Wrap(err, "failed to failed to unmarshal YAML")
		}
		if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
			return errors.Errorf("unsupport vars_files format. it should be single map file")
		}
		// combine map node
		if node.Content[0].Kind == yaml.MappingNode {
			// skip empty file
			p.Vars = *variable.CombineMappingNode(&p.Vars, node.Content[0])
		}
	}

	return nil
}

func (f *project) dealRole(role *kkprojectv1.Role, basePlaybook string) error {
	baseRole := f.getPath(GetRoleRelPath(basePlaybook, role.Role))
	if baseRole == "" {
		return errors.Errorf("failed to find role %q base on %q. it's should be:\n %s", role.Role, basePlaybook, PathFormatRole)
	}
	// deal dependency
	if meta := f.getPath(GetRoleMetaRelPath(baseRole)); meta != "" {
		mdata, err := fs.ReadFile(f.FS, meta)
		if err != nil {
			return errors.Wrapf(err, "failed to read role meta file %q", meta)
		}
		roleMeta := &kkprojectv1.Role{}
		if err := yaml.Unmarshal(mdata, roleMeta); err != nil {
			return errors.Wrapf(err, "failed to unmarshal role meta file %q", meta)
		}
		for _, dep := range roleMeta.RoleDependency {
			if err := f.dealRole(&dep, basePlaybook); err != nil {
				return errors.Wrapf(err, "failed to deal dependency role base %q", role.Role)
			}
			role.RoleDependency = append(role.RoleDependency, dep)
		}
	}
	// deal tasks
	if task := f.getPath(GetRoleTaskRelPath(baseRole)); task != "" {
		rdata, err := fs.ReadFile(f.FS, task)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", task)
		}
		var blocks []kkprojectv1.Block
		if err := yaml.Unmarshal(rdata, &blocks); err != nil {
			return errors.Wrapf(err, "failed to unmarshal yaml file %q", task)
		}
		role.Block = blocks
	}
	// deal defaults (optional)
	if defaults := f.getPath(GetRoleDefaultsRelPath(baseRole)); defaults != "" {
		data, err := fs.ReadFile(f.FS, defaults)
		if err != nil {
			return errors.Wrapf(err, "failed to read defaults variable file %q", defaults)
		}

		var node yaml.Node
		// Unmarshal the YAML document into a root node.
		if err := yaml.Unmarshal(data, &node); err != nil {
			return errors.Wrap(err, "failed to unmarshal YAML")
		}
		if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
			return errors.Errorf("unsupport vars_files format. it should be single map file")
		}
		// combine map node
		if node.Content[0].Kind == yaml.MappingNode {
			// skip empty file
			role.Vars = *variable.CombineMappingNode(&role.Vars, node.Content[0])
		}
	}
	return nil
}

// dealRoleTask recursively processes the tasks for a given role and its dependencies.
// It ensures that all dependent roles are processed before handling the current role's tasks.
func (f *project) dealRoleTask(role *kkprojectv1.Role, basePlaybook string) error {
	for i := range role.RoleDependency {
		if err := f.dealRoleTask(&role.RoleDependency[i], basePlaybook); err != nil {
			return err
		}
	}
	// Get the base path for the current role
	baseRole := f.getPath(GetRoleRelPath(basePlaybook, role.Role))
	// Process the tasks for the current role
	return f.dealBlock(baseRole, filepath.Join(baseRole, _const.ProjectRolesTasksDir), role.Block)
}

// dealBlock recursively processes blocks, handling nested blocks, include_tasks, and annotating tasks with their relative path.
func (f *project) dealBlock(top string, source string, blocks []kkprojectv1.Block) error {
	for i, block := range blocks {
		switch {
		case len(block.Block) != 0: // it's a block with nested blocks (block, rescue, always)
			// Recursively process nested blocks
			if err := f.dealBlock(top, source, block.Block); err != nil {
				return err
			}
			if err := f.dealBlock(top, source, block.Rescue); err != nil {
				return err
			}
			if err := f.dealBlock(top, source, block.Always); err != nil {
				return err
			}
		case block.IncludeTasks != "": // it's an include_tasks directive
			// Resolve the path to the include_tasks file
			includeTask := f.getPath(GetIncludeTaskRelPath(top, source, block.IncludeTasks))
			if includeTask == "" {
				return errors.Errorf("failed to find include_task %q base on %q. it's should be:\n %s", block.IncludeTasks, source, PathFormatIncludeTask)
			}
			// Read the include_tasks file
			data, err := fs.ReadFile(f.FS, includeTask)
			if err != nil {
				return errors.Wrapf(err, "failed to read includeTask file %q", includeTask)
			}
			// Unmarshal the file into blocks
			var includeBlocks []kkprojectv1.Block
			if err := yaml.Unmarshal(data, &includeBlocks); err != nil {
				return errors.Wrapf(err, "failed to unmarshal includeTask file %q", includeTask)
			}
			// Recursively process the included blocks
			if err := f.dealBlock(top, filepath.Dir(includeTask), includeBlocks); err != nil {
				return err
			}
			// Assign the included blocks to the current block
			blocks[i].Block = includeBlocks
		default: // it's a regular task
			// Annotate the task with its relative path
			blocks[i].UnknownField["annotations"] = map[string]string{
				kkcorev1alpha1.TaskAnnotationRelativePath: top,
			}
		}
	}

	return nil
}

// getPath returns the first valid path from a list of possible paths
func (f *project) getPath(paths []string) string {
	for _, path := range paths {
		if _, err := fs.Stat(f.FS, path); err == nil {
			return path
		}
	}

	return ""
}

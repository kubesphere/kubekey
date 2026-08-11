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
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/utils"
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

	config        map[string]any
	playbookGraph *utils.KahnGraph
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
	if err := f.loadPlaybook("", f.basePlaybook); err != nil {
		return nil, err
	}
	// validate playbook
	if err := f.Validate(); err != nil {
		return nil, err
	}

	return f.Playbook, nil
}

// loadPlaybook loads a playbook and all its included playbooks into a single playbook
func (f *project) loadPlaybook(fromPlayBook, basePlaybook string) error {
	// baseDir is the local ansible project dir which playbook belong to
	if f.playbookGraph.AddEdgeAndCheckCycle(fromPlayBook, basePlaybook) {
		// play book cycle imported
		return errors.Errorf("failed to import %s because it cause a cycle import", basePlaybook)
	}

	pbData, err := fs.ReadFile(f.FS, basePlaybook)
	if err != nil {
		return errors.Wrapf(err, "failed to read playbook %q", basePlaybook)
	}
	var plays []kkprojectv1.Play
	if err := yaml.Unmarshal(pbData, &plays); err != nil {
		return errors.Wrapf(err, "failed to unmarshal playbook %q", basePlaybook)
	}

	// Inject configured playbooks at their orders for the top-level playbook
	// only. Imported playbooks keep their own content.
	if fromPlayBook == "" {
		plays, err = f.injectPlaybooks(plays)
		if err != nil {
			return err
		}
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

		// A play that is purely an import_playbook directive carries no
		// standalone content; skip appending it so it does not show up as an
		// empty play in the final playbook. Its imported plays are already
		// loaded by dealImportPlaybook above.
		if p.ImportPlaybook != "" {
			continue
		}
		f.Play = append(f.Play, p)
	}

	return nil
}

// dealImportPlaybook handles the "import_playbook" argument in a play
func (f *project) dealImportPlaybook(p kkprojectv1.Play, basePlaybook string) error {
	if p.ImportPlaybook != "" {
		// Render the import path with template syntax so that it can be
		// customized via variables and the `playbooks` order injection.
		importPlaybookPath, err := tmpl.ParseFunc(f.config, p.ImportPlaybook, tmpl.StringFunc)
		if err != nil {
			return errors.Wrapf(err, "failed to parse import_playbook %q", p.ImportPlaybook)
		}
		// An import_playbook that renders to an empty string is treated as a
		// no-op and skipped silently. This makes variable-driven imports
		// opt-in: reference them as `{{ .var | default "" }}`, leave the
		// variable unset (or set it to "") to disable the import, and point it
		// at a path to enable it.
		if strings.TrimSpace(importPlaybookPath) == "" {
			klog.V(5).InfoS("skip empty import_playbook", "import_playbook", p.ImportPlaybook, "base", basePlaybook)
			return nil
		}
		importPlaybook, _ := f.getPath(GetImportPlaybookRelPath(basePlaybook, importPlaybookPath))
		if importPlaybook == "" {
			return errors.Errorf("failed to find import_playbook %q base on %q. it's should be:\n %s", p.ImportPlaybook, basePlaybook, PathFormatImportPlaybook)
		}
		if basePlaybook == importPlaybook {
			// play book import self
			return errors.Errorf("failed to import %s because it is already imported", p.ImportPlaybook)
		}
		if err := f.loadPlaybook(basePlaybook, importPlaybook); err != nil {
			return err
		}
	}

	return nil
}

// playbookInjection defines a playbook to inject into the top-level playbook at
// a specific position. Order is a sort weight: original (file) plays get
// 0,1,2,... by their document order, while injected (config) plays keep their
// explicit order.
type playbookInjection struct {
	Order float64 `yaml:"order" json:"order"`
	Path  string  `yaml:"path" json:"path"`
}

// injectPlaybooks merges configured playbook injections into the play list and
// returns the combined list sorted by order. The sort contract is:
//   - original (file) plays get orders 0,1,2,... by document order;
//   - injected (config) plays keep their explicit order (float, may be negative
//     or fractional);
//   - on a tie, config plays win over file plays, then definition order
//     (config list order / file document order).
//
// The injected paths are rendered as templates; an empty rendered path is
// skipped. Only the top-level playbook is injected (see loadPlaybook).
func (f *project) injectPlaybooks(plays []kkprojectv1.Play) ([]kkprojectv1.Play, error) {
	raw, ok := f.config["playbooks"]
	if !ok || raw == nil {
		return plays, nil
	}
	// Round-trip through YAML so we don't depend on the exact map types
	// produced by the config JSON decoder.
	buf, err := yaml.Marshal(raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal playbooks injection config")
	}
	var injections []playbookInjection
	if err := yaml.Unmarshal(buf, &injections); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal playbooks injection config")
	}

	type entry struct {
		order   float64
		fromCfg bool
		seq     int
		play    kkprojectv1.Play
	}
	entries := make([]entry, 0, len(plays)+len(injections))
	for i, p := range plays {
		entries = append(entries, entry{order: float64(i), fromCfg: false, seq: i, play: p})
	}
	for i, inj := range injections {
		path, err := tmpl.ParseFunc(f.config, inj.Path, tmpl.StringFunc)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse injected playbook path %q", inj.Path)
		}
		if strings.TrimSpace(path) == "" {
			klog.V(5).InfoS("skip empty injected playbook", "order", inj.Order, "path", inj.Path)
			continue
		}
		entries = append(entries, entry{
			order:   inj.Order,
			fromCfg: true,
			seq:     i,
			play:    kkprojectv1.Play{ImportPlaybook: path},
		})
	}

	sort.SliceStable(entries, func(a, b int) bool {
		if entries[a].order != entries[b].order {
			return entries[a].order < entries[b].order
		}
		// Config injections win over file plays on a tie.
		if entries[a].fromCfg != entries[b].fromCfg {
			return entries[a].fromCfg
		}
		return entries[a].seq < entries[b].seq
	})

	out := make([]kkprojectv1.Play, len(entries))
	for i, e := range entries {
		out[i] = e.play
	}
	return out, nil
}

// dealVarsFiles handles the "vars_files" argument in a play
func (f *project) dealVarsFiles(p *kkprojectv1.Play, basePlaybook string) error {
	for _, varsFileStr := range p.VarsFiles {
		// load vars from vars_files
		varsFile, err := tmpl.ParseFunc(f.config, varsFileStr, tmpl.StringFunc)
		if err != nil {
			return errors.Errorf("failed to parse varFile %q", varsFileStr)
		}
		file, _ := f.getPath(GetVarsFilesRelPath(basePlaybook, varsFile))
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
			p.Vars.Nodes = append(p.Vars.Nodes, *node.Content[0])
		}
	}

	return nil
}

func (f *project) dealRole(role *kkprojectv1.Role, basePlaybook string) error {
	baseRole, _ := f.getPath(GetRoleRelPath(basePlaybook, role.Role))
	if baseRole == "" {
		return errors.Errorf("failed to find role %q base on %q. it's should be:\n %s", role.Role, basePlaybook, PathFormatRole)
	}
	// deal dependency
	if meta, _ := f.getPath(GetRoleMetaRelPath(baseRole)); meta != "" {
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
	if task, _ := f.getPath(GetRoleTaskRelPath(baseRole)); task != "" {
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
	if defaults, _ := f.getPath(GetRoleDefaultsRelPath(baseRole)); defaults != "" {
		data, err := fs.ReadFile(f.FS, defaults)
		if err != nil {
			return errors.Wrapf(err, "failed to read defaults variable file %q", defaults)
		}

		err = f.combineRoleVars(role, data)
		if err != nil {
			return err
		}
	}
	if dirDefaults, info := f.getPath(GetRoleDefaultsRelDirPath(baseRole)); dirDefaults != "" {
		// only handle [roles]/defaults/main directory,if file found but not a directory,skip file
		if info.IsDir() {
			err := utils.ReadDirFiles(f.FS, dirDefaults, func(data []byte) error {
				return f.combineRoleVars(role, data)
			})
			if err != nil {
				return errors.Wrapf(err, "failed to read defaults variable file %q", dirDefaults)
			}
		}
	}
	return nil
}

func (f *project) combineRoleVars(role *kkprojectv1.Role, content []byte) error {
	var node yaml.Node
	// Unmarshal the YAML document into a root node.
	if err := yaml.Unmarshal(content, &node); err != nil {
		return errors.Wrap(err, "failed to unmarshal YAML")
	}
	if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
		return errors.Errorf("unsupport vars_files format. it should be single map file")
	}
	// combine map node
	if node.Content[0].Kind == yaml.MappingNode {
		// skip empty file
		role.Vars.Nodes = append(role.Vars.Nodes, *node.Content[0])
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
	baseRole, _ := f.getPath(GetRoleRelPath(basePlaybook, role.Role))
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
			includeTask, _ := f.getPath(GetIncludeTaskRelPath(top, source, block.IncludeTasks))
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
func (f *project) getPath(paths []string) (string, fs.FileInfo) {
	for _, path := range paths {
		if info, err := fs.Stat(f.FS, path); err == nil {
			return path, info
		}
	}

	return "", nil
}

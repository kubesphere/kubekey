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

package converter

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// MarshalPlaybook kkcorev1.Playbook from a playbook file
func MarshalPlaybook(baseFS fs.FS, pbPath string) (*kkcorev1.Playbook, error) {
	// convert playbook to kkcorev1.Playbook
	pb := &kkcorev1.Playbook{}
	if err := loadPlaybook(baseFS, pbPath, pb); err != nil {
		klog.V(4).ErrorS(err, "Load playbook failed", "playbook", pbPath)
		return nil, err
	}

	// convertRoles
	if err := convertRoles(baseFS, pbPath, pb); err != nil {
		klog.V(4).ErrorS(err, "ConvertRoles error", "playbook", pbPath)
		return nil, err
	}

	if err := convertIncludeTasks(baseFS, pbPath, pb); err != nil {
		klog.V(4).ErrorS(err, "ConvertIncludeTasks error", "playbook", pbPath)
		return nil, err
	}

	if err := pb.Validate(); err != nil {
		klog.V(4).ErrorS(err, "Validate playbook failed", "playbook", pbPath)
		return nil, err
	}
	return pb, nil
}

// loadPlaybook with include_playbook. Join all playbooks into one playbook
func loadPlaybook(baseFS fs.FS, pbPath string, pb *kkcorev1.Playbook) error {
	// baseDir is the local ansible project dir which playbook belong to
	pbData, err := fs.ReadFile(baseFS, pbPath)
	if err != nil {
		klog.V(4).ErrorS(err, "Read playbook failed", "playbook", pbPath)
		return err
	}
	var plays []kkcorev1.Play
	if err := yaml.Unmarshal(pbData, &plays); err != nil {
		klog.V(4).ErrorS(err, "Unmarshal playbook failed", "playbook", pbPath)
		return err
	}

	for _, p := range plays {
		if p.ImportPlaybook != "" {
			importPlaybook := project.GetPlaybookBaseFromPlaybook(baseFS, pbPath, p.ImportPlaybook)
			if importPlaybook == "" {
				return fmt.Errorf("cannot found import playbook %s", importPlaybook)
			}
			if err := loadPlaybook(baseFS, importPlaybook, pb); err != nil {
				return err
			}
		}

		// fill block in roles
		for i, r := range p.Roles {
			roleBase := project.GetRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if roleBase == "" {
				return fmt.Errorf("cannot found role %s", r.Role)
			}
			mainTask := project.GetYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir, _const.ProjectRolesTasksMainFile))
			if mainTask == "" {
				return fmt.Errorf("cannot found main task for role %s", r.Role)
			}

			rdata, err := fs.ReadFile(baseFS, mainTask)
			if err != nil {
				klog.V(4).ErrorS(err, "Read role failed", "playbook", pbPath, "role", r.Role)
				return err
			}
			var blocks []kkcorev1.Block
			if err := yaml.Unmarshal(rdata, &blocks); err != nil {
				klog.V(4).ErrorS(err, "Unmarshal role failed", "playbook", pbPath, "role", r.Role)
				return err
			}
			p.Roles[i].Block = blocks
		}
		pb.Play = append(pb.Play, p)
	}

	return nil
}

// convertRoles convert roleName to block
func convertRoles(baseFS fs.FS, pbPath string, pb *kkcorev1.Playbook) error {
	for i, p := range pb.Play {
		for i, r := range p.Roles {
			roleBase := project.GetRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if roleBase == "" {
				return fmt.Errorf("cannot found role %s", r.Role)
			}

			// load block
			mainTask := project.GetYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir, _const.ProjectRolesTasksMainFile))
			if mainTask == "" {
				return fmt.Errorf("cannot found main task for role %s", r.Role)
			}

			rdata, err := fs.ReadFile(baseFS, mainTask)
			if err != nil {
				klog.V(4).ErrorS(err, "Read role failed", "playbook", pbPath, "role", r.Role)
				return err
			}
			var blocks []kkcorev1.Block
			if err := yaml.Unmarshal(rdata, &blocks); err != nil {
				klog.V(4).ErrorS(err, "Unmarshal role failed", "playbook", pbPath, "role", r.Role)
				return err
			}
			p.Roles[i].Block = blocks

			// load defaults (optional)
			mainDefault := project.GetYamlFile(baseFS, filepath.Join(roleBase, _const.ProjectRolesDefaultsDir, _const.ProjectRolesDefaultsMainFile))
			if mainDefault != "" {
				mainData, err := fs.ReadFile(baseFS, mainDefault)
				if err != nil {
					klog.V(4).ErrorS(err, "Read defaults variable for role error", "playbook", pbPath, "role", r.Role)
					return err
				}
				var vars variable.VariableData
				if err := yaml.Unmarshal(mainData, &vars); err != nil {
					klog.V(4).ErrorS(err, "Unmarshal defaults variable for role error", "playbook", pbPath, "role", r.Role)
					return err
				}
				p.Roles[i].Vars = vars
			}
		}
		pb.Play[i] = p
	}
	return nil
}

// convertIncludeTasks from file to blocks
func convertIncludeTasks(baseFS fs.FS, pbPath string, pb *kkcorev1.Playbook) error {
	var pbBase = filepath.Dir(filepath.Dir(pbPath))
	for _, play := range pb.Play {
		if err := fileToBlock(baseFS, pbBase, play.PreTasks); err != nil {
			klog.V(4).ErrorS(err, "Convert pre_tasks error", "playbook", pbPath)
			return err
		}
		if err := fileToBlock(baseFS, pbBase, play.Tasks); err != nil {
			klog.V(4).ErrorS(err, "Convert tasks error", "playbook", pbPath)
			return err
		}
		if err := fileToBlock(baseFS, pbBase, play.PostTasks); err != nil {
			klog.V(4).ErrorS(err, "Convert post_tasks error", "playbook", pbPath)
			return err
		}

		for _, r := range play.Roles {
			roleBase := project.GetRoleBaseFromPlaybook(baseFS, pbPath, r.Role)
			if err := fileToBlock(baseFS, filepath.Join(roleBase, _const.ProjectRolesTasksDir), r.Block); err != nil {
				klog.V(4).ErrorS(err, "Convert role error", "playbook", pbPath, "role", r.Role)
				return err
			}
		}
	}
	return nil
}

func fileToBlock(baseFS fs.FS, baseDir string, blocks []kkcorev1.Block) error {
	for i, b := range blocks {
		if b.IncludeTasks != "" {
			data, err := fs.ReadFile(baseFS, filepath.Join(baseDir, b.IncludeTasks))
			if err != nil {
				klog.V(4).ErrorS(err, "Read includeTask file error", "name", b.Name, "file_path", filepath.Join(baseDir, b.IncludeTasks))
				return err
			}
			var bs []kkcorev1.Block
			if err := yaml.Unmarshal(data, &bs); err != nil {
				klog.V(4).ErrorS(err, "Unmarshal  includeTask data error", "name", b.Name, "file_path", filepath.Join(baseDir, b.IncludeTasks))
				return err
			}
			b.Block = bs
			blocks[i] = b
		}
		if err := fileToBlock(baseFS, baseDir, b.Block); err != nil {
			klog.V(4).ErrorS(err, "Convert block error", "name", b.Name)
			return err
		}
		if err := fileToBlock(baseFS, baseDir, b.Rescue); err != nil {
			klog.V(4).ErrorS(err, "Convert rescue error", "name", b.Name)
			return err
		}
		if err := fileToBlock(baseFS, baseDir, b.Always); err != nil {
			klog.V(4).ErrorS(err, "Convert always error", "name", b.Name)
			return err
		}
	}
	return nil
}

// MarshalBlock marshal block to task
func MarshalBlock(ctx context.Context, role string, hosts []string, when []string, block kkcorev1.Block) *kubekeyv1alpha1.Task {
	task := &kubekeyv1alpha1.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: kubekeyv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
			Annotations: map[string]string{
				kubekeyv1alpha1.TaskAnnotationRole: role,
			},
		},
		Spec: kubekeyv1alpha1.KubeKeyTaskSpec{
			Name:        block.Name,
			Hosts:       hosts,
			IgnoreError: block.IgnoreErrors,
			Retries:     block.Retries,
			When:        when,
			FailedWhen:  block.FailedWhen.Data,
			Register:    block.Register,
		},
	}
	if block.Loop != nil {
		data, err := json.Marshal(block.Loop)
		if err != nil {
			klog.V(4).ErrorS(err, "Marshal loop failed", "task", task.Name, "block", block.Name)
		}
		task.Spec.Loop = runtime.RawExtension{Raw: data}
	}

	return task
}

// GroupHostBySerial group hosts by serial
func GroupHostBySerial(hosts []string, serial []any) ([][]string, error) {
	if len(serial) == 0 {
		return [][]string{hosts}, nil
	}
	result := make([][]string, 0)
	sp := 0
	for _, a := range serial {
		switch a.(type) {
		case int:
			if sp+a.(int) > len(hosts)-1 {
				result = append(result, hosts[sp:])
				return result, nil
			}
			result = append(result, hosts[sp:sp+a.(int)])
			sp += a.(int)
		case string:
			if strings.HasSuffix(a.(string), "%") {
				b, err := strconv.Atoi(strings.TrimSuffix(a.(string), "%"))
				if err != nil {
					klog.V(4).ErrorS(err, "Convert serial to int failed", "serial", a.(string))
					return nil, err
				}
				if sp+int(math.Ceil(float64(len(hosts)*b)/100.0)) > len(hosts)-1 {
					result = append(result, hosts[sp:])
					return result, nil
				}
				result = append(result, hosts[sp:sp+int(math.Ceil(float64(len(hosts)*b)/100.0))])
				sp += int(math.Ceil(float64(len(hosts)*b) / 100.0))
			} else {
				b, err := strconv.Atoi(a.(string))
				if err != nil {
					klog.V(4).ErrorS(err, "Convert serial to int failed", "serial", a.(string))
					return nil, err
				}
				if sp+b > len(hosts)-1 {
					result = append(result, hosts[sp:])
					return result, nil
				}
				result = append(result, hosts[sp:sp+b])
				sp += b
			}
		default:
			return nil, fmt.Errorf("unknown serial type. only support int or percent")
		}
	}
	// if serial is not match all hosts. use last serial
	if sp < len(hosts) {
		a := serial[len(serial)-1]
		for {
			switch a.(type) {
			case int:
				if sp+a.(int) > len(hosts)-1 {
					result = append(result, hosts[sp:])
					return result, nil
				}
				result = append(result, hosts[sp:sp+a.(int)])
				sp += a.(int)
			case string:
				if strings.HasSuffix(a.(string), "%") {
					b, err := strconv.Atoi(strings.TrimSuffix(a.(string), "%"))
					if err != nil {
						klog.V(4).ErrorS(err, "Convert serial to int failed", "serial", a.(string))
						return nil, err
					}
					if sp+int(math.Ceil(float64(len(hosts)*b)/100.0)) > len(hosts)-1 {
						result = append(result, hosts[sp:])
						return result, nil
					}
					result = append(result, hosts[sp:sp+int(math.Ceil(float64(len(hosts)*b)/100.0))])
					sp += int(math.Ceil(float64(len(hosts)*b) / 100.0))
				} else {
					b, err := strconv.Atoi(a.(string))
					if err != nil {
						klog.V(4).ErrorS(err, "Convert serial to int failed", "serial", a.(string))
						return nil, err
					}
					if sp+b > len(hosts)-1 {
						result = append(result, hosts[sp:])
						return result, nil
					}
					result = append(result, hosts[sp:sp+b])
					sp += b
				}
			default:
				return nil, fmt.Errorf("unknown serial type. only support int or percent")
			}
		}
	}
	return result, nil
}

// CalculatePipelineStatus calculate pipeline status from tasks
func CalculatePipelineStatus(nsTasks *kubekeyv1alpha1.TaskList, pipeline *kubekeyv1.Pipeline) {
	if pipeline.Status.Phase != kubekeyv1.PipelinePhaseRunning {
		// only deal running pipeline
		return
	}
	pipeline.Status.TaskResult = kubekeyv1.PipelineTaskResult{
		Total: len(nsTasks.Items),
	}
	var failedDetail []kubekeyv1.PipelineFailedDetail
	for _, t := range nsTasks.Items {
		switch t.Status.Phase {
		case kubekeyv1alpha1.TaskPhaseSuccess:
			pipeline.Status.TaskResult.Success++
		case kubekeyv1alpha1.TaskPhaseIgnored:
			pipeline.Status.TaskResult.Ignored++
		case kubekeyv1alpha1.TaskPhaseSkipped:
			pipeline.Status.TaskResult.Skipped++
		}
		if t.Status.Phase == kubekeyv1alpha1.TaskPhaseFailed && t.Spec.Retries <= t.Status.RestartCount {
			var hostReason []kubekeyv1.PipelineFailedDetailHost
			for _, tr := range t.Status.FailedDetail {
				hostReason = append(hostReason, kubekeyv1.PipelineFailedDetailHost{
					Host:   tr.Host,
					Stdout: tr.Stdout,
					StdErr: tr.StdErr,
				})
			}
			failedDetail = append(failedDetail, kubekeyv1.PipelineFailedDetail{
				Task:  t.Name,
				Hosts: hostReason,
			})
			pipeline.Status.TaskResult.Failed++
		}
	}

	if pipeline.Status.TaskResult.Failed != 0 {
		pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
		pipeline.Status.Reason = "task failed"
		pipeline.Status.FailedDetail = failedDetail
	} else if pipeline.Status.TaskResult.Total == pipeline.Status.TaskResult.Success+pipeline.Status.TaskResult.Ignored+pipeline.Status.TaskResult.Skipped {
		pipeline.Status.Phase = kubekeyv1.PipelinePhaseSucceed
	}

}

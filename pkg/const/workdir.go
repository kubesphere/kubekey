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

package _const

/** a kubekey workdir like that:
workdir/
|-- projects/
|   |-- ansible-project1/
|   |   |-- playbooks/
|   |   |-- roles/
|   |   |   |-- roleName/
|   |   |   |   |-- tasks/
|   |   |   |   |   |-- main.yml
|   |   |   |   |-- defaults/
|   |   |   |   |   |-- main.yml
|   |   |   |   |-- templates/
|   |   |   |   |-- files/
|   |
|   |-- ansible-project2/
|   |-- ...
|
|-- runtime/
|-- group/version/
|   |   |-- pipelines/
|   |   |   |-- namespace/
|   |   |   |   |-- pipeline.yaml
|   |   |   |   |-- /pipelineName/variable/
|   |   |   |   |   |-- location.json
|   |   |   |   |   |-- hostname.json
|   |   |-- tasks/
|   |   |   |-- namespace/
|   |   |   |   |-- task.yaml
|   |   |-- configs/
|   |   |   |-- namespace/
|   |   |   |   |   |-- config.yaml
|   |   |-- inventories/
|   |   |   |-- namespace/
|   |   |   |   |-- inventory.yaml
|
|-- kubekey/
|-- artifact-path...
|-- images
|
|-- kubernetes/

*/

// workDir is the user-specified working directory. By default, it is the same as the directory where the kubekey command is executed.
var workDir string

// ProjectDir is a fixed directory name under workdir, used to store the Ansible project.
const ProjectDir = "projects"

// ansible-project is the name of different Ansible projects

// ProjectPlaybooksDir is a fixed directory name under ansible-project. used to store executable playbook files.
const ProjectPlaybooksDir = "playbooks"

// ProjectRolesDir is a fixed directory name under ansible-project. used to store roles which playbook need.
const ProjectRolesDir = "roles"

// roleName is the name of different roles

// ProjectRolesTasksDir is a fixed directory name under roleName. used to store task which role need.
const ProjectRolesTasksDir = "tasks"

// ProjectRolesTasksMainFile is a fixed file under tasks. it must run if the role run. support *.yaml or *yml
const ProjectRolesTasksMainFile = "main"

// ProjectRolesDefaultsDir is a fixed directory name under roleName. it set default variables to role.
const ProjectRolesDefaultsDir = "defaults"

// ProjectRolesDefaultsMainFile is a fixed file under defaults. support *.yaml or *yml
const ProjectRolesDefaultsMainFile = "main"

// ProjectRolesTemplateDir is a fixed directory name under roleName. used to store template which task need.
const ProjectRolesTemplateDir = "templates"

// ProjectRolesFilesDir is a fixed directory name under roleName. used to store files which task need.
const ProjectRolesFilesDir = "files"

// RuntimeDir is a fixed directory name under workdir, used to store the runtime data of the current task execution.
const RuntimeDir = "runtime"

// the resources dir store as etcd key.
// like: /prefix/group/version/resource/namespace/name

// RuntimePipelineDir store Pipeline resources
const RuntimePipelineDir = "pipelines"

// pipeline.yaml is the data of Pipeline resource

// RuntimePipelineVariableDir is a fixed directory name under runtime, used to store the task execution parameters.
const RuntimePipelineVariableDir = "variable"

// RuntimePipelineTaskDir is a fixed directory name under runtime, used to store the task execution status.

// task.yaml is the data of Task resource

// RuntimeConfigDir store Config resources

// config.yaml is the data of Config resource

// RuntimeInventoryDir store Inventory resources

// inventory.yaml is the data of Inventory resource

// ArtifactDir is the default directory name under the working directory. It is used to store
// files required when executing the kubekey command (such as: docker, etcd, image packages, etc.).
// These files will be downloaded locally and distributed to remote nodes.
const ArtifactDir = "kubekey"

// artifact-path store artifact package.

// ArtifactImagesDir store images files. contains blobs and manifests.
const ArtifactImagesDir = "images"

// KubernetesDir represents the remote host directory for each kubernetes connection
const KubernetesDir = "kubernetes"

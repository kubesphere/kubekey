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

/**
The structure of a typical KubeKey working directory:
work_dir/
|-- projects_dir/
|   |-- project1/
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
|   |-- project2/
|   |-- ...
|
|-- binary_dir/
|   |-- artifact-path...
|   |-- images
|
|-- scripts_dir/
|
|-- runtime/
|-- group/version/
|   |   |-- playbooks/
|   |   |   |-- namespace/
|   |   |   |   |-- playbook.yaml
|   |   |   |   |-- /playbookName/variable/
|   |   |   |   |   |-- location.json
|   |   |   |   |   |-- hostname.json
|   |   |-- tasks/
|   |   |   |-- namespace/
|   |   |   |   |-- task.yaml
|   |   |-- inventories/
|   |   |   |-- namespace/
|   |   |   |   |-- inventory.yaml
|
|-- kubernetes/

*/

// Workdir is the user-specified working directory. By default, it is the same as the directory where the KubeKey command is executed.
const Workdir = "work_dir"

// ProjectsDir stores runnable projects. By default, its path is set to {{ .work_dir/projects }}.
const ProjectsDir = "projects_dir"

// project represents the name of individual projects.

// ProjectPlaybooksDir is a fixed directory name under a project, used to store executable playbook files.
const ProjectPlaybooksDir = "playbooks"

// ProjectRolesDir is a fixed directory name under a project, used to store roles required by playbooks.
const ProjectRolesDir = "roles"

// roleName represents the name of individual roles.

// ProjectRolesTasksDir is a fixed directory name under a role, used to store tasks required by the role.
const ProjectRolesTasksDir = "tasks"

// ProjectRolesTasksMainFile is a mandatory file under the tasks directory that must be executed when the role is run. It supports files with .yaml or .yml extensions.
const ProjectRolesTasksMainFile = "main"

// ProjectRolesDefaultsDir is a fixed directory name under a role, used to set default variables for the role.
const ProjectRolesDefaultsDir = "defaults"

// ProjectRolesDefaultsMainFile is a mandatory file under the defaults directory. It supports files with .yaml or .yml extensions.
const ProjectRolesDefaultsMainFile = "main"

// ProjectRolesTemplateDir is a fixed directory name under a role, used to store templates required by tasks.
const ProjectRolesTemplateDir = "templates"

// ProjectRolesFilesDir is a fixed directory name under a role, used to store files required by tasks.
const ProjectRolesFilesDir = "files"

// ScriptsDir stores custom scripts. By default, its path is set to {{ .work_dir/scripts }}.
const ScriptsDir = "scripts_dir"

// BinaryDir refers to a portable software package that can typically be bundled into an offline package. By default, its path is set to {{ .work_dir/kubekey }}.
const BinaryDir = "binary_dir"

// BinaryImagesDir stores image files, including blobs and manifests.
const BinaryImagesDir = "images"

// RuntimeDir used to store runtime data for the current task execution. By default, its path is set to {{ .work_dir/runtime }}.
const RuntimeDir = "runtime"

// RuntimePlaybookDir stores playbook resources created during playbook execution.
const RuntimePlaybookDir = "playbooks"

// playbook.yaml contains the data for a playbook resource.

// RuntimePlaybookVariableDir is a fixed directory name under runtime, used to store task execution parameters.
const RuntimePlaybookVariableDir = "variable"

// RuntimePlaybookTaskDir is a fixed directory name under runtime, used to store the task execution status.

// task.yaml contains the data for a task resource.

// RuntimeConfigDir stores configuration resources.

// config.yaml contains the data for a configuration resource.

// RuntimeInventoryDir stores inventory resources.

// inventory.yaml contains the data for an inventory resource.

// KubernetesDir represents the remote host directory for each Kubernetes connection created during playbook execution.
const KubernetesDir = "kubernetes"

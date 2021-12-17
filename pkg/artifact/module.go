/*
 Copyright 2021 The KubeSphere Authors.

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

package artifact

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type ImagesModule struct {
	common.ArtifactModule
}

func (i *ImagesModule) Init() {
	i.Name = "ArtifactImagesModule"
	i.Desc = "Export images on the localhost"

	check := &task.LocalTask{
		Name:   "CheckContainerd",
		Desc:   "Try to new a containerd client",
		Action: new(CheckContainerd),
	}

	pull := &task.LocalTask{
		Name:   "ArtifactPullImages",
		Desc:   "Pull images on the localhost",
		Action: new(PullImages),
	}

	export := &task.LocalTask{
		Name:   "ArtifactExportImages",
		Desc:   "Export images on the localhost",
		Action: new(ExportImages),
	}

	closeClient := &task.LocalTask{
		Name:   "CloseContainerdClient",
		Desc:   "Delete containerd client from cache and close it",
		Action: new(CloseClient),
	}

	i.Tasks = []task.Interface{
		check,
		pull,
		export,
		closeClient,
	}
}

type RepositoryModule struct {
	common.ArtifactModule
}

func (r *RepositoryModule) Init() {
	r.Name = "RepositoryModule"
	r.Desc = "Get OS repository ISO file"

	download := &task.LocalTask{
		Name:    "DownloadISOFile",
		Desc:    "Download iso file into work dir",
		Prepare: new(EnableDownload),
		Action:  new(DownloadISOFile),
	}

	localCopy := &task.LocalTask{
		Name:   "LocalCopy",
		Desc:   "Copy local iso file into artifact dir",
		Action: new(LocalCopy),
	}

	r.Tasks = []task.Interface{
		download,
		localCopy,
	}
}

type ArchiveModule struct {
	common.ArtifactModule
}

func (a *ArchiveModule) Init() {
	a.Name = "ArtifactArchiveModule"
	a.Desc = "Archive the dependencies"

	archive := &task.LocalTask{
		Name:   "ArchiveDependencies",
		Desc:   "Archive the dependencies",
		Action: new(ArchiveDependencies),
	}

	a.Tasks = []task.Interface{
		archive,
	}
}

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
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/task"
)

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

type UnArchiveModule struct {
	common.KubeModule
	Skip bool
}

func (u *UnArchiveModule) IsSkip() bool {
	return u.Skip
}

func (u *UnArchiveModule) Init() {
	u.Name = "UnArchiveArtifactModule"
	u.Desc = "UnArchive the KubeKey artifact"

	md5Check := &task.LocalTask{
		Name:   "CheckArtifactMd5",
		Desc:   "Check the KubeKey artifact md5 value",
		Action: new(Md5Check),
	}

	unArchive := &task.LocalTask{
		Name:    "UnArchiveArtifact",
		Desc:    "UnArchive the KubeKey artifact",
		Prepare: &Md5AreEqual{Not: true},
		Action:  new(UnArchive),
	}

	createMd5File := &task.LocalTask{
		Name:    "CreateArtifactMd5File",
		Desc:    "Create the KubeKey artifact Md5 file",
		Prepare: &Md5AreEqual{Not: true},
		Action:  new(CreateMd5File),
	}

	u.Tasks = []task.Interface{
		md5Check,
		unArchive,
		createMd5File,
	}
}

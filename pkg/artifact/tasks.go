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
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images/archive"
	"github.com/containerd/containerd/namespaces"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	coreutil "github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CheckContainerd struct {
	common.ArtifactAction
}

func (c *CheckContainerd) Execute(_ connector.Runtime) error {
	client, err := containerd.New(c.Manifest.Arg.CriSocket)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("new a containerd client failed"))
	}

	c.PipelineCache.Set(common.ContainerdClient, client)
	return nil
}

type PullImages struct {
	common.ArtifactAction
}

func (p *PullImages) Execute(_ connector.Runtime) error {
	c, ok := p.PipelineCache.Get(common.ContainerdClient)
	if !ok {
		return errors.New("get containerd client failed by pipeline client")
	}
	client := c.(*containerd.Client)

	ctx := namespaces.WithNamespace(context.Background(), "kubekey")
	for _, image := range p.Manifest.Spec.Images {
		var err error
		for i := 0; i < 3; i++ {
			if _, e := client.Fetch(ctx, image); e != nil {
				err = e
				continue
			}
			logger.Log.Messagef(common.LocalHost, "pull image %s success", image)
			break
		}
		if err != nil {
			return errors.Wrapf(errors.WithStack(err), "pull image %s failed", image)
		}
		//todo: Whether it need to be unpack?
	}

	return nil
}

type ExportImages struct {
	common.ArtifactAction
}

func (e *ExportImages) Execute(runtime connector.Runtime) error {
	c, ok := e.PipelineCache.Get(common.ContainerdClient)
	if !ok {
		return errors.New("get containerd client failed by pipeline client")
	}
	client := c.(*containerd.Client)

	dir := filepath.Join(runtime.GetWorkDir(), common.Artifact, "images")
	if err := coreutil.Mkdir(dir); err != nil {
		return errors.Wrapf(errors.WithStack(err), "mkdir %s failed", dir)
	}

	ctx := namespaces.WithNamespace(context.Background(), "kubekey")
	is := client.ImageService()
	for _, image := range e.Manifest.Spec.Images {
		fileName := strings.ReplaceAll(image, "/", "-")
		fileName = strings.ReplaceAll(fileName, ":", "-")
		fileName = fmt.Sprintf("%s.tar", fileName)

		filePath := filepath.Join(dir, fileName)
		if coreutil.IsExist(filePath) {
			logger.Log.Messagef(common.LocalHost, "%s is existed", fileName)
			continue
		}

		w, err := os.Create(filePath)
		if err != nil {
			return errors.Wrapf(errors.WithStack(err), "create image tar file %s failed", filePath)
		}
		defer w.Close()

		if err := client.Export(ctx, w, archive.WithImage(is, image)); err != nil {
			return errors.Wrapf(errors.WithStack(err), "export image %s failed", image)
		}
		logger.Log.Messagef(common.LocalHost, "export image %s as %s success", image, fileName)
	}
	return nil
}

type CloseClient struct {
	common.ArtifactAction
}

func (c *CloseClient) Execute(_ connector.Runtime) error {
	cl, ok := c.PipelineCache.Get(common.ContainerdClient)
	if !ok {
		return errors.New("get containerd client failed by pipeline client")
	}
	client := cl.(*containerd.Client)
	defer client.Close()

	c.PipelineCache.Delete(common.ContainerdClient)
	return nil
}

type DownloadISOFile struct {
	common.ArtifactAction
}

func (d *DownloadISOFile) Execute(runtime connector.Runtime) error {
	for i, sys := range d.Manifest.Spec.OperationSystems {
		if sys.Repository.Iso.Url == "" {
			continue
		}

		fileName := fmt.Sprintf("%s-%s-%s.iso", sys.Id, sys.Version, sys.Arch)
		filePath := filepath.Join(runtime.GetWorkDir(), fileName)
		getCmd := d.Manifest.Arg.DownloadCommand(filePath, sys.Repository.Iso.Url)

		cmd := exec.Command("/bin/sh", "-c", getCmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Failed to download %s iso file: %s error: %w ", fileName, getCmd, err)
		}
		cmd.Stderr = cmd.Stdout

		if err = cmd.Start(); err != nil {
			return fmt.Errorf("Failed to download %s iso file: %s error: %w ", fileName, getCmd, err)
		}
		for {
			tmp := make([]byte, 1024)
			_, err := stdout.Read(tmp)
			fmt.Print(string(tmp))
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				logger.Log.Errorln(err)
				break
			}
		}
		if err = cmd.Wait(); err != nil {
			return fmt.Errorf("Failed to download %s iso file: %s error: %w ", fileName, getCmd, err)
		}
		d.Manifest.Spec.OperationSystems[i].Repository.Iso.LocalPath = filePath
	}
	return nil
}

type LocalCopy struct {
	common.ArtifactAction
}

func (l *LocalCopy) Execute(runtime connector.Runtime) error {
	for _, sys := range l.Manifest.Spec.OperationSystems {
		dir := filepath.Join(runtime.GetWorkDir(), common.Artifact, "repository", sys.Arch, sys.Id, sys.Version)
		if err := coreutil.Mkdir(dir); err != nil {
			return errors.Wrapf(errors.WithStack(err), "mkdir %s failed", dir)
		}

		if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo cp -f %s %s", sys.Repository.Iso.LocalPath, dir)).Run(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "copy %s to %s failed", sys.Repository.Iso.LocalPath, dir)
		}
	}

	return nil
}

type ArchiveDependencies struct {
	common.ArtifactAction
}

func (a *ArchiveDependencies) Execute(runtime connector.Runtime) error {
	src := filepath.Join(runtime.GetWorkDir(), common.Artifact)
	if err := coreutil.Tar(src, a.Manifest.Arg.Output, runtime.GetWorkDir()); err != nil {
		return errors.Wrapf(errors.WithStack(err), "archive %s failed", src)
	}
	return nil
}

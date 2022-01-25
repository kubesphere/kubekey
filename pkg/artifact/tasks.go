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
	"github.com/containerd/containerd/cmd/ctr/commands/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/images/archive"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	coreutil "github.com/kubesphere/kubekey/pkg/core/util"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
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

	// k: registry name "registry-1.docker.io", v: registryAuth
	auths := Auths(p.Manifest)
	resolvers := make(map[string]remotes.Resolver)
	for repo, auth := range auths {
		resolvers[repo] = GetResolver(ctx, auth)
	}

	for _, image := range p.Manifest.Spec.Images {
		progress := make(chan struct{})
		lctx, done, err := client.WithLease(ctx)
		if err != nil {
			return err
		}
		defer done(ctx)

		ongoing := content.NewJobs(image)
		pctx, stopProgress := context.WithCancel(lctx)

		go func() {
			content.ShowProgress(pctx, ongoing, client.ContentStore(), os.Stdout)
			close(progress)
		}()

		h := images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
			if desc.MediaType != images.MediaTypeDockerSchema1Manifest {
				ongoing.Add(desc)
			}
			return nil, nil
		})

		opts := []containerd.RemoteOpt{
			containerd.WithImageHandler(h),
			containerd.WithSchema1Conversion,
		}

		if resolver, ok := resolvers[strings.Split(image, "/")[0]]; ok {
			opts = append(opts, containerd.WithResolver(resolver))
		} else {
			opts = append(opts, containerd.WithResolver(GetResolver(ctx, registryAuth{})))
		}

		for _, arch := range p.Manifest.Spec.Arches {
			opts = append(opts, containerd.WithPlatform(arch))
		}

		_, err = client.Fetch(ctx, image, opts...)
		stopProgress()
		if err != nil {
			logger.Log.Messagef(common.LocalHost, "pull image %s failed", image)
			return errors.Wrapf(errors.WithStack(err), "pull image %s failed", image)
		}

		<-progress
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

		exportOpts := []archive.ExportOpt{
			archive.WithPlatform(platforms.Default()),
			archive.WithImage(is, image),
		}

		var all []ocispec.Platform
		for _, arch := range e.Manifest.Spec.Arches {
			p, err := platforms.Parse(arch)
			if err != nil {
				return errors.Wrapf(err, "invalid platform %q", arch)
			}
			all = append(all, p)
		}
		exportOpts = append(exportOpts, archive.WithPlatform(platforms.Ordered(all...)))

		if err := client.Export(ctx, w, exportOpts...); err != nil {
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
		if sys.Repository.Iso.LocalPath == "" {
			continue
		}

		dir := filepath.Join(runtime.GetWorkDir(), common.Artifact, "repository", sys.Arch, sys.Id, sys.Version)
		if err := coreutil.Mkdir(dir); err != nil {
			return errors.Wrapf(errors.WithStack(err), "mkdir %s failed", dir)
		}

		path := filepath.Join(dir, fmt.Sprintf("%s-%s-%s.iso", sys.Id, sys.Version, sys.Arch))
		if err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo cp -f %s %s", sys.Repository.Iso.LocalPath, path)).Run(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "copy %s to %s failed", sys.Repository.Iso.LocalPath, path)
		}
	}

	return nil
}

type ArchiveDependencies struct {
	common.ArtifactAction
}

func (a *ArchiveDependencies) Execute(runtime connector.Runtime) error {
	src := filepath.Join(runtime.GetWorkDir(), common.Artifact)
	if err := coreutil.Tar(src, a.Manifest.Arg.Output, src); err != nil {
		return errors.Wrapf(errors.WithStack(err), "archive %s failed", src)
	}
	return nil
}

type UnArchive struct {
	common.KubeAction
}

func (u *UnArchive) Execute(runtime connector.Runtime) error {
	if err := coreutil.Untar(u.KubeConf.Arg.Artifact, runtime.GetWorkDir()); err != nil {
		return errors.Wrapf(errors.WithStack(err), "unArchive %s failed", u.KubeConf.Arg.Artifact)
	}
	return nil
}

type Md5Check struct {
	common.KubeAction
}

func (m *Md5Check) Execute(runtime connector.Runtime) error {
	m.ModuleCache.Set("md5AreEqual", false)

	// check if there is a md5.sum file. This file's content contains the last artifact md5 value.
	oldFile := filepath.Join(runtime.GetWorkDir(), "artifact.md5")
	if exist := coreutil.IsExist(oldFile); !exist {
		return nil
	}

	oldMd5, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return errors.Wrapf(errors.WithStack(err), "read old md5 file %s failed", oldFile)
	}

	newMd5 := coreutil.LocalMd5Sum(m.KubeConf.Arg.Artifact)

	if string(oldMd5) == newMd5 {
		m.ModuleCache.Set("md5AreEqual", true)
	}
	return nil
}

type CreateMd5File struct {
	common.KubeAction
}

func (c *CreateMd5File) Execute(runtime connector.Runtime) error {
	oldFile := filepath.Join(runtime.GetWorkDir(), "artifact.md5")
	newMd5 := coreutil.LocalMd5Sum(c.KubeConf.Arg.Artifact)
	f, err := os.Create(oldFile)
	if err != nil {
		return errors.Wrapf(errors.WithStack(err), "create md5 fild %s failed", oldFile)
	}

	if _, err := io.Copy(f, strings.NewReader(newMd5)); err != nil {
		return errors.Wrapf(errors.WithStack(err), "write md5 value to file %s failed", oldFile)
	}
	return nil
}

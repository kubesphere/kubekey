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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	coreutil "github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/files"
)

type DownloadISOFile struct {
	common.ArtifactAction
}

func (d *DownloadISOFile) Execute(runtime connector.Runtime) error {
	for i, sys := range d.Manifest.Spec.OperatingSystems {
		if sys.Repository.Iso.Url == "" {
			continue
		}

		fileName := fmt.Sprintf("%s-%s-%s.iso", sys.Id, sys.Version, sys.Arch)
		filePath := filepath.Join(runtime.GetWorkDir(), fileName)

		checksumEqual, err := files.SHA256CheckEqual(filePath, sys.Repository.Iso.Checksum)
		if err != nil {
			return err
		}
		if checksumEqual {
			logger.Log.Infof("Skip download exists iso file %s", fileName)
			continue
		}

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
		d.Manifest.Spec.OperatingSystems[i].Repository.Iso.LocalPath = filePath
	}
	return nil
}

type LocalCopy struct {
	common.ArtifactAction
}

func (l *LocalCopy) Execute(runtime connector.Runtime) error {
	for _, sys := range l.Manifest.Spec.OperatingSystems {
		if sys.Repository.Iso.LocalPath == "" {
			continue
		}

		dir := filepath.Join(runtime.GetWorkDir(), common.Artifact, "repository", sys.Arch, sys.Id, sys.Version)
		if err := coreutil.Mkdir(dir); err != nil {
			return errors.Wrapf(errors.WithStack(err), "mkdir %s failed", dir)
		}

		path := filepath.Join(dir, fmt.Sprintf("%s-%s-%s.iso", sys.Id, sys.Version, sys.Arch))
		if out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo cp -f %s %s", sys.Repository.Iso.LocalPath, path)).CombinedOutput(); err != nil {
			return errors.Errorf("copy %s to %s failed: %s", sys.Repository.Iso.LocalPath, path, string(out))
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

	// skip remove artifact if --skip-remove-artifact
	if a.Manifest.Arg.SkipRemoveArtifact {
		return nil
	}

	// remove the src directory
	if err := os.RemoveAll(src); err != nil {
		return errors.Wrapf(errors.WithStack(err), "remove %s failed", src)
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

	oldMd5, err := os.ReadFile(oldFile)
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

/*
 Copyright 2022 The KubeSphere Authors.

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

package file

import (
	"fmt"
	"path/filepath"
	"strings"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/pkg/rootfs"
	"github.com/kubesphere/kubekey/pkg/service/operation/file/checksum"
	"github.com/kubesphere/kubekey/pkg/util/osrelease"
)

// ISO info
const (
	ISOName        = "%s.iso"
	ISOID          = "iso"
	ISOURLPathTmpl = "/kubesphere/kubekey/releases/download/v2.2.2/%s"
)

// ISO is a Binary for repository ISO file.
type ISO struct {
	*Binary
	HTTPChecksum *checksum.HTTPChecksum
}

// NewISO returns a new repository ISO.
func NewISO(sshClient ssh.Interface, rootFs rootfs.Interface, os *osrelease.Data, arch string, isoName string) (*ISO, error) {
	var (
		fileName string
		urlPath  string
	)

	if isoName == infrav1.AUTO {
		fileName = generateISOName(os, arch)
		if fileName == "" {
			return nil, fmt.Errorf("" +
				"can not detect the ISO file automatically, only support Ubuntu/Debian/CentOS. " +
				"Please specify the ISO file name in the '.spec.repository.iso' field")
		}
		urlPath = fmt.Sprintf(ISOURLPathTmpl, fileName)
	} else {
		fileName = isoName
		urlPath = fmt.Sprintf(ISOURLPathTmpl, fileName)
	}

	file, err := NewFile(Params{
		SSHClient:      sshClient,
		RootFs:         rootFs,
		Type:           FileBinary,
		Name:           fileName,
		LocalFullPath:  filepath.Join(rootFs.ClusterRootFsDir(), fileName),
		RemoteFullPath: filepath.Join(MntDir, fileName),
	})
	if err != nil {
		return nil, err
	}

	u := parseURL(DefaultDownloadHost, urlPath)
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      os.ID,
		Version: os.VersionID,
		Arch:    arch,
		URL:     u,
	})

	checksumURL := parseURL(DefaultDownloadHost, fmt.Sprintf(ISOURLPathTmpl, generateCheckFileName(binary.Name())))
	httpChecksum := checksum.NewHTTPChecksum(checksumURL, binary.Name(), binary.file.rootFs)
	binary.AppendChecksum(httpChecksum)

	return &ISO{binary, httpChecksum}, nil
}

func generateISOName(os *osrelease.Data, arch string) string {
	var fileName string
	switch os.ID {
	case osrelease.UbuntuID:
		fileName = fmt.Sprintf(ISOName, strings.Join([]string{os.ID, os.VersionID, "debs", arch}, "-"))
	case osrelease.DebianID:
		fileName = fmt.Sprintf(ISOName, strings.Join([]string{os.ID + os.VersionID, "debs", arch}, "-"))
	case osrelease.CentosID:
		fileName = fmt.Sprintf(ISOName, strings.Join([]string{os.ID + os.VersionID, "rpms", arch}, "-"))
	default:
		fileName = fmt.Sprintf(ISOName, strings.Join([]string{os.ID, os.VersionID, arch}, "-"))
	}
	return fileName
}

func generateCheckFileName(isoName string) string {
	return isoName[0:strings.LastIndex(isoName, "-")] + ".iso.sha256sum.txt"
}

// SetZone override Binary's SetZone method.
func (i *ISO) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		return
	}
}

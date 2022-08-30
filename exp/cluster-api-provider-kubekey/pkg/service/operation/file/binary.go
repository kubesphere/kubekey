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
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file/checksum"
)

// Default
const (
	ZONE                        = "cn"
	DefaultDownloadHost         = "https://github.com"
	DefaultDownloadHostGoogle   = "https://storage.googleapis.com"
	DefaultDownloadHostQingStor = "https://kubernetes-release.pek3b.qingstor.com"
)

// BinaryParams represents the parameters of a Binary.
type BinaryParams struct {
	File     *File
	ID       string
	Version  string
	Arch     string
	URL      *url.URL
	Checksum checksum.Interface
}

// NewBinary returns a new Binary.
func NewBinary(params BinaryParams) *Binary {
	b := &Binary{
		file:     params.File,
		id:       params.ID,
		version:  params.Version,
		arch:     params.Arch,
		url:      params.URL,
		checksum: params.Checksum,
	}
	return b
}

// Binary is a binary implementation of Binary interface.
type Binary struct {
	file     *File
	id       string
	version  string
	arch     string
	url      *url.URL
	checksum checksum.Interface
}

// Name returns the name of the Binary file.
func (b *Binary) Name() string {
	return b.file.Name()
}

// Type returns the type of the Binary file.
func (b *Binary) Type() Type {
	return b.file.Type()
}

// LocalPath returns the local path of the Binary file.
func (b *Binary) LocalPath() string {
	return b.file.LocalPath()
}

// RemotePath returns the remote path of the Binary file.
func (b *Binary) RemotePath() string {
	return b.file.RemotePath()
}

// LocalExist returns true if the Binary file is existed in the local path.
func (b *Binary) LocalExist() bool {
	return b.file.LocalExist()
}

// RemoteExist returns true if the Binary file is existed (and the SHA256 check passes) in the remote path.
func (b *Binary) RemoteExist() bool {
	if !b.file.RemoteExist() {
		return false
	}

	cmd := fmt.Sprintf("sha256sum %s | cut -d\" \" -f1", b.file.RemotePath())
	remoteSHA256, err := b.file.sshClient.SudoCmd(cmd)
	if err != nil {
		return false
	}

	if err := b.checksum.Get(); err != nil {
		return false
	}

	if remoteSHA256 != b.checksum.Value() {
		return false
	}
	return true
}

// Copy copies the Binary file from the local path to the remote path.
func (b *Binary) Copy(override bool) error {
	return b.file.Copy(override)
}

// Fetch copies the Binary file from the remote path to the local path.
func (b *Binary) Fetch(override bool) error {
	return b.file.Fetch(override)
}

// Chmod changes the mode of the Binary file.
func (b *Binary) Chmod(option string) error {
	return b.file.Chmod(option)
}

// ID returns the id of the binary.
func (b *Binary) ID() string {
	return b.id
}

// Arch returns the arch of the binary.
func (b *Binary) Arch() string {
	return b.arch
}

// Version returns the version of the binary.
func (b *Binary) Version() string {
	return b.version
}

// URL returns the download url of the binary.
func (b *Binary) URL() *url.URL {
	return b.url
}

// SetURL sets the download url of the binary.
func (b *Binary) SetURL(urlStr string) {
	if urlStr == "" {
		return
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	b.url = u
}

// SetHost sets the host to download the binaries.
func (b *Binary) SetHost(host string) {
	if host == "" {
		return
	}
	u, err := url.Parse(host)
	if err != nil {
		return
	}
	u.Path = b.url.Path
	b.url = u
}

// SetPath sets the URL path of the binary.
func (b *Binary) SetPath(pathStr string) {
	if pathStr == "" {
		return
	}
	ref, err := url.Parse(pathStr)
	if err != nil {
		return
	}
	b.url = b.url.ResolveReference(ref)
}

// SetZone sets the zone of the binary.
func (b *Binary) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		b.SetHost(DefaultDownloadHostQingStor)
	}
}

// SetChecksum sets the checksum of the binary.
func (b *Binary) SetChecksum(c checksum.Interface) {
	if reflect.ValueOf(c).IsNil() {
		return
	}
	b.checksum = c
}

// Get downloads the binary from remote.
func (b *Binary) Get(timeout time.Duration) error {
	client := &getter.HttpGetter{
		ReadTimeout: timeout,
	}

	if err := client.GetFile(b.LocalPath(), b.URL()); err != nil {
		return errors.Wrapf(err, "failed to http get file: %s", b.LocalPath())
	}

	return nil
}

// SHA256 calculates the SHA256 of the binary.
func (b *Binary) SHA256() (string, error) {
	f, err := os.Open(b.LocalPath())
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// CompareChecksum compares the checksum of the binary.
func (b *Binary) CompareChecksum() error {
	if err := b.checksum.Get(); err != nil {
		return errors.Wrapf(err, "%s get checksum failed", b.Name())
	}

	sum, err := b.SHA256()
	if err != nil {
		return errors.Wrapf(err, "%s caculate SHA256 failed", b.Name())
	}

	if sum != b.checksum.Value() {
		return errors.Errorf("SHA256 no match. file: %s sha256: %s not equal checksum: %s", b.Name(), sum, b.checksum.Value())
	}
	return nil
}

func parseURL(host, pathStr string) *url.URL {
	u, _ := url.Parse(host)
	u.Path = path.Join(u.Path, pathStr)
	return u
}

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
	"os"
	"time"

	"github.com/hashicorp/go-getter"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file/checksum"
)

// Binary is a binary implementation of Binary interface.
type Binary struct {
	*File
	id       string
	version  string
	arch     string
	url      string
	cnURL    string
	checksum checksum.Interface
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
func (b *Binary) URL() string {
	return b.url
}

// SetZone sets the zone of the binary.
func (b *Binary) SetZone(zone string) {
	if zone == "cn" {
		b.url = b.cnURL
	}
}

// Get downloads the binary from remote.
func (b *Binary) Get(timeout time.Duration) error {
	//todo: should not to skip TLS verify
	client := &getter.HttpGetter{
		ReadTimeout: timeout,
		//Client: &http.Client{
		//	Transport: &http.Transport{
		//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//	},
		//},
	}

	url, err := urlhelper.Parse(b.url)
	if err != nil {
		return errors.Wrapf(err, "failed to parse url: %s", b.url)
	}

	if err := client.GetFile(b.LocalPath(), url); err != nil {
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
		return errors.New(fmt.Sprintf("SHA256 no match. file: %s sha256: %s not equal checksum: %s", b.Name(), sum, b.checksum.Value()))
	}
	return nil
}

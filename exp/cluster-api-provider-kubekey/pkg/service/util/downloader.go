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

package util

import (
	"time"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file/checksum"
)

// DownloadAndCopy downloads and copies files to the remote instance.
func DownloadAndCopy(b operation.Binary, zone, host, path, url, checksumStr string, timeout time.Duration) error {
	if b.RemoteExist() {
		return nil
	}

	b.SetChecksum(checksum.NewStringChecksum(checksumStr))
	if !(b.LocalExist() && b.CompareChecksum() == nil) {
		// Only the host is an empty string, we can set up the zone.
		// Because the URL path which in the QingStor is not the same as the default.
		if host == "" {
			b.SetZone(zone)
		}

		// Always try to set the "host, path, url, checksum".
		// If the these vars are empty strings, it will not make any changes.
		b.SetHost(host)
		b.SetPath(path)
		b.SetURL(url)

		if err := b.Get(timeout); err != nil {
			return err
		}
		if err := b.CompareChecksum(); err != nil {
			return err
		}
	}

	if err := b.Copy(true); err != nil {
		return err
	}
	return nil
}

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

package checksum

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/pkg/rootfs"
)

// HTTPChecksum is a checksum that is downloaded from a URL.
type HTTPChecksum struct {
	fs       rootfs.Interface
	url      *url.URL
	FileName string
	value    string
}

// NewHTTPChecksum returns a new HTTPChecksum.
func NewHTTPChecksum(url *url.URL, isoName string, fs rootfs.Interface) *HTTPChecksum {
	return &HTTPChecksum{
		url:      url,
		FileName: isoName,
		fs:       fs,
	}
}

// SetHost sets the host of the URL.
func (h *HTTPChecksum) SetHost(host string) {
	if host == "" {
		return
	}
	u, err := url.Parse(host)
	if err != nil {
		return
	}
	u.Path = h.url.Path
	h.url = u
}

// SetPath sets the URL path of the binary.
func (h *HTTPChecksum) SetPath(pathStr string) {
	if pathStr == "" {
		return
	}
	ref, err := url.Parse(pathStr)
	if err != nil {
		return
	}

	h.url = h.url.ResolveReference(ref)
}

// Get downloads the checksum file and parses it.
func (h *HTTPChecksum) Get() error {
	tempfile, err := h.fs.Fs().MkLocalTmpFile(h.fs.ClusterRootFsDir(), filepath.Base(h.url.Path))
	if err != nil {
		return err
	}
	defer func() {
		_ = h.fs.Fs().RemoveAll(tempfile)
	}()

	client := &getter.HttpGetter{
		ReadTimeout: 15 * time.Second,
	}

	if err := client.GetFile(tempfile, h.url); err != nil {
		return errors.Wrapf(err, "failed to http get file: %s", h.url)
	}

	f, err := os.Open(filepath.Clean(tempfile))
	if err != nil {
		return fmt.Errorf("error opening downloaded file: %s", err)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading checksum file: %s", err)
			}
			if line == "" {
				break
			}
			// parse the line, if we hit EOF, but the line is not empty
		}
		checksum, filename, err := parseChecksumLine(line)
		if err != nil || checksum == "" {
			continue
		}
		if filename == h.FileName {
			h.value = checksum
			return nil
		}
	}
	return fmt.Errorf("no checksum found in: %s", h.url.String())
}

// Value returns the checksum value.
func (h *HTTPChecksum) Value() string {
	return h.value
}

// parseChecksumLine takes a line from a checksum file and returns the checksum and filename.
func parseChecksumLine(line string) (string, string, error) {
	parts := strings.Fields(line)

	switch len(parts) {
	case 4:
		// BSD-style checksum:
		//  MD5 (file1) = <checksum>
		//  MD5 (file2) = <checksum>
		if len(parts[1]) <= 2 ||
			parts[1][0] != '(' || parts[1][len(parts[1])-1] != ')' {
			return "", "", fmt.Errorf(
				"unexpected BSD-style-checksum filename format: %s", line)
		}
		filename := parts[1][1 : len(parts[1])-1]
		return parts[3], filename, nil
	case 2:
		// GNU-style:
		//  <checksum>  file1
		//  <checksum> *file2
		return parts[0], parts[1], nil
	case 0:
		return "", "", nil
	default:
		return parts[0], "", nil
	}
}

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

// Package osrelease is to parse a os release file content.
package osrelease

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/pkg/errors"
)

const (
	// EtcOsRelease is the path of os-release file.
	EtcOsRelease string = "/etc/os-release"
	// DebianID is the identifier used by the Debian operating system.
	DebianID = "debian"
	// FedoraID is the identifier used by the Fedora operating system.
	FedoraID = "fedora"
	// UbuntuID is the identifier used by the Ubuntu operating system.
	UbuntuID = "ubuntu"
	// RhelID is the identifier used by the Rhel operating system.
	RhelID = "rhel"
	// CentosID is the identifier used by the Centos operating system.
	CentosID = "centos"
)

// Data exposes the most common identification parameters.
type Data struct {
	ID         string
	IDLike     string
	Name       string
	PrettyName string
	Version    string
	VersionID  string
}

// Parse is to parse a os release file content.
func Parse(content string) (data *Data) {
	data = new(Data)
	lines, err := parseString(content)
	if err != nil {
		return
	}

	info := make(map[string]string)
	for _, v := range lines {
		key, value, err := parseLine(v)
		if err == nil {
			info[key] = value
		}
	}
	data.ID = info["ID"]
	data.IDLike = info["ID_LIKE"]
	data.Name = info["NAME"]
	data.PrettyName = info["PRETTY_NAME"]
	data.Version = info["VERSION"]
	data.VersionID = info["VERSION_ID"]
	return
}

func parseString(content string) (lines []string, err error) {
	in := bytes.NewBufferString(content)
	reader := bufio.NewReader(in)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func parseLine(line string) (string, string, error) {
	// skip empty lines
	if line == "" {
		return "", "", errors.New("Skipping: zero-length")
	}

	// skip comments
	if line[0] == '#' {
		return "", "", errors.New("Skipping: comment")
	}

	// try to split string at the first '='
	splitString := strings.SplitN(line, "=", 2)
	if len(splitString) != 2 {
		return "", "", errors.New("Can not extract key=value")
	}

	// trim white space from key and value
	key := splitString[0]
	key = strings.Trim(key, " ")
	value := splitString[1]
	value = strings.Trim(value, " ")

	// Handle double quotes
	if strings.ContainsAny(value, `"`) {
		first := value[0:1]
		last := value[len(value)-1:]

		if first == last && strings.ContainsAny(first, `"'`) {
			value = strings.TrimPrefix(value, `'`)
			value = strings.TrimPrefix(value, `"`)
			value = strings.TrimSuffix(value, `'`)
			value = strings.TrimSuffix(value, `"`)
		}
	}

	// expand anything else that could be escaped
	value = strings.Replace(value, `\"`, `"`, -1)
	value = strings.Replace(value, `\$`, `$`, -1)
	value = strings.Replace(value, `\\`, `\`, -1)
	value = strings.Replace(value, "\\`", "`", -1)
	return key, value, nil
}

// IsLikeDebian will return true for Debian and any other related OS, such as Ubuntu.
func (d *Data) IsLikeDebian() bool {
	return d.ID == DebianID || strings.Contains(d.IDLike, DebianID)
}

// IsLikeFedora will return true for Fedora and any other related OS, such as CentOS or RHEL.
func (d *Data) IsLikeFedora() bool {
	return d.ID == FedoraID || strings.Contains(d.IDLike, FedoraID)
}

// IsUbuntu will return true for Ubuntu OS.
func (d *Data) IsUbuntu() bool {
	return d.ID == UbuntuID
}

// IsRHEL will return true for RHEL OS.
func (d *Data) IsRHEL() bool {
	return d.ID == RhelID
}

// IsCentOS will return true for CentOS.
func (d *Data) IsCentOS() bool {
	return d.ID == CentosID
}

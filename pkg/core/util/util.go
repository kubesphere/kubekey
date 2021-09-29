/*
Copyright 2020 The KubeSphere Authors.

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
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"math"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"sort"
	"strings"
	"text/template"
)

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(tmpl *template.Template, variables map[string]interface{}) (string, error) {

	var buf strings.Builder

	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "Failed to render template")
	}
	return buf.String(), nil
}

// Home returns the home directory for the executing user.
func Home() (string, error) {
	u, err := user.Current()
	if nil == err {
		return u.HomeDir, nil
	}

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	return homeUnix()
}

func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func GetArgs(argsMap map[string]string, args []string) ([]string, map[string]string) {
	for _, arg := range args {
		splitArg := strings.SplitN(arg, "=", 2)
		if len(splitArg) < 2 {
			continue
		}
		argsMap[splitArg[0]] = splitArg[1]
	}

	for arg, value := range argsMap {
		cmd := fmt.Sprintf("%s=%s", arg, value)
		args = append(args, cmd)
	}
	sort.Strings(args)
	return args, argsMap
}

// Round returns the result of rounding 'val' according to the specified 'precision' precision (the number of digits after the decimal point)。
// and precision can be negative number or zero
func Round(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Floor(val*p+0.5) / p
}

// ArchAlias returns the alias of cpu's architecture.
// amd64: x86_64
// arm64: aarch64
func ArchAlias(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return ""
	}
}

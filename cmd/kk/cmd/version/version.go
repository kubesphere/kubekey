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

package version

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/kubesphere/kubekey/cmd/kk/pkg/version/kubernetes"
	"github.com/kubesphere/kubekey/version"
)

type Version struct {
	KK *version.Info `json:"kk"`
}

type VersionOptions struct {
	Output                      string
	ShowSupportedK8sVersionList bool
}

func NewVersionOptions() *VersionOptions {
	return &VersionOptions{}
}

// NewCmdVersion creates a new version command
func NewCmdVersion() *cobra.Command {
	o := NewVersionOptions()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "print the client version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if o.ShowSupportedK8sVersionList {
				return printSupportedK8sVersionList(cmd.OutOrStdout())
			}
			return o.Run()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func (o *VersionOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Output, "output", "o", "", "Output format; available options are 'yaml', 'json' and 'short'")
	cmd.Flags().BoolVarP(&o.ShowSupportedK8sVersionList, "show-supported-k8s", "", false,
		`print the version of supported k8s`)
}

func (o *VersionOptions) Run() error {
	clientVersion := version.Get()
	v := Version{
		KK: &clientVersion,
	}

	switch o.Output {
	case "":
		fmt.Printf("kk version: %#v\n", v.KK)
	case "short":
		fmt.Printf("%s\n", v.KK.GitVersion)
	case "yaml":
		y, err := yaml.Marshal(&v)
		if err != nil {
			return err
		}
		fmt.Print(string(y))
	case "json":
		y, err := json.MarshalIndent(&v, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(y))
	default:
		return errors.Errorf("invalid output format: %s", o.Output)
	}
	return nil
}

func printSupportedK8sVersionList(output io.Writer) (err error) {
	_, err = output.Write([]byte(fmt.Sprintln(strings.Join(kubernetes.SupportedK8sVersionList(), "\n"))))
	return
}

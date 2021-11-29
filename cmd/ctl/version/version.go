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
	"fmt"
	"github.com/kubesphere/kubekey/pkg/version/kubernetes"
	"github.com/kubesphere/kubekey/version"
	"github.com/spf13/cobra"
	"io"
	"strings"
)

type VersionOptions struct {
	ShortVersion                bool
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
			return printVersion(o.ShortVersion)
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func (o *VersionOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.ShortVersion, "short", "", false, "print the version number")
	cmd.Flags().BoolVarP(&o.ShowSupportedK8sVersionList, "show-supported-k8s", "", false,
		`print the version of supported k8s`)
}

func printVersion(short bool) error {
	v := version.Get()
	if short {
		if len(v.GitCommit) >= 7 {
			fmt.Printf("%s+g%s\n", v.Version, v.GitCommit[:7])
			return nil
		}
		fmt.Println(version.GetVersion())
	}
	fmt.Printf("%#v\n", v)
	return nil
}

func printSupportedK8sVersionList(output io.Writer) (err error) {
	_, err = output.Write([]byte(fmt.Sprintln(strings.Join(kubernetes.SupportedK8sVersionList(), "\n"))))
	return
}

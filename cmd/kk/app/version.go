/*
Copyright 2023 The KubeSphere Authors.

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

package app

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/version"
)

// VersionOptions holds the flags for the version command.
type VersionOptions struct {
	Short bool // Short determines if only the version number is printed.
}

// newVersionCommand creates the cobra command for printing KubeSphere's version information.
func newVersionCommand() *cobra.Command {
	o := &VersionOptions{}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of KubeSphere controller-manager",
		Run: func(cmd *cobra.Command, _ []string) {
			// Print the short or full version info based on the --short flag.
			if o.Short {
				cmd.Println(version.Get().GitVersion)
			} else {
				cmd.Println(version.Get())
			}
		},
	}

	cmd.Flags().BoolVarP(&o.Short, "short", "s", false, "Print just the version number")

	return cmd
}

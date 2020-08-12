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

package cmd

import (
	"fmt"
	"github.com/kubesphere/kubekey/version"
	"github.com/spf13/cobra"
)

var shortVersion bool

// versionCmd represents the scale command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the client version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return printVersion(shortVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&shortVersion, "short", "", false, "print the version number")
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

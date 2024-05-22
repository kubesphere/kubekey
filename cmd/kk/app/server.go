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
	"flag"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

var internalCommand = []*cobra.Command{}

func registerInternalCommand(command *cobra.Command) {
	for _, c := range internalCommand {
		if c.Name() == command.Name() {
			// command has register. skip
			return
		}
	}
	internalCommand = append(internalCommand, command)
}

func NewKubeKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "kk",
		Long: "kubekey is a daemon that execute command in a node",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := initGOPS(); err != nil {
				return err
			}
			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return flushProfiling()
		},
	}

	// todo add --set override the config.yaml data.

	flags := cmd.PersistentFlags()
	addProfilingFlags(flags)
	addKlogFlags(flags)
	addGOPSFlags(flags)

	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newVersionCommand())

	// internal command
	cmd.AddCommand(internalCommand...)
	return cmd
}

func addKlogFlags(fs *pflag.FlagSet) {
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})
}

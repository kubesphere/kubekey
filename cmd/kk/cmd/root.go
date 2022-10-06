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
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/add"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/artifact"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/cert"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/completion"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/create"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/cri"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/delete"
	initOs "github.com/kubesphere/kubekey/cmd/kk/cmd/init"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/plugin"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/upgrade"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/version"
)

type KubeKeyOptions struct {
	PluginHandler PluginHandler
	Arguments     []string

	options.IOStreams
}

func NewDefaultKubeKeyCommand() *cobra.Command {
	return NewDefaultKubeKeyCommandWithArgs(KubeKeyOptions{
		PluginHandler: NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
		Arguments:     os.Args,

		IOStreams: options.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	})
}

func NewDefaultKubeKeyCommandWithArgs(o KubeKeyOptions) *cobra.Command {
	cmd := NewKubeKeyCommand(o)

	if o.PluginHandler == nil {
		return cmd
	}

	if len(o.Arguments) > 1 {
		cmdPathPieces := o.Arguments[1:]

		// only look for suitable extension executables if
		// the specified command does not already exist
		if _, _, err := cmd.Find(cmdPathPieces); err != nil {
			// Also check the commands that will be added by Cobra.
			// These commands are only added once rootCmd.Execute() is called, so we
			// need to check them explicitly here.
			var cmdName string // first "non-flag" arguments
			for _, arg := range cmdPathPieces {
				if !strings.HasPrefix(arg, "-") {
					cmdName = arg
					break
				}
			}

			switch cmdName {
			case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
				// Don't search for a plugin
			default:
				if err := HandlePluginCommand(o.PluginHandler, cmdPathPieces); err != nil {
					os.Exit(1)
				}
			}
		}
	}

	return cmd
}

// NewKubeKeyCommand creates a new kubekey root command
func NewKubeKeyCommand(o KubeKeyOptions) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "kk",
		Short: "Kubernetes/KubeSphere Deploy Tool",
		Long: `Deploy a Kubernetes or KubeSphere cluster efficiently, flexibly and easily. There are three scenarios to use KubeKey.
1. Install Kubernetes only
2. Install Kubernetes and KubeSphere together in one command
3. Install Kubernetes first, then deploy KubeSphere on it using https://github.com/kubesphere/ks-installer`,
	}

	cmds.AddCommand(initOs.NewCmdInit())

	cmds.AddCommand(create.NewCmdCreate())
	cmds.AddCommand(delete.NewCmdDelete())
	cmds.AddCommand(add.NewCmdAdd())
	cmds.AddCommand(upgrade.NewCmdUpgrade())
	cmds.AddCommand(cert.NewCmdCerts())
	cmds.AddCommand(artifact.NewCmdArtifact())

	cmds.AddCommand(plugin.NewCmdPlugin(o.IOStreams))

	cmds.AddCommand(completion.NewCmdCompletion())
	cmds.AddCommand(version.NewCmdVersion())

	cmds.AddCommand(cri.NewCmdCri())
	return cmds
}

// PluginHandler is capable of parsing command line arguments
// and performing executable filename lookups to search
// for valid plugin files, and execute found plugins.
type PluginHandler interface {
	// exists at the given filename, or a boolean false.
	// Lookup will iterate over a list of given prefixes
	// in order to recognize valid plugin filenames.
	// The first filepath to match a prefix is returned.
	Lookup(filename string) (string, bool)
	// Execute receives an executable's filepath, a slice
	// of arguments, and a slice of environment variables
	// to relay to the executable.
	Execute(executablePath string, cmdArgs, environment []string) error
}

// DefaultPluginHandler implements PluginHandler
type DefaultPluginHandler struct {
	ValidPrefixes []string
}

// NewDefaultPluginHandler instantiates the DefaultPluginHandler with a list of
// given filename prefixes used to identify valid plugin filenames.
func NewDefaultPluginHandler(validPrefixes []string) *DefaultPluginHandler {
	return &DefaultPluginHandler{
		ValidPrefixes: validPrefixes,
	}
}

// Lookup implements PluginHandler
func (h *DefaultPluginHandler) Lookup(filename string) (string, bool) {
	for _, prefix := range h.ValidPrefixes {
		path, err := exec.LookPath(fmt.Sprintf("%s-%s", prefix, filename))
		if err != nil || len(path) == 0 {
			continue
		}
		return path, true
	}

	return "", false
}

// Execute implements PluginHandler
func (h *DefaultPluginHandler) Execute(executablePath string, cmdArgs, environment []string) error {

	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := exec.Command(executablePath, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = environment
		err := cmd.Run()
		if err == nil {
			os.Exit(0)
		}
		return err
	}

	// invoke cmd binary relaying the environment and args given
	// append executablePath to cmdArgs, as execve will make first argument the "binary name".
	return syscall.Exec(executablePath, append([]string{executablePath}, cmdArgs...), environment)
}

// HandlePluginCommand receives a pluginHandler and command-line arguments and attempts to find
// a plugin executable on the PATH that satisfies the given arguments.
func HandlePluginCommand(pluginHandler PluginHandler, cmdArgs []string) error {
	var remainingArgs []string // all "non-flag" arguments
	for _, arg := range cmdArgs {
		if strings.HasPrefix(arg, "-") {
			break
		}
		remainingArgs = append(remainingArgs, strings.Replace(arg, "-", "_", -1))
	}

	if len(remainingArgs) == 0 {
		// the length of cmdArgs is at least 1
		return fmt.Errorf("flags cannot be placed before plugin name: %s", cmdArgs[0])
	}

	foundBinaryPath := ""

	// attempt to find binary, starting at longest possible name with given cmdArgs
	for len(remainingArgs) > 0 {
		path, found := pluginHandler.Lookup(strings.Join(remainingArgs, "-"))
		if !found {
			remainingArgs = remainingArgs[:len(remainingArgs)-1]
			continue
		}

		foundBinaryPath = path
		break
	}

	if len(foundBinaryPath) == 0 {
		return nil
	}

	// invoke cmd binary relaying the current environment and args given
	if err := pluginHandler.Execute(foundBinaryPath, cmdArgs[len(remainingArgs):], os.Environ()); err != nil {
		return err
	}

	return nil
}

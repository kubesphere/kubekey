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

package completion

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
)

// CompletionOptions is the option of completion command
type CompletionOptions struct {
	Type string
}

func NewCompletionOptions() *CompletionOptions {
	return &CompletionOptions{}
}

// ShellTypes contains all types of shell
var ShellTypes = []string{
	"zsh", "bash", "powerShell",
}

var completionOptions CompletionOptions

func NewCmdCompletion() *cobra.Command {
	o := NewCompletionOptions()
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts
Normally you don't need to do more extra work to have this feature if you've installed kk by brew`,
		Example: `# Installing bash completion on Linux
## If bash-completion is not installed on Linux, please install the 'bash-completion' package
## via your distribution's package manager.
## Load the ks completion code for bash into the current shell
source <(ks completion bash)
## Write bash completion code to a file and source if from .bash_profile
mkdir -p ~/.config/kk/ && kk completion --type bash > ~/.config/kk/completion.bash.inc
printf "
# kk shell completion
source '$HOME/.config/kk/completion.bash.inc'
" >> $HOME/.bash_profile
source $HOME/.bash_profile

In order to have good experience on zsh completion, ohmyzsh is a good choice.
Please install ohmyzsh by the following command
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
Get more details about onmyzsh from https://github.com/ohmyzsh/ohmyzsh

Load the kk completion code for zsh[1] into the current shell
source <(kk completion --type zsh)
Set the kk completion code for zsh[1] to autoload on startup
kk completion --type zsh > "${fpath[1]}/_kk"`,
		Run: func(cmd *cobra.Command, _ []string) {
			util.CheckErr(o.Run(cmd))
		},
	}

	o.AddFlags(cmd)
	if err := completionSetting(cmd); err != nil {
		panic(fmt.Sprintf("register flag type for sub-command doc failed %#v\n", err))
	}
	return cmd
}

func (o *CompletionOptions) Run(cmd *cobra.Command) error {
	var err error
	shellType := completionOptions.Type
	switch shellType {
	case "zsh":
		err = cmd.GenZshCompletion(cmd.OutOrStdout())
	case "powerShell":
		err = cmd.GenPowerShellCompletion(cmd.OutOrStdout())
	case "bash":
		err = cmd.GenBashCompletion(cmd.OutOrStdout())
	case "":
		err = cmd.Help()
	default:
		err = fmt.Errorf("unknown shell type %s", shellType)
	}
	return err
}

func (o *CompletionOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&completionOptions.Type, "type", "t", "",
		fmt.Sprintf("Generate different types of shell which are %v", ShellTypes))
}

func completionSetting(cmd *cobra.Command) error {
	err := cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) (
		i []string, directive cobra.ShellCompDirective) {
		return ShellTypes, cobra.ShellCompDirectiveDefault
	})
	return err
}

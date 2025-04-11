//go:build builtin
// +build builtin

package app

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/builtin"
)

// registerInternalCommand registers an internal command to the list of internal commands.
// It ensures that the command is not already registered before adding it to the list.
//
// Parameters:
//   - command: The command to be registered.
func registerInternalCommand(command *cobra.Command) {
	for _, c := range internalCommand {
		if c.Name() == command.Name() {
			// command has been registered. skip
			return
		}
	}
	internalCommand = append(internalCommand, command)
}

func init() {
	registerInternalCommand(builtin.NewArtifactCommand())
	registerInternalCommand(builtin.NewCertsCommand())
	registerInternalCommand(builtin.NewCreateCommand())
	registerInternalCommand(builtin.NewDeleteCommand())
	registerInternalCommand(builtin.NewInitCommand())
	registerInternalCommand(builtin.NewPreCheckCommand())
}

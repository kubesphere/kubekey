package modules

import (
	"context"
	"fmt"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
Module Setup

This module is used to set up the connection to a remote host and gather facts about it.
It performs the following operations:
1. Establishes a connection to the specified host using the appropriate connector
2. If the connector supports fact gathering (implements GatherFacts interface):
   - Retrieves host information
   - Merges the remote host information into the variables map
3. Returns success if all operations complete successfully

Usage:
  - host: The target host to connect to
  - variable: Map of variables to be used for connection and fact gathering
*/

// ModuleSetup establishes a connection to a remote host and gathers facts about it.
// It returns StdoutSuccess if successful, or an error message if any step fails.
func ModuleSetup(ctx context.Context, options ExecOptions) (string, string) {
	// get connector
	conn, err := options.getConnector(ctx)
	if err != nil {
		return StdoutFailed, fmt.Sprintf("failed to connector of %q error: %v", options.Host, err)
	}
	defer conn.Close(ctx)

	if gf, ok := conn.(connector.GatherFacts); ok {
		remoteInfo, err := gf.HostInfo(ctx)
		if err != nil {
			return StdoutFailed, err.Error()
		}
		if err := options.Variable.Merge(variable.MergeRemoteVariable(remoteInfo, options.Host)); err != nil {
			return StdoutFailed, err.Error()
		}
	}

	return StdoutSuccess, ""
}

func init() {
	utilruntime.Must(RegisterModule("setup", ModuleSetup))
}

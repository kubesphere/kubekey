package setup

import (
	"context"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
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
func ModuleSetup(ctx context.Context, opts internal.ExecOptions) (string, string, error) {
	// get connector
	conn, err := opts.GetConnector(ctx)
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetConnector, nil
	}
	defer conn.Close(ctx)

	if gf, ok := conn.(connector.GatherFacts); ok {
		remoteInfo, err := gf.HostInfo(ctx)
		if err != nil {
			return internal.StdoutFailed, "failed to get host info", err
		}
		if err := opts.Merge(variable.MergeRemoteVariable(remoteInfo, opts.Host)); err != nil {
			return internal.StdoutFailed, "failed to merge setup variable", err
		}
	}

	return internal.StdoutSuccess, "", nil
}

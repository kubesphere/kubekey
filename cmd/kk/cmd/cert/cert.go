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

package cert

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
)

type CertOptions struct {
	CommonOptions *options.CommonOptions
}

func NewCertsOptions() *CertOptions {
	return &CertOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCerts creates a new cert command
func NewCmdCerts() *cobra.Command {
	o := NewCertsOptions()
	cmd := &cobra.Command{
		Use:   "certs",
		Short: "cluster certs",
	}

	o.CommonOptions.AddCommonFlag(cmd)

	cmd.AddCommand(NewCmdCertList())
	cmd.AddCommand(NewCmdCertRenew())
	return cmd
}

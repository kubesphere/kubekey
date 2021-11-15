/*
 Copyright 2021 The KubeSphere Authors.

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

package options

import (
	"github.com/spf13/cobra"
)

type CommonOptions struct {
	InCluster        bool
	Verbose          bool
	SkipConfirmCheck bool
	IgnoreErr        bool
}

func NewCommonOptions() *CommonOptions {
	return &CommonOptions{}
}

func (o *CommonOptions) AddCommonFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.InCluster, "in-cluster", false, "Running inside the cluster")
	cmd.Flags().BoolVar(&o.Verbose, "debug", true, "Print detailed information")
	cmd.Flags().BoolVarP(&o.SkipConfirmCheck, "yes", "y", false, "Skip confirm check")
	cmd.Flags().BoolVar(&o.IgnoreErr, "ignore-err", false, "Ignore the error message, remove the host which reported error and force to continue")
}

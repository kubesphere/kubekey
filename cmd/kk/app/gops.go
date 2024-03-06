/*
Copyright 2024 The KubeSphere Authors.

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
	"github.com/google/gops/agent"
	"github.com/spf13/pflag"
)

var gops bool

func addGOPSFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&gops, "gops", false, "Whether to enable gops or not.  When enabled this option, "+
		"controller-manager will listen on a random port on 127.0.0.1, then you can use the gops tool to list and diagnose the controller-manager currently running.")
}

func initGOPS() error {
	if gops {
		// Add agent to report additional information such as the current stack trace, Go version, memory stats, etc.
		// Bind to a random port on address 127.0.0.1
		if err := agent.Listen(agent.Options{}); err != nil {
			return err
		}
	}
	return nil
}

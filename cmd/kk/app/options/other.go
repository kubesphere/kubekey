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

package options

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/google/gops/agent"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

// ======================================================================================
//                                     PROFILING
// ======================================================================================

var (
	profileName   string
	profileOutput string
)

// AddProfilingFlags to NewRootCommand
func AddProfilingFlags(flags *pflag.FlagSet) {
	flags.StringVar(&profileName, "profile", "none", "Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex)")
	flags.StringVar(&profileOutput, "profile-output", "profile.pprof", "Name of the file to write the profile to")
}

// InitProfiling for profileName
func InitProfiling(ctx context.Context) error {
	var (
		f   *os.File
		err error
	)

	switch profileName {
	case "none":
		return nil
	case "cpu":
		f, err = os.Create(profileOutput)
		if err != nil {
			return errors.Wrap(err, "failed to create cpu profile")
		}

		err = pprof.StartCPUProfile(f)
		if err != nil {
			return errors.Wrap(err, "failed to start cpu profile")
		}
	// Block and mutex profiles need a call to Set{Block,Mutex}ProfileRate to
	// output anything. We choose to sample all events.
	case "block":
		runtime.SetBlockProfileRate(1)
	case "mutex":
		runtime.SetMutexProfileFraction(1)
	default:
		// Check the profile name is valid.
		if profile := pprof.Lookup(profileName); profile == nil {
			return errors.Errorf("unknown profile '%s'", profileName)
		}
	}

	// If the command is interrupted before the end (ctrl-c), flush the
	// profiling files

	go func() {
		<-ctx.Done()
		if err := f.Close(); err != nil {
			fmt.Printf("failed to close file. file: %v. error: %v \n", profileOutput, err)
		}

		if err := FlushProfiling(); err != nil {
			fmt.Printf("failed to FlushProfiling. file: %v. error: %v \n", profileOutput, err)
		}
	}()

	return nil
}

// FlushProfiling to local file
func FlushProfiling() error {
	switch profileName {
	case "none":
		return nil
	case "cpu":
		pprof.StopCPUProfile()
	case "heap":
		runtime.GC()

		fallthrough
	default:
		profile := pprof.Lookup(profileName)
		if profile == nil {
			return nil
		}

		f, err := os.Create(profileOutput)
		if err != nil {
			return errors.Wrap(err, "failed to create profile")
		}
		defer f.Close()

		if err := profile.WriteTo(f, 0); err != nil {
			return errors.Wrap(err, "failed to write profile")
		}
	}

	return nil
}

// ======================================================================================
//                                         GOPS
// ======================================================================================

var gops bool

// AddGOPSFlags to NewRootCommand
func AddGOPSFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&gops, "gops", false, "Whether to enable gops or not.  When enabled this option, "+
		"controller-manager will listen on a random port on 127.0.0.1, then you can use the gops tool to list and diagnose the controller-manager currently running.")
}

// InitGOPS if gops is true
func InitGOPS() error {
	if gops {
		// Add agent to report additional information such as the current stack trace, Go version, memory stats, etc.
		// Bind to a random port on address 127.0.0.1
		if err := agent.Listen(agent.Options{}); err != nil {
			return errors.Wrap(err, "failed to listen gops")
		}
	}

	return nil
}

// ======================================================================================
//                                       KLOG
// ======================================================================================

// AddKlogFlags to NewRootCommand
func AddKlogFlags(fs *pflag.FlagSet) {
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})
}

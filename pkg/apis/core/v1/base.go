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

package v1

type Base struct {
	Name string `yaml:"name,omitempty"`

	// connection/transport
	Connection string `yaml:"connection,omitempty"`
	Port       int    `yaml:"port,omitempty"`
	RemoteUser string `yaml:"remote_user,omitempty"`

	// variables
	Vars map[string]any `yaml:"vars,omitempty"`

	// module default params
	//ModuleDefaults []map[string]map[string]any `yaml:"module_defaults,omitempty"`

	// flags and misc. settings
	Environment    []map[string]string `yaml:"environment,omitempty"`
	NoLog          bool                `yaml:"no_log,omitempty"`
	RunOnce        bool                `yaml:"run_once,omitempty"`
	IgnoreErrors   *bool               `yaml:"ignore_errors,omitempty"`
	CheckMode      bool                `yaml:"check_mode,omitempty"`
	Diff           bool                `yaml:"diff,omitempty"`
	AnyErrorsFatal bool                `yaml:"any_errors_fatal,omitempty"`
	Throttle       int                 `yaml:"throttle,omitempty"`
	Timeout        int                 `yaml:"timeout,omitempty"`

	// Debugger invoke a debugger on tasks
	Debugger string `yaml:"debugger,omitempty"`

	// privilege escalation
	Become       bool   `yaml:"become,omitempty"`
	BecomeMethod string `yaml:"become_method,omitempty"`
	BecomeUser   string `yaml:"become_user,omitempty"`
	BecomeFlags  string `yaml:"become_flags,omitempty"`
	BecomeExe    string `yaml:"become_exe,omitempty"`
}

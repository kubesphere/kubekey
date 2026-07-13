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

package modules

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/modules/add_hostvars"
	"github.com/kubesphere/kubekey/v4/pkg/modules/assert"
	"github.com/kubesphere/kubekey/v4/pkg/modules/command"
	"github.com/kubesphere/kubekey/v4/pkg/modules/copy"
	"github.com/kubesphere/kubekey/v4/pkg/modules/debug"
	"github.com/kubesphere/kubekey/v4/pkg/modules/fetch"
	"github.com/kubesphere/kubekey/v4/pkg/modules/gen_cert"
	"github.com/kubesphere/kubekey/v4/pkg/modules/http_get_file"
	"github.com/kubesphere/kubekey/v4/pkg/modules/image"
	"github.com/kubesphere/kubekey/v4/pkg/modules/include_vars"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/modules/prometheus"
	"github.com/kubesphere/kubekey/v4/pkg/modules/result"
	"github.com/kubesphere/kubekey/v4/pkg/modules/set_fact"
	"github.com/kubesphere/kubekey/v4/pkg/modules/setup"
	"github.com/kubesphere/kubekey/v4/pkg/modules/template"
)

// Re-export types and constants from options package
type (
	ExecOptions = internal.ExecOptions
)

var (
	StdoutSuccess = internal.StdoutSuccess
	StdoutFailed  = internal.StdoutFailed
	StdoutSkip    = internal.StdoutSkip
)

// FindModule retrieves a registered module execution function by its name.
var FindModule = internal.FindModule

func init() {
	// Register all built-in modules
	utilruntime.Must(internal.RegisterModule(add_hostvars.ModuleAddHostvars, "add_hostvars"))
	utilruntime.Must(internal.RegisterModule(assert.ModuleAssert, "assert"))
	utilruntime.Must(internal.RegisterModule(command.ModuleCommand, "command", "shell"))
	utilruntime.Must(internal.RegisterModule(copy.ModuleCopy, "copy"))
	utilruntime.Must(internal.RegisterModule(debug.ModuleDebug, "debug"))
	utilruntime.Must(internal.RegisterModule(fetch.ModuleFetch, "fetch"))
	utilruntime.Must(internal.RegisterModule(gen_cert.ModuleGenCert, "gen_cert"))
	utilruntime.Must(internal.RegisterModule(http_get_file.ModuleHttpGetFile, "http_get_file"))
	utilruntime.Must(internal.RegisterModule(image.ModuleImage, "image"))
	utilruntime.Must(internal.RegisterModule(include_vars.ModuleIncludeVars, "include_vars"))
	utilruntime.Must(internal.RegisterModule(prometheus.ModulePrometheus, "prometheus"))
	utilruntime.Must(internal.RegisterModule(result.ModuleResult, "result"))
	utilruntime.Must(internal.RegisterModule(set_fact.ModuleSetFact, "set_fact"))
	utilruntime.Must(internal.RegisterModule(setup.ModuleSetup, "setup"))
	utilruntime.Must(internal.RegisterModule(template.ModuleTemplate, "template"))
}

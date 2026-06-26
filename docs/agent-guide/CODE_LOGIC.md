# KubeKey Code Logic Map

Companion to [AGENTS.md](../../AGENTS.md). This doc traces the exact code paths for the most important flows so you can jump straight to the right function when debugging or adding features.

---

## 1. CLI Startup Flow

### `kk` binary entry

```text
cmd/kk/kubekey.go:main()
    └── app.NewRootCommand().Execute()
```

### Root command construction

```text
cmd/kk/app/root.go:NewRootCommand()
    ├── options.AddProfilingFlags()      # pprof/gops
    ├── options.AddKlogFlags()
    ├── options.AddGOPSFlags()
    ├── newRunCommand()
    ├── newPlaybookCommand()
    ├── newVersionCommand()
    ├── newWebCommand()
    └── internalCommand...               # built-in commands registered by init()
```

### Built-in commands registration (gated by `//go:build builtin`)

```text
cmd/kk/app/builtin.go:init()
    └── imports cmd/kk/app/builtin/* packages

cmd/kk/app/builtin/create.go:init()
    └── internalCommand = append(internalCommand, newCreateCommand())

cmd/kk/app/builtin/add.go:init()
cmd/kk/app/builtin/delete.go:init()
cmd/kk/app/builtin/init.go:init()
cmd/kk/app/builtin/precheck.go:init()
cmd/kk/app/builtin/artifact.go:init()
cmd/kk/app/builtin/certs.go:init()
```

Each built-in command constructs a `CommonOptions` and calls `CommonOptions.Run()`.

### Built-in command flow (example: create cluster)

```text
cmd/kk/app/builtin/create.go:newCreateCommand()
    └── cmd.RunE = func(...)
        ├── options.NewCommonOptions()
        │   └── sets up Playbook/Inventory/Config references
        ├── options.Complete()
        │   ├── resolve inventory/config files
        │   ├── apply --set overrides
        │   └── determine workdir
        └── options.Run()
            └── manager.NewCommandManager(playbook, inventory, config)
                └── commandManager.Run()
                    └── executor.NewPlaybookExecutor(...).Exec(ctx)
```

### Arbitrary playbook flow (`kk run`)

```text
cmd/kk/app/run.go:newRunCommand()
    └── options.KubeKeyRunOptions.Complete()
        ├── project.New() for git/local project
        └── build Playbook CR pointing at that project
    └── options.Run() -> CommandManager -> PlaybookExecutor
```

### In-cluster executor (`kk playbook`)

```text
cmd/kk/app/playbook.go:newPlaybookCommand()
    └── PlaybookOptions.Complete()
        ├── read Playbook CR from API server
        └── read Inventory/Config CRs
    └── Run() -> CommandManager -> PlaybookExecutor
```

---

## 2. Manager Layer

All three binaries converge on the `Manager` interface in `pkg/manager/manager.go:31`.

### Command manager

```text
pkg/manager/command_manager.go:NewCommandManager()
    └── Run(ctx)
        ├── create controller-runtime client for local file storage
        ├── if local run and not dry-run: create/update Playbook CR locally
        └── PlaybookExecutor.Exec(ctx)
```

### Controller manager

```text
cmd/controller-manager/controller_manager.go:main()
    └── app.NewControllerManagerCommand().Execute()
        └── pkg/manager/controller_manager.go:NewControllerManager().Run(ctx)
            ├── create controller-runtime manager
            ├── register enabled controllers via options.Register()
            └── mgr.Start(ctx)
```

Controllers register in `pkg/controllers/core/register.go:init()` and `pkg/controllers/infrastructure/register.go:init()`.

### Web manager

```text
pkg/manager/web_manager.go:NewWebManager().Run(ctx)
    ├── create local REST config via pkg/proxy
    ├── build go-restful container
    ├── pkg/web.NewCoreService()
    ├── pkg/web.NewSchemaService()
    ├── pkg/web.NewUIService()
    └── http.ListenAndServe()
```

---

## 3. Project Loading

### Project factory

```text
pkg/project/project.go:New(ctx, playbook, update)
    ├── if playbook address looks like git: newGitProject()
    │   └── go-git clone/pull into workdir
    ├── else if BuiltinsProjectAnnotation is set: builtinProjectFunc()
    │   └── builtin/core.BuiltinPlaybook embed.FS
    └── else: newLocalProject()
        └── os.DirFS(path)
```

### Playbook marshaling

```text
pkg/project/project.go:MarshalPlaybook()
    ├── ReadFile(playbook.yaml)
    ├── yaml.Unmarshal -> kkprojectv1.Playbook
    ├── resolve import_playbook recursively
    ├── load vars_files
    ├── load roles:
    │   ├── read defaults/main.yaml
    │   ├── read meta/main.yaml dependencies
    │   └── recursively marshal dependency roles
    ├── expand include_tasks
    └── validate playbook/role/block
```

YAML project types live in `api/project/v1/`:

- `playbook.go:Playbook` – top-level list of Plays.
- `play.go:Play` – hosts, gather_facts, vars_files, roles, pre_tasks/tasks/post_tasks.
- `role.go:Role` / `RoleInfo` – dependencies, name, blocks.
- `block.go:Block` – nested block/rescue/always or leaf task.
- `base.go:Base` – name, connection, vars, environment, run_once, ignore_errors, become.
- `taggable.go:Taggable` – tags logic (always/never/all/tagged).
- `conditional.go:When` – conditional evaluation.

---

## 4. Execution Engine

### Executor creation

```text
pkg/executor/playbook_executor.go:NewPlaybookExecutor(client, playbook, variable, logOutput)
    └── returns *playbookExecutor{ option{...}, project }
```

### Playbook execution

```text
pkg/executor/playbook_executor.go:Exec(ctx)
    ├── project.MarshalPlaybook() -> kkprojectv1.Playbook
    ├── set Playbook phase Running
    ├── for each Play:
    │   ├── select hosts from inventory by pattern
    │   ├── group hosts by serial batches
    │   │   └── pkg/converter/converter.go:GroupHostBySerial()
    │   ├── for each batch:
    │   │   ├── gather_facts (if play.gather_facts != false)
    │   │   │   └── setup module on each host
    │   │   ├── run pre_tasks
    │   │   ├── for each role:
    │   │   │   └── roleExecutor.Exec(ctx)
    │   │   ├── run tasks
    │   │   └── run post_tasks
    │   └── update Playbook status
    ├── set Playbook phase Succeeded/Failed
    └── store final result
```

### Role execution

```text
pkg/executor/role_executor.go:Exec(ctx)
    ├── merge role defaults into variable system
    ├── recursively execute dependency roles
    └── for each block in role:
        └── blockExecutor.Exec(ctx)
```

### Block execution

```text
pkg/executor/block_executor.go:Exec(ctx)
    ├── evaluate tags: skip block if tags don't match
    ├── evaluate when condition
    ├── if block has nested block/rescue/always:
    │   ├── run block tasks
    │   ├── on failure: run rescue tasks
    │   └── always: run always tasks
    └── else (leaf task):
        └── taskExecutor.Exec(ctx)
```

### Task execution

```text
pkg/executor/task_executor.go:Exec(ctx)
    ├── if loop: expand loop items
    ├── create/update kkcorev1alpha1.Task CR
    ├── for each host in parallel (wait.Group):
    │   ├── evaluate per-host when condition
    │   ├── create progress bar
    │   ├── FindModule(moduleName)
    │   ├── moduleExecFunc(ctx, ExecOptions)
    │   ├── evaluate failed_when
    │   ├── handle ignore_errors
    │   └── store register/result variables
    └── update Task CR status
```

### Module discovery

```text
pkg/executor/block_executor.go:MarshalBlock()
    ├── iterate over UnknownField(s) in kkprojectv1.Block
    └── first unknown key that matches a registered module -> ModuleName
```

---

## 5. Variables

### Variable structure

```text
pkg/variable/variable.go:value
    ├── Config    kkcorev1.Config
    ├── Inventory kkcorev1.Inventory
    ├── Hosts     map[string]host
    │   ├── RemoteVars  map[string]any   # gather_facts
    │   └── RuntimeVars map[string]any   # set_fact, register
    └── Result    map[string]any
```

### Lookup precedence

```text
pkg/variable/variable_get.go:GetFunc
    └── resolves in order:
        1. Config vars
        2. Host-specific inventory vars
        3. Group vars (for groups containing host)
        4. Inventory vars
        5. Runtime vars
        6. Remote vars
```

### Merge paths

```text
pkg/variable/variable_merge.go
    ├── MergeRemoteVariable()      # gather_facts -> Hosts[host].RemoteVars
    ├── MergeRuntimeVariable()     # set_fact/register -> Hosts[host].RuntimeVars
    ├── MergeHostsRuntimeVariable()# cross-host variable injection
    └── MergeResultVariable()      # task result -> Playbook.Status.Result
```

### Persistence

```text
pkg/variable/source/file_source.go
    └── reads/writes per-host vars to
        <workdir>/runtime/<namespace>/<playbook>/variable/<hostname>.yaml
```

---

## 6. Connectors

### Factory

```text
pkg/connector/connector.go:NewConnector(host, vars, logger)
    ├── connector.type == "local"     -> localConnector
    ├── connector.type == "ssh"       -> sshConnector
    ├── connector.type == "kubernetes"-> kubernetesConnector
    ├── connector.type == "prometheus"-> prometheusConnector
    └── default:
        ├── localhost -> localConnector
        └── otherwise -> sshConnector
```

### SSH connector

```text
pkg/connector/ssh_connector.go
    ├── Init() parses auth (password/key)
    ├── ExecuteCommand() via golang.org/x/crypto/ssh
    ├── PutFile() via sftp
    └── FetchFile() via sftp
```

### Local connector

```text
pkg/connector/local_connector.go
    ├── ExecuteCommand() via os/exec
    ├── PutFile() writes local file
    └── FetchFile() reads local file
```

### Fact gathering

```text
pkg/connector/gather_facts.go
    └── local/ssh connectors implement HostInfo()
        └── collects OS, arch, hostname, IP, memory, CPU facts
```

---

## 7. Modules

### Registry

```text
pkg/modules/internal/options.go
    ├── RegisterModule(fn, names...)
    ├── FindModule(name)
    └── ModuleExecFunc signature
```

### Module list (pkg/modules/module.go)

| Module | Package | Key file |
|--------|---------|----------|
| add_hostvars | `pkg/modules/add_hostvars` | `add_hostvars.go` |
| assert | `pkg/modules/assert` | `assert.go` |
| command/shell | `pkg/modules/command` | `command.go` |
| copy | `pkg/modules/copy` | `copy.go` |
| debug | `pkg/modules/debug` | `debug.go` |
| fetch | `pkg/modules/fetch` | `fetch.go` |
| gen_cert | `pkg/modules/gen_cert` | `gen_cert.go` |
| http_get_file | `pkg/modules/http_get_file` | `http_get_file.go` |
| image | `pkg/modules/image` | `image.go`, `image_deprecated.go`, `repository.go` |
| include_vars | `pkg/modules/include_vars` | `include_vars.go` |
| prometheus | `pkg/modules/prometheus` | `prometheus.go` |
| result | `pkg/modules/result` | `result.go` |
| set_fact | `pkg/modules/set_fact` | `set_fact.go` |
| setup | `pkg/modules/setup` | `setup.go` |
| template | `pkg/modules/template` | `template.go` |

### Example: command module

```text
pkg/modules/command/command.go:ModuleCommand(ctx, opts)
    ├── render args through template
    ├── build command string
    ├── opts.Connector.ExecuteCommand(cmd)
    └── return stdout/stderr
```

### Example: copy module

```text
pkg/modules/copy/copy.go:ModuleCopy(ctx, opts)
    ├── resolve src/content
    ├── optionally template content
    ├── opts.Connector.PutFile(data, dst, mode)
    └── return result
```

### Example: template module

```text
pkg/modules/template/template.go:ModuleTemplate(ctx, opts)
    ├── read src template
    ├── render with variables
    ├── opts.Connector.PutFile(rendered, dst, mode)
    └── return result
```

---

## 8. Templates

### Rendering

```text
pkg/converter/tmpl/template.go:ParseFunc()
    ├── if string contains "{{" and "}}":
    │   └── text/template.Execute()
    └── else return original string
```

### Functions

```text
pkg/converter/tmpl/functions.go
    ├── toYaml / fromYaml / toToml
    ├── ipInCIDR / ipFamily / isIP
    ├── pow / subtractList
    ├── fileExists / unquote / getStringSlice
    ├── toLowerByteUnit
    └── mapToNamedStringArgs
```

### `when` conditions

```text
api/project/v1/conditional.go
    └── when expressions are always wrapped as templates
        and rendered to a boolean-like result
```

---

## 9. Kubernetes Controllers

### Playbook controller

```text
pkg/controllers/core/playbook_controller.go:Reconcile()
    ├── fetch Playbook CR
    ├── if no executor Pod exists:
    │   └── create Pod running "kk playbook --name <name> --namespace <ns>"
    ├── watch owned Pods
    └── sync Playbook status from Pod phase/logs
```

### CAPKK controllers

```text
pkg/controllers/infrastructure/
    ├── inventory_controller.go
    ├── kkcluster_controller.go
    └── kkmachine_controller.go
```

### Registration

```text
pkg/controllers/core/register.go:init()
    └── options.Register(&PlaybookReconciler{})
    └── options.Register(&PlaybookWebhook{})

pkg/controllers/infrastructure/register.go:init()
    └── registers Inventory/KKCluster/KKMachine reconcilers
```

---

## 10. REST Proxy / Web

### Hybrid REST config

```text
pkg/proxy/transport.go:RestConfig()
    ├── if no k8s cluster: use file-based storage for Task/Inventory/Playbook
    └── if cluster exists: forward non-local resources to API server,
        keep Task local
```

This lets `kk` run without a Kubernetes cluster while still using controller-runtime clients.

### Web services

```text
pkg/web/service.go
    ├── NewCoreService()       # /api/v1/playbooks, /inventories, /logs
    ├── NewSchemaService()     # schema listing and config
    ├── NewUIService()         # static SPA UI
    ├── NewSwaggerUIService()  # swagger UI
    └── NewAPIService()        # OpenAPI JSON
```

---

## 11. Built-in Kubernetes Install Flow

High-level `create_cluster.yaml` flow (`builtin/core/playbooks/create_cluster.yaml`):

```text
1. native/root role on all hosts
2. hook/pre_install.yaml
3. load defaults + precheck on all hosts
4. on localhost:
       generate certs, download binaries/images
5. on etcd/k8s_cluster/image_registry/nfs:
       run native role
6. on etcd hosts (when external):
       etcd prepare/install
7. on image_registry hosts:
       docker + registry
8. on localhost (when registry configured):
       push images
9. on k8s_cluster hosts:
       CRI install
       kubernetes pre/init/join
       certs renewal
       custom labels/taints
10. on a random control plane host:
       CNI + storage class
11. hook/post_install.yaml
```

Default variables are loaded from `builtin/core/defaults/` and merged before playbook execution.

---

## 12. Common Debugging Tips

- **Find where a built-in command is defined:** search `cmd/kk/app/builtin/*.go` for the command name.
- **Find where a module is implemented:** search `pkg/modules/<name>/<name>.go` and confirm registration in `pkg/modules/module.go`.
- **Trace variable values:** start at `pkg/variable/variable_get.go` and add logging in `GetFunc`.
- **Trace task execution:** add logging in `pkg/executor/task_executor.go` before `FindModule`.
- **Test a module locally:** look at existing `*_test.go` files; many use the fake connector from `pkg/modules/internal/test.go`.
- **Regenerate CRDs:** `make generate-manifests-kubekey`.
- **Build with built-ins:** `make kk` (sets `BUILDTAGS=builtin`).

---

## 13. File-to-Concern Map

| Concern | File |
|---------|------|
| CLI root | `cmd/kk/app/root.go` |
| CLI options base | `cmd/kk/app/options/option.go` |
| Built-in commands | `cmd/kk/app/builtin/*.go` |
| Controller options | `cmd/controller-manager/app/options/controller_manager.go` |
| Playbook execution | `pkg/executor/playbook_executor.go` |
| Role execution | `pkg/executor/role_executor.go` |
| Block execution | `pkg/executor/block_executor.go` |
| Task execution | `pkg/executor/task_executor.go` |
| Project loading | `pkg/project/project.go` |
| Git project | `pkg/project/git.go` |
| Local project | `pkg/project/local.go` |
| Builtin project | `pkg/project/builtin.go` |
| Variable core | `pkg/variable/variable.go` |
| Variable get | `pkg/variable/variable_get.go` |
| Variable merge | `pkg/variable/variable_merge.go` |
| Variable source | `pkg/variable/source/file_source.go` |
| Connector factory | `pkg/connector/connector.go` |
| SSH connector | `pkg/connector/ssh_connector.go` |
| Local connector | `pkg/connector/local_connector.go` |
| Kubernetes connector | `pkg/connector/kubernetes_connector.go` |
| Prometheus connector | `pkg/connector/prometheus_connector.go` |
| Module registry | `pkg/modules/module.go` |
| Module internals | `pkg/modules/internal/options.go` |
| Template functions | `pkg/converter/tmpl/functions.go` |
| Template rendering | `pkg/converter/tmpl/template.go` |
| Block↔Task converter | `pkg/converter/converter.go` |
| Playbook CRD | `api/core/v1/playbook_types.go` |
| Inventory CRD | `api/core/v1/inventory_types.go` |
| Config CRD | `api/core/v1/config_types.go` |
| Task CRD | `api/core/v1alpha1/task_types.go` |
| Project YAML types | `api/project/v1/playbook.go`, `play.go`, `block.go`, `role.go`, `base.go`, `taggable.go`, `conditional.go` |
| Playbook controller | `pkg/controllers/core/playbook_controller.go` |
| Web services | `pkg/web/service.go` |
| REST proxy | `pkg/proxy/transport.go` |
| Constants/workdir | `pkg/const/common.go`, `pkg/const/workdir.go`, `pkg/const/scheme.go` |

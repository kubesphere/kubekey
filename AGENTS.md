# KubeKey Agent Guide

This file is optimized for AI agents (like OpenCode) working on the KubeKey v4 codebase. It provides a fast path to understand project design, code logic, and conventions. For user-facing documentation, see [README.md](README.md) and [docs/en](docs/en).

## 1. What is KubeKey?

KubeKey v4 is a Go-based task execution framework modeled on Ansible. Its primary use case is installing and managing Kubernetes clusters, but the core engine is generic: it loads playbook projects (YAML), executes tasks across hosts via connectors (SSH/local/Kubernetes/Prometheus), and provides built-in modules (command, copy, template, image, etc.).

Two binaries are produced:

- `kk` – CLI tool that runs playbooks locally or inside a Kubernetes pod.
- `kk-controller-manager` – Kubernetes operator that watches `Playbook` CRs and spawns executor pods.

## 2. Repository Layout

```text
/Users/liujian/code/kubesphere/kubekey
├── cmd/kk                    # CLI binary entry
│   ├── kubekey.go            # main()
│   └── app/
│       ├── root.go           # cobra root command
│       ├── run.go            # "kk run" (arbitrary playbook)
│       ├── playbook.go       # "kk playbook" (in-cluster executor)
│       ├── web.go            # "kk web" (HTTP UI/API server)
│       ├── builtin.go        # built-in commands gated by "builtin" tag
│       ├── builtin/*.go      # create/add/delete/init/precheck/artifact/certs
│       └── options/          # CLI option structs and completion
├── cmd/controller-manager    # Operator binary entry
│   ├── controller_manager.go # main()
│   └── app/                  # controller-runtime manager setup
├── api/                      # Separate Go module for CRD Go types
│   ├── core/v1               # Playbook, Inventory, Config CRDs
│   ├── core/v1alpha1         # Task CRD
│   ├── project/v1            # YAML-parsed project types
│   └── capkk                 # Cluster API provider types
├── pkg/
│   ├── executor/             # Playbook/role/block/task execution engine
│   ├── project/              # Project loading (builtin/local/git)
│   ├── modules/              # Built-in modules (command/copy/template/...)
│   ├── variable/             # Variable merging and lookup
│   ├── connector/            # SSH/local/k8s/prometheus connectors
│   ├── converter/            # Block↔Task conversion, template rendering
│   ├── manager/              # commandManager/controllerManager/webManager
│   ├── controllers/          # Kubernetes reconcilers and webhooks
│   ├── proxy/                # Hybrid REST API proxy (local file + k8s)
│   ├── web/                  # HTTP services (REST + static UI)
│   ├── const/                # Constants, scheme, workdir helpers
│   └── utils/                # Small utilities
├── builtin/core/             # Embedded playbooks/roles (requires "builtin" tag)
│   ├── playbooks/            # create_cluster.yaml, add_nodes.yaml, ...
│   ├── roles/                # native, etcd, kubernetes, cri, certs, ...
│   └── defaults/             # default variables
├── plugins/                  # Optional community playbooks/roles
├── config/                   # Generated CRDs, Helm charts, Kustomize
├── docs/en/framework/        # User-facing framework docs
├── Makefile                  # Build targets, generate, test, lint
├── go.mod                    # Main module
├── go.work                   # Workspace including ./api
└── version/                  # Build-time version injection
```

## 3. Execution Architecture

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│  CLI: kk create cluster / kk run / kk web                                  │
│  Controller: kk-controller-manager watches Playbook CR                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  cmd/kk/app/options/...                                                     │
│  Build kkcorev1.Playbook + Inventory + Config from flags/files             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  pkg/manager/command_manager.go                                             │
│  Creates executor.NewPlaybookExecutor                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  pkg/project/                                                               │
│  Load project: builtin (embed) / local (os.DirFS) / git (go-git)           │
│  MarshalPlaybook parses YAML into kkprojectv1.Playbook                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  pkg/executor/                                                              │
│  playbookExecutor → plays/serial → roleExecutor → blockExecutor            │
│  → taskExecutor creates kkcorev1alpha1.Task CR → module runs per host      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  pkg/modules/ + pkg/connector/                                              │
│  Module uses connector (SSH/local/k8s/prometheus) to act on hosts          │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 4. Entry Points

### CLI

- `cmd/kk/kubekey.go:27` calls `app.NewRootCommand().Execute()`.
- `cmd/kk/app/root.go:29` builds the root `kk` command and adds subcommands:
  - `kk run` (`cmd/kk/app/run.go:25`) – run arbitrary playbook from local path or Git.
  - `kk playbook` (`cmd/kk/app/playbook.go:19`) – in-cluster executor sidecar.
  - `kk web` (`cmd/kk/app/web.go:19`) – start REST/UI server.
  - `kk version` (`cmd/kk/app/version.go`).
  - Built-in commands (`cmd/kk/app/builtin.go` + `cmd/kk/app/builtin/*.go`) gated by the `builtin` build tag.

### Controller Manager

- `cmd/controller-manager/controller_manager.go:24` calls `app.NewControllerManagerCommand().Execute()`.
- Controllers register themselves via `init()` functions gated by build tags, e.g. `pkg/controllers/core/register.go:12`.

## 5. Core Packages

### 5.1 Executor (`pkg/executor/`)

Nested executors by scope:

| File | Executor | Responsibility |
|------|----------|----------------|
| `executor.go:14` | `Executor` interface | `Exec(ctx) error` |
| `playbook_executor.go:61` | `playbookExecutor` | Loads project, iterates plays, handles hosts/serial/gather_facts |
| `role_executor.go:14` | `roleExecutor` | Merges role vars, runs dependencies, executes role blocks |
| `block_executor.go:20` | `blockExecutor` | Handles blocks/rescue/always, tags, when, include_tasks |
| `task_executor.go:29` | `taskExecutor` | Creates `Task` CR, runs module per host in parallel |

Play execution order (`playbook_executor.go:151`):

```text
for each serial batch of hosts:
    pre_tasks
    for each role (with recursive dependencies):
        role blocks
    tasks
    post_tasks
```

### 5.2 Project (`pkg/project/`)

`Project` interface (`project.go:44`):

```go
type Project interface {
    MarshalPlaybook() (*kkprojectv1.Playbook, error)
    Stat(path string) (os.FileInfo, error)
    WalkDir(path string, f fs.WalkDirFunc) error
    ReadFile(path string) ([]byte, error)
    Rel(root string, path string) (string, error)
}
```

`project.New` (`project.go:61`) selects implementation:

- Git-like address → `newGitProject` (`git.go:37`).
- `BuiltinsProjectAnnotation` → `builtinProjectFunc` (`builtin.go:33`).
- Otherwise → `newLocalProject` (`local.go:30`).

`MarshalPlaybook` (`project.go:106`) handles `import_playbook`, `vars_files`, roles (with `meta/dependencies`, `tasks/main.yaml`, `defaults`), `include_tasks`, and validation.

### 5.3 Modules (`pkg/modules/`)

Modules register in `module.go:54`:

```go
utilruntime.Must(internal.RegisterModule(command.ModuleCommand, "command", "shell"))
utilruntime.Must(internal.RegisterModule(copy.ModuleCopy, "copy"))
utilruntime.Must(internal.RegisterModule(template.ModuleTemplate, "template"))
// ... add_hostvars, assert, debug, fetch, gen_cert, http_get_file, image,
//     include_vars, prometheus, result, set_fact, setup
```

Module signature (`internal/options.go:116`):

```go
type ModuleExecFunc func(ctx context.Context, opts ExecOptions) (stdout string, stderr string, err error)
```

A task block's module is discovered from its first `UnknownField` matching a registered module name (`block_executor.go:169`).

### 5.4 Variables (`pkg/variable/`)

`Variable` interface (`variable.go:50`):

```go
type Variable interface {
    Get(getFunc GetFunc) (any, error)
    Merge(mergeFunc MergeFunc) error
}
```

Variable precedence (highest to lowest, `variable_get.go:133`):

1. Config vars
2. Host-specific inventory vars
3. Group vars (for groups containing host)
4. Inventory `vars`
5. Runtime vars (`set_fact`, `register`)
6. Remote vars (gather_facts)

Sources: `MemorySource` and `FileSource` (persists per-host vars under `<workdir>/runtime/.../variable/<hostname>.yaml`).

### 5.5 Connectors (`pkg/connector/`)

`Connector` interface (`connector.go:49`):

```go
type Connector interface {
    Init(ctx context.Context) error
    Close(ctx context.Context) error
    PutFile(ctx context.Context, src []byte, dst string, mode fs.FileMode) error
    FetchFile(ctx context.Context, src string, dst io.Writer) error
    ExecuteCommand(ctx context.Context, cmd string) ([]byte, []byte, error)
}
```

Selection (`connector.go:70`):

- `connector.type=local` → `localConnector`
- `connector.type=ssh` → `sshConnector`
- `connector.type=kubernetes` → `kubernetesConnector`
- `connector.type=prometheus` → `prometheusConnector`
- Default: localhost → local, otherwise SSH.

### 5.6 API Types

- `api/core/v1/playbook_types.go` – `Playbook` CRD.
- `api/core/v1/inventory_types.go` – `Inventory` CRD.
- `api/core/v1/config_types.go` – `Config` CRD.
- `api/core/v1alpha1/task_types.go` – `Task` CRD (per-task execution unit).
- `api/project/v1/` – YAML-parsed `Playbook`, `Play`, `Role`, `Block`, `Base`, `Taggable`, `Conditional`.

The API is a separate Go module referenced via `replace` in the root `go.mod` and included in `go.work`.

## 6. Built-in Playbooks

Built-in content is embedded only when the `builtin` build tag is set (`builtin/core/fs.go:26`). Key playbooks:

- `builtin/core/playbooks/create_cluster.yaml` – full Kubernetes/KubeSphere install.
- `builtin/core/playbooks/add_nodes.yaml`
- `builtin/core/playbooks/delete_cluster.yaml`
- `builtin/core/playbooks/init_os.yaml`
- `builtin/core/playbooks/precheck.yaml`
- `builtin/core/playbooks/certs_renew.yaml`
- `builtin/core/playbooks/artifact_export.yaml`

Key roles under `builtin/core/roles/`:

- `native/` – common OS/package setup.
- `defaults/` – default variables.
- `precheck/` – environment checks.
- `download/` – binary/image downloads.
- `certs/` – certificate generation.
- `etcd/` – etcd deployment.
- `cri/` – container runtime installation.
- `kubernetes/` – kubeadm/kubelet/kubectl setup.
- `cni/` – CNI installation.
- `image-registry/` – private registry setup.

## 7. Build & Test

Important Makefile targets:

| Target | Purpose |
|--------|---------|
| `make kk` | Build `kk` binary with `BUILDTAGS=builtin` |
| `make build-kk-dev` | Dev build with branch-based version |
| `make controller-manager` | Build operator image |
| `make generate` | Deepcopy, CRDs, RBAC, modules, goimports |
| `make generate-manifests-kubekey` | Generate CRDs to `config/kubekey/crds/` |
| `make test` | Run unit/integration tests with envtest |
| `make lint` | Run golangci-lint |

Build tags:

- `builtin` – includes embedded playbooks/roles and built-in CLI commands.
- `clusterapi` – CAPKK controller-manager image build.

## 8. Code Conventions

- Package aliases: `kkcorev1`, `kkcorev1alpha1`, `kkprojectv1`.
- `pkg/const` is imported as `_const` to avoid keyword collision.
- Options structs have `Flags()` and `Complete()` methods.
- Error handling uses `github.com/cockroachdb/errors` with `errors.Wrapf` / `errors.Join`.
- Modules return `(stdout, stderr, err)` triples.
- Controllers and built-in commands register via `init()` gated by build tags.
- Templates use Go `text/template` + Sprig + custom functions in `pkg/converter/tmpl/`.

## 9. Common Tasks for Agents

### Add a new CLI flag

Edit the relevant options struct in `cmd/kk/app/options/` (e.g. `run.go`), add the flag in `Flags()`, and map it in `Complete()`.

### Add a new module

1. Create a package under `pkg/modules/<name>/`.
2. Implement `internal.ModuleExecFunc`.
3. Register it in `pkg/modules/module.go`.
4. Add user docs in `docs/en/framework/modules/<name>.md`.
5. Add unit tests (`<name>_test.go`).

### Add a new built-in playbook/role

1. Add YAML to `builtin/core/playbooks/` or `builtin/core/roles/`.
2. If adding a CLI command, add it in `cmd/kk/app/builtin/*.go` (gated by `builtin` tag).
3. Run `make generate` and `make kk` to verify embedding.

### Add a new controller/webhook

1. Create reconciler under `pkg/controllers/<group>/`.
2. Register it in `pkg/controllers/<group>/register.go` with `options.Register`.
3. Add RBAC markers and run `make generate-manifests-kubekey`.

### Change variable precedence or lookup

Start in `pkg/variable/variable.go` and `pkg/variable/variable_get.go`; persistence is in `pkg/variable/source/`.

## 10. Key File Quick Reference

| Concern | File |
|---------|------|
| CLI root | `cmd/kk/app/root.go` |
| CLI options | `cmd/kk/app/options/option.go` |
| Built-in commands | `cmd/kk/app/builtin/*.go` |
| Controller options | `cmd/controller-manager/app/options/controller_manager.go` |
| Playbook execution | `pkg/executor/playbook_executor.go` |
| Role execution | `pkg/executor/role_executor.go` |
| Block execution | `pkg/executor/block_executor.go` |
| Task execution | `pkg/executor/task_executor.go` |
| Project loading | `pkg/project/project.go` |
| Variable core | `pkg/variable/variable.go` |
| Variable get | `pkg/variable/variable_get.go` |
| Variable merge | `pkg/variable/variable_merge.go` |
| Connector factory | `pkg/connector/connector.go` |
| Module registry | `pkg/modules/module.go` |
| Module internals | `pkg/modules/internal/options.go` |
| Template functions | `pkg/converter/tmpl/functions.go` |
| Playbook CRD | `api/core/v1/playbook_types.go` |
| Inventory CRD | `api/core/v1/inventory_types.go` |
| Task CRD | `api/core/v1alpha1/task_types.go` |
| Project YAML types | `api/project/v1/playbook.go`, `play.go`, `block.go`, `role.go` |
| Playbook controller | `pkg/controllers/core/playbook_controller.go` |
| Web services | `pkg/web/service.go` |
| REST proxy | `pkg/proxy/transport.go` |

## 11. Related Docs

- [README.md](README.md) – user-facing intro.
- [docs/en/framework/README.md](docs/en/framework/README.md) – writing custom playbooks.
- [docs/agent-guide/CODE_LOGIC.md](docs/agent-guide/CODE_LOGIC.md) – detailed code logic flows (companion to this file).

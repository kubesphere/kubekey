# KubeKey Agent Guide

This file contains the **unified rules and conventions** that every AI agent working on the KubeKey v4 codebase must follow.

It is intentionally **not** a project tour or per-role workflow. For project architecture and code logic details, see:

- [README.md](README.md) – user-facing intro.
- [docs/en/framework/README.md](docs/en/framework/README.md) – writing custom playbooks.
- [.opencode/agents/CODE_LOGIC.md](.opencode/agents/CODE_LOGIC.md) – detailed code logic flows.

For per-role instructions, see `.opencode/agents/`:

- [architect.md](.opencode/agents/architect.md) – analyze requirements and produce design documents.
- [developer.md](.opencode/agents/developer.md) – implement code based on the design document.
- [reviewer.md](.opencode/agents/reviewer.md) – review code and produce review/PR artifacts.
- [tester.md](.opencode/agents/tester.md) – plan and execute tests.
- [maintainer.md](.opencode/agents/maintainer.md) – keep documentation, README and CHANGELOG up to date.

## 1. What is KubeKey?

KubeKey v4 is a Go-based task execution framework modeled on Ansible. Its primary use case is installing and managing Kubernetes clusters, but the core engine is generic: it loads playbook projects (YAML), executes tasks across hosts via connectors (SSH/local/Kubernetes/Prometheus), and provides built-in modules (command, copy, template, image, etc.).

Two binaries are produced:

- `kk` – CLI tool that runs playbooks locally or inside a Kubernetes pod.
- `kk-controller-manager` – Kubernetes operator that watches `Playbook` CRs and spawns executor pods.

## 2. Repository Layout

```text
/Users/liujian/code/kubesphere/kubekey
├── cmd/kk                    # CLI binary entry
├── cmd/controller-manager    # Operator binary entry
├── api/                      # Separate Go module for CRD Go types
├── pkg/                      # Core packages
│   ├── executor/             # Playbook/role/block/task execution engine
│   ├── project/              # Project loading (builtin/local/git)
│   ├── modules/              # Built-in modules
│   ├── variable/             # Variable merging and lookup
│   ├── connector/            # SSH/local/k8s/prometheus connectors
│   ├── converter/            # Block↔Task conversion, template rendering
│   ├── manager/              # commandManager/controllerManager/webManager
│   ├── controllers/          # Kubernetes reconcilers and webhooks
│   ├── proxy/                # Hybrid REST API proxy
│   ├── web/                  # HTTP services
│   ├── const/                # Constants, scheme, workdir helpers
│   └── utils/                # Small utilities
├── builtin/core/             # Embedded playbooks/roles (requires "builtin" tag)
├── plugins/                  # Optional community playbooks/roles
├── config/                   # Generated CRDs, Helm charts, Kustomize
├── docs/                     # Documentation
│   └── en/framework/         # User-facing framework docs
├── Makefile                  # Build targets, generate, test, lint
├── go.mod                    # Main module
├── go.work                   # Workspace including ./api
├── version/                  # Build-time version injection
├── .opencode/agents/         # Agent role definitions
└── _output/agents/           # Agent-generated intermediate artifacts
```

## 3. Universal Conventions

All agents must follow these conventions when producing or modifying code.

### 3.1 Logging

Choose the appropriate log level.

| Level | Usage |
|-------|-------|
| `klog.Info` | Main business events. |
| `klog.Warning` | Recoverable abnormal situations. |
| `klog.Error` | Errors requiring attention. |
| `klog.V(4)` | Framework execution flow. Examples: `project`, `proxy`, `variable`, `connector`, `web`, `manager`, `controllers`, `executor`. |
| `klog.V(5)` | Extension modules. Examples: `module`, `converter`. |
| `klog.V(6)` | Debug information. May include detailed intermediate values and execution flow. |

### 3.2 Errors

Wrap errors only where they originate.

- Lower layers should use `errors.Wrap` (or equivalent) to add context.
- Upper layers should return the error directly unless adding meaningful business context.
- Do not repeatedly wrap the same error.

KubeKey uses `github.com/cockroachdb/errors` with `errors.Wrapf` / `errors.Join`.

### 3.3 Naming

Keep names concise. Prefer meaningful short names. Avoid unnecessary abbreviations and verbose names.

Avoid:

```go
tmpData
managerObject
projectConfiguration
```

Prefer:

```go
cfg
proj
mgr
conn
```

### 3.4 Architecture

Prefer modifying existing code instead of introducing new abstractions.

- Do not introduce new structs or interfaces unless there is a clear benefit.
- Keep APIs stable.
- Minimize public surface.
- Favor composition over inheritance-like patterns.
- Do not repeat inherited conditions (e.g. `when`) at every level; declare them at the highest applicable scope.

### 3.5 Go Conventions

- Package aliases: `kkcorev1`, `kkcorev1alpha1`, `kkprojectv1`.
- `pkg/const` is imported as `_const` to avoid keyword collision.
- Options structs have `Flags()` and `Complete()` methods.
- Modules return `(stdout, stderr, err)` triples.
- Controllers and built-in commands register via `init()` gated by build tags.
- Templates use Go `text/template` + Sprig + custom functions in `pkg/converter/tmpl/`.

## 4. Build & Test

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

## 5. Agent Workflow

Agents collaborate through files under `_output/agents/`. The directory is already ignored by Git and is meant for cross-session handoff only.

```text
_output/agents/
├── design.md       # Architect output
├── dev-summary.md  # Developer output
├── review.md       # Reviewer output
├── pr.md           # Reviewer output
├── commit.txt      # Reviewer output
├── test-plan.md    # Tester output
└── test-result.md  # Tester output
```

Pipeline:

1. **Architect** reads the requirement and writes `_output/agents/design.md`.
2. **Developer** reads `design.md` and writes code + `_output/agents/dev-summary.md`.
3. **Reviewer** reads `design.md` and `dev-summary.md`, reviews the code, and writes `_output/agents/review.md`, `_output/agents/pr.md`, `_output/agents/commit.txt`.
4. **Tester** reads `design.md` and `dev-summary.md`, writes `_output/agents/test-plan.md`, runs tests, and writes `_output/agents/test-result.md`.
5. **Maintainer** reads `dev-summary.md` and `review.md`, then updates README, CHANGELOG, docs and examples.

Each role is defined in `.opencode/agents/<role>.md`. Agents must stay within their own responsibilities.

## 6. Git Commit & PR Conventions

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>: <short description>
```

Common types:

- `feat:` – new feature
- `fix:` – bug fix
- `refactor:` – code refactoring
- `docs:` – documentation only
- `test:` – tests
- `chore:` – build, dependencies, tooling

Examples:

```text
feat: support multiple ssh private keys
fix: preserve proxy configuration during reconnect
```

### Pull Request Description

A PR description should include:

- **What** changed and **why**.
- **How** it was implemented (briefly).
- **Testing** performed.
- **Risks / breaking changes**.

The Reviewer agent generates `pr.md` and `commit.txt` based on the developer summary and review findings.

## 7. Related Docs

- [README.md](README.md)
- [docs/en/framework/README.md](docs/en/framework/README.md)
- [.opencode/agents/CODE_LOGIC.md](.opencode/agents/CODE_LOGIC.md)
- `.opencode/agents/architect.md`
- `.opencode/agents/developer.md`
- `.opencode/agents/reviewer.md`
- `.opencode/agents/tester.md`
- `.opencode/agents/maintainer.md`

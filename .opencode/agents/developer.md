# Developer Agent

## Role

Implement the design described in `_output/agents/design.md`.

## Input

- `_output/agents/design.md`
- [AGENTS.md](../../AGENTS.md) – universal conventions.
- [CODE_LOGIC.md](./CODE_LOGIC.md) – detailed code logic flows.

## Output

- Code changes in the repository.
- `_output/agents/dev-summary.md`

## Responsibilities

- Read `design.md` thoroughly.
- Implement the design with minimal, clean changes.
- Keep code simple.
- Optimize for performance only when necessary.
- Avoid over-engineering.
- Challenge the design if it turns out to be infeasible or overly complex during implementation.

## Constraints

- Do not write design documents.
- Do not generate PR descriptions or commit messages (Reviewer does that).
- Do not generate test plans (Tester does that), though unit tests for new code are encouraged.

## Principles

### 1. Prefer Modifying Existing Code

Do not create new packages, managers, interfaces, or utilities just to add a feature unless there is a real need.

### 2. Interfaces Must Earn Their Place

Before adding a new interface, ensure it satisfies at least one of:

- Decouples two modules.
- Improves testability.
- Has more than one implementation.

### 3. Keep APIs Stable

Avoid changing public function signatures or CRD schemas unless the design explicitly requires it.

### 4. Follow Go Idioms

- Handle errors explicitly.
- Avoid naked returns.
- Prefer early returns.
- Keep functions short and focused.

## Output Template: `_output/agents/dev-summary.md`

```markdown
# Development Summary: <title>

## Modified Files

| File | Change |
|------|--------|
| pkg/... | ... |

## Rationale

Brief explanation of why each change was made.

## Impact Scope

- Affected packages
- Affected user-facing behavior

## Areas to Review

- List any sections that need extra attention from the Reviewer.

## Known Limitations

- Anything intentionally left out or deferred.
```

## Common Tasks Reference

### Add a new CLI flag

Edit the relevant options struct in `cmd/kk/app/options/` (e.g. `run.go`), add the flag in `Flags()`, and map it in `Complete()`.

### Add a new module

1. Create a package under `pkg/modules/<name>/`.
2. Implement `internal.ModuleExecFunc`.
3. Register it in `pkg/modules/module.go`.
4. Add user docs in `docs/en/framework/modules/<name>.md`.
5. Add unit tests (`<name>_test.go`).

A module must be **decoupled from the local OS**. It runs against a target host through the `Connector` interface (`pkg/connector/connector.go`), which may be local, SSH, Kubernetes, or Prometheus. Therefore:

- Do **not** assume the target is Linux or that a shell is available.
- Do **not** use Bash-specific syntax (`&&`, `||`, pipes, here-docs) unless the module is explicitly shell-oriented.
- Do **not** hard-code Linux paths such as `/usr/local/bin`, `/etc`, `/tmp`, or `/opt`.
- Use `Connector.ExecuteCommand` for remote commands, `Connector.PutFile` for uploading files, and `Connector.FetchFile` for downloading files.
- When paths are required, accept them as module arguments with sensible defaults. Use the path package (which always uses /) for target paths on Unix-like systems, and reserve path/filepath strictly for local host paths to avoid path separator mismatches when running kk from a different OS (e.g., Windows/macOS) than the target.
- If a shell is necessary, prefer the most portable subset (`sh` over `bash`) and document the requirement clearly.

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

## Playbook / Role Writing Tips

### Avoid repeating `when` conditions

`when` conditions are inherited from role → block → task. Do not repeat the same condition at every level. Declare it once at the highest applicable scope:

- Use role-level `when` if the condition applies to the entire role.
- Use block-level `when` if the condition applies to a group of tasks.
- Use task-level `when` only for task-specific gating.

This keeps playbooks concise and easier to maintain.

### Avoid repeating `tags`

`tags` are inherited the same way as `when` (role → block → task). Declare tags at the highest applicable scope and let nested blocks/tasks inherit them. Do not repeat the same tag at every level.

Special tags:

- `always` – runs unless explicitly skipped.
- `never` – never runs unless explicitly included.
- `all` / `tagged` – meta tags used in `--tags` / `--skip-tags` filters.

#### Common pitfall

A task runs only when **all inherited `when` conditions** evaluate to true. Adding the same condition again at a lower level does not make the task "more likely" to run; it only creates redundant checks. If a task should not be governed by a parent `when`, move it out of that role/block instead of duplicating or contradicting the condition.

## Key File Quick Reference

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
| Project YAML types | `api/project/v1/playbook.go`, `play.go`, `block.go`, `role.go`, `base.go`, `taggable.go`, `conditional.go` |
| Playbook controller | `pkg/controllers/core/playbook_controller.go` |
| Web services | `pkg/web/service.go` |
| REST proxy | `pkg/proxy/transport.go` |

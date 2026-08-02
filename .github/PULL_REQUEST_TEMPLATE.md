<!--
Thanks for sending a pull request! Here are some tips for you:

1. If you want **faster** PR reviews, read how: https://github.com/kubesphere/community/blob/master/developer-guide/development/the-pr-author-guide-to-getting-through-code-review.md
2. In case you want to know how your PR got reviewed, read: https://github.com/kubesphere/community/blob/master/developer-guide/development/code-review-guide.md
3. Here are some coding conventions followed by KubeSphere community: https://github.com/kubesphere/community/blob/master/developer-guide/development/coding-conventions.md
-->

### What type of PR is this?
<!--
Add one of the following kinds:
/kind bug
/kind cleanup
/kind documentation
/kind feature
/kind design
/kind dependencies
/kind test

Optionally add one or more of the following kinds if applicable:
/kind api-change
/kind deprecation
/kind failing-test
/kind flake
/kind regression
-->

## What does this PR do

<One sentence describing what this PR does, echoing the commit subject>

### Background / Motivation

<Why this change is needed: what problem or requirement prompted it. Don't restate the code.>

### Implementation

<Core approach, and why this was chosen over alternatives.>

### Key Changes

| File | What changed |
|---|---|
| `<path/to/file>` | <what was done in this file> |

## Impact

- **Affected modules**: <module / package name>
- **API changes**: <list before/after signatures if any; N/A if none>
- **Database / state changes**: <migration or state change + rollback if any; N/A if none>
- **Config changes**: <new / modified config items and defaults; N/A if none>
- **Dependency changes**: <new / upgraded deps and versions; N/A if none>

## Breaking Changes

<Detail the change and caller migration steps if any; write "None" otherwise>

### Which issue(s) this PR fixes:
<!--
Usage: `Fixes #<issue number>`, or `Fixes (paste link of issue)`.
_If PR is about `failing-tests or flakes`, please post the related issues/tests in a comment and do not use `Fixes`_*
-->
Fixes #

## Testing

### Verification performed

- [ ] Unit tests pass (`<command>`)
- [ ] Integration / E2E tests pass (`<command>`)
- [ ] Manual verification

### Steps to verify

1. <step>
2. <step>
3. <expected result>

### Test coverage

<New / modified test cases, or why no tests are needed>

## Rollback

<How to roll back if it goes wrong: plain revert / additional steps (data rollback, config restore, toggle off)>

### Does this PR introduce a user-facing change?
<!--
If no, just write "None" in the release-note block below.
If yes, a release note is required:
Enter your extended release note in the block below. If the PR requires additional action from users switching to the new release, include the string "action required".

For more information on release notes see: https://github.com/kubernetes/community/blob/master/contributors/guide/release-notes.md
-->
```release-note

```

## Checklist

- [ ] Code self-reviewed, no debug code or commented-out dead code
- [ ] No secrets, tokens, or `.env` files committed
- [ ] Commit message follows Conventional Commits
- [ ] All commits have DCO sign-off (`Signed-off-by`)
- [ ] All commits are GPG-signed (GitHub shows Verified)
- [ ] Tests added or updated
- [ ] Docs / CHANGELOG updated (if needed)
- [ ] Local lint and build pass (`make build`)
- [ ] Breaking changes marked above

### Additional documentation, usage docs, etc.:
<!--
This section can be blank if this pull request does not require a release note.
Please use the following format for linking documentation or pass the
section below:
- [KEP]: <link>
- [Usage]: <link>
- [Other doc]: <link>
-->
```docs

```

## Notes for Reviewer

<What you want the reviewer to focus on, known trade-offs or open questions; N/A if none>

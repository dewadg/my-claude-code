---
name: glab-release
description: >-
  Cut a versioned GitLab release with the glab CLI: pick the previous tag,
  compare commits since then on the default branch, build a changelog grouped
  by conventional-commit type, and create the release (which pushes the tag and
  triggers the tag pipeline). Use this whenever the user wants to "create a
  release", "cut a release", "tag a version", "release X.Y.Z", "ship a new
  version", "publish a release", or asks for a changelog/release notes for a new
  tag — even if they don't name glab explicitly. Prefer this over a raw
  `git tag` when the project uses GitLab and wants release notes.
---

# glab Release

Create a GitLab Release from the default branch with an auto-generated,
grouped changelog. Mirrors how releases are cut by hand, but without missing
the safety checks that bite later (wrong branch, out-of-sync local, tag that
silently re-triggers a deploy pipeline).

## When to use

Trigger on intent to release/tag a new version: "create a release", "cut
1.9.0", "tag a version", "release notes for the next version". If the user only
wants a local annotated tag with no GitLab Release page and no notes, a plain
`git tag` is enough — but most release requests want the notes, so default to
this skill.

## Why a release matters here

A GitLab Release creates the tag. In most pipelines the tag is what triggers
build/deploy (`only: tags`). So creating a release is not just bookkeeping — it
can build + push images and surface manual deploy jobs. Surface that to the
user before pushing, because it is hard to undo a published tag + pipeline.

## Required input

- **New version** — the tag to create (e.g. `1.9.0`). Ask if not given.
- **Previous version** — the tag to diff against for the changelog. If not
  given, default to the latest existing tag (`git describe --tags --abbrev=0`)
  and tell the user which one you picked.

## Steps

### 1. Preflight (read-only — never skip)

Confirm the tooling and repo state before touching anything remote.

```bash
glab --version                       # glab installed?
git remote -v | head -1              # GitLab origin present?
git symbolic-ref --short refs/remotes/origin/HEAD | sed 's#origin/##'  # default branch
git tag -l "<prev>"                  # previous tag exists?
git tag -l "<new>"                   # new tag must NOT already exist
```

Resolve the real default branch — do not assume `master`. Many repos use
`main`. If `origin/HEAD` is unset, fall back to checking whether `main` or
`master` exists.

### 2. Sync check

The release is cut from the **remote** branch tip, so confirm local matches
remote — otherwise the changelog reflects commits the server doesn't have, or
misses commits it does.

```bash
git fetch origin <default-branch> -q
git rev-parse HEAD
git rev-parse origin/<default-branch>
```

If they differ, stop and tell the user (local ahead → unpushed work; local
behind → stale). Let them decide before proceeding.

### 3. Build the changelog

List non-merge commits since the previous tag, on the remote branch tip:

```bash
git log --no-merges --pretty="%s" <prev>..origin/<default-branch>
```

Group subjects by conventional-commit prefix into these headings (omit any
heading with no commits). Strip the `type(scope):` prefix from each line and
keep the human-readable summary:

| Prefix | Heading |
|---|---|
| `feat` | ### Features |
| `fix` | ### Fixes |
| `perf` | ### Performance |
| `refactor` | ### Refactor |
| `ci` | ### CI/CD |
| `build` | ### Build |
| `docs` | ### Docs |
| `test` | ### Tests |
| `chore` / other | ### Chore |

Order headings as in the table (user-facing first). For commits with no
recognizable prefix, put them under Chore rather than dropping them — a commit
missing from the notes is worse than one in the wrong bucket.

End the notes with a compare link so readers can see the full diff. Derive the
project path from the origin URL:

```
**Full diff:** https://gitlab.com/<group>/<project>/-/compare/<prev>...<new>
```

### 4. Confirm, then create

Show the user the assembled notes + the fact that the tag will trigger the tag
pipeline. On approval:

```bash
glab release create <new> --ref <default-branch> --name "<new>" --notes "<changelog>"
```

`glab release create` makes the tag if it doesn't exist, targeting the branch
tip. No separate `git tag`/`git push` needed.

### 5. Confirm the pipeline (optional)

The tag push starts a pipeline. Surface it so the user knows build/deploy is in
flight:

```bash
glab ci list -P 3 | head -8
```

If deploy jobs are `when: manual`, note they wait for a manual trigger and
won't auto-run.

## Example

**Input:** "create a release tagged 1.9.0, compare against 1.8.3"

**Changelog produced:**

```markdown
## Changelog (1.8.3 → 1.9.0)

### CI/CD
- Add Swarm deploy jobs

### Build
- Upgrade to Go 1.26

### Refactor
- Improve testability

### Docs
- Update docs

**Full diff:** https://gitlab.com/dewadg/random-repo/-/compare/1.8.3...1.9.0
```

**Command:**

```bash
glab release create 1.9.0 --ref main --name "1.9.0" --notes "<changelog above>"
```

## Footguns

- **Wrong branch name** — `master` vs `main`. Always resolve the real default.
- **Out-of-sync local** — diffing local when remote differs gives a wrong
  changelog. Fetch + compare SHAs first.
- **Re-using an existing tag** — `glab release create` on an existing tag can
  fail or point the release at an unexpected commit. Check `git tag -l` first.
- **Silent pipeline trigger** — the tag fires `only: tags` jobs. Tell the user;
  a published tag is hard to walk back.

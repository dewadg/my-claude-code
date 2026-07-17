---
name: pr-description
description: Write the description/body for a pull request or merge request from the actual branch diff, in a fixed emoji-sectioned markdown template (Overview, Related Issues, Technical Changes, Testing). Use this whenever the user is about to open or is preparing a PR/MR on any host (GitHub, GitLab, Bitbucket) — including phrasings like "create a PR", "raise an MR", "write the PR body", "what should I put in the description", "summarize this branch for review", or when they hand you a diff and ask for reviewer-facing prose. This skill produces the description text only; it never opens the PR/MR itself.
---

# PR/MR description

You are writing the text a reviewer reads *before* they read a single line of the diff. That reader is cold: they don't know the branch, the ticket, or why any of it exists. Your job is to make the diff make sense in about ninety seconds.

The output is **the description only**. Do not run `gh pr create`, `glab mr create`, or any GitLab/GitHub MCP write tool — the caller decides where the text goes. Print the finished markdown in a fenced block so it can be copied verbatim, and stop.

## 1. Gather the real evidence

Never write from the branch name or commit subjects alone — they lie by omission. Read the diff.

```bash
git merge-base --fork-point origin/main HEAD || git merge-base origin/main HEAD   # base commit
git log --oneline <base>..HEAD
git diff --stat <base>..HEAD
git diff <base>..HEAD                                                             # the actual hunks
```

Determine the base branch from the repo rather than assuming: check `git symbolic-ref refs/remotes/origin/HEAD`, or the default branch shown by the host. `main` and `master` are both common; some repos target `develop`.

If the diff is large, still read every changed file's hunks — but read them looking for *intent*, not lines. You are trying to answer "what decision did the author make, and what did it force them to touch?" A rename that sprays across forty files is one sentence, not forty bullets.

If the user gave you a diff or a range directly, use that instead of inferring one.

## 2. Find the linked issue

Try these in order, stopping at the first hit:

1. **Branch name** — `feat/glab54`, `fix/ISSUE-123`, `1234-add-rbac` all encode a ticket number.
2. **Commit messages** — `Closes #88`, `glab54`, `JIRA-12` in subjects or trailers.
3. **Spec/proposal files touched by the diff** — e.g. an OpenSpec change's `proposal.md` often carries a `## References` line with the issue URL.
4. **Ask the user** — if none of the above yields anything, ask once for the ticket URL rather than inventing one or silently dropping the section.

Once you have a number, fetch the real title and URL (GitLab/GitHub MCP tools, or `gh issue view` / the API) rather than guessing at a title. A wrong title in the table is worse than no table. If the issue genuinely doesn't exist — a chore branch, a hotfix — omit the whole **Related Issues** section instead of leaving an empty table.

## 3. Write it

Use this template exactly, including the emoji. Fixed structure is the point: reviewers on this team scan for the same headings every time, so they know where to look without reading.

````markdown
## 🎯 Overview
[What this PR accomplishes and why, for someone seeing it cold.]

## 🔗 Related Issues
| Issue | Title |
|-------|-------|
| [{ticket-number}]({ticket-url}) | [summary] |

## 🔧 Technical Changes

### 🔄 Core logic
[Narrative paragraph.]
- `path/to/file` — what changed

```ts
// before → after (only when it clarifies)
```

### 📤 Context / data flow
[Narrative + file bullets as needed.]

### 📝 Messages / UX
[…]

### 🔒 Security / dependencies
[…]

## 🧪 Testing
- [ ] [How it was/should be verified]
````

### The subsections are adaptive

`Core logic`, `Context / data flow`, `Messages / UX`, `Security / dependencies` are the *usual* buckets, not a mandatory checklist. Include a subsection only when the diff actually has something for it — a two-file bugfix that touches no auth and no copy should not carry a `🔒 Security / dependencies` heading saying "None". Empty headings train reviewers to skim past headings.

By the same token, if the diff has a real category none of the four names — database migrations, config/infra, generated code, test-only scaffolding, a performance rework — add a subsection for it with a fitting emoji and a clear name. The four are a starting vocabulary, not a ceiling.

Order subsections so the most consequential change comes first. That is usually `Core logic`, but not always; a PR whose whole point is a dependency bump should lead with it.

## What good looks like in each section

**🎯 Overview** — two to five sentences of prose. State what the branch does *and why it was worth doing*. The "why" is the part the diff can't tell the reviewer: a bug's user-visible symptom, a ticket's business ask, a constraint that forced the approach. Avoid restating the commit log; avoid the word "This PR" three times.

**🔧 Technical Changes** — each subsection opens with a short narrative paragraph explaining the shape of the change, *then* the file bullets. The paragraph is what a reviewer reads; the bullets are what they navigate by. A bare list of files with "updated X" next to each is a changelog, not a description — it tells the reviewer nothing they couldn't get from `git diff --stat`.

File bullets: `` `path/to/file` — what changed ``. One line, present tense, specific. "Added handler" is noise; "adds `GET /api/health`, returns 200 with DB ping result" is signal. Group related files under one bullet when they moved as a unit.

Code snippets: include one only when a before/after makes something click faster than prose would — a subtle condition flip, a changed signature, a new payload shape. A snippet that just repeats a hunk the reviewer will read in thirty seconds anyway is padding. Most subsections need zero snippets.

**🧪 Testing** — checkboxes describing how this was verified, or how the reviewer should verify it. Be concrete: name the test package or command (`go test -race ./internal/health/`), the endpoint you curled, the UI flow you clicked. If a change is genuinely untested, say so honestly as an unchecked box — a reviewer would rather see `- [ ] No automated coverage; verified manually against staging` than a checked box that's a lie.

## Tone

Write plain declarative sentences. No "This PR aims to", no "Additionally, it should be noted", no marketing. Match the repo's own vocabulary — if the codebase says "conversation" and "operator", don't invent "chat session" and "agent". If a project `CLAUDE.md` or contributing guide describes domain terms, borrow them.

Length should track the diff. A one-file fix gets an Overview, one subsection, and a Testing box — forcing it into four subsections is worse than useless. A cross-cutting feature earns the full structure.

## Before you print

Reread it as the reviewer: does the Overview alone tell them whether they're the right person to review this? Can they find the riskiest hunk from the file bullets? If a ticket is linked, is the title real?

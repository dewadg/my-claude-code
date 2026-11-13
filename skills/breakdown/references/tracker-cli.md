# Tracker CLIs

Use `glab` (GitLab) and `gh` (GitHub). Fall back to the gitlab MCP only when the CLI is missing or
unauthenticated. Pick the tracker from the URL/identifier, or `git remote -v`.

## Read (free to use)

```bash
glab issue view 42 --comments
glab issue list --label "feature" --state opened --per-page 100
glab issue list --milestone "Q4" --output json        # backlog dump for --groom
glab label list

gh issue view 42 --comments
gh issue list --label feature --state open --limit 200 --json number,title,labels,body,updatedAt
gh search issues --repo org/repo "is:open label:feature"
gh label list
```

## Write (only after `--write` and an explicit yes)

Bodies go in files — never inline multi-line markdown in a shell argument.

```bash
glab issue create --title "[BE] POST /refunds — issue a partial refund" \
  --description "$(cat ticket-03.md)" --label "backend,size::M" \
  --linked-issues 42 --link-type relates_to
glab issue update 42 --description "$(cat parent.md)"     # parent: separate confirmation

gh issue create --title "[FE] Order detail — refund entry point" \
  --body-file ticket-06.md --label frontend,size:M
gh issue edit 42 --body-file parent.md                     # parent: separate confirmation
```

## Rules

- List the repo's labels first and reuse its scheme (seam, size, priority) — don't invent labels.
  No size label scheme exists → put the size in the body, don't create labels without asking.
- Children carry `Parent: #42` in Context/Background.
- GitHub sub-issues (when the repo uses them):
  `gh api repos/{owner}/{repo}/issues/42/sub_issues -F sub_issue_id=<child id>` — the child's
  numeric `id` (`gh issue view <n> --json id` gives a node id; use
  `gh api repos/{owner}/{repo}/issues/<n> --jq .id` for the numeric one).
- Create in dependency order so `blocked by` can reference real numbers; then patch references if
  needed.
- Report every created URL. On a mid-run failure, stop, report what was created, don't retry blind.

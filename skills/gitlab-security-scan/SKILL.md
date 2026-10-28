---
name: gitlab-security-scan
description: >-
  Set up security scanning in a project's GitLab CI on the Free tier: per-package
  Semgrep SAST jobs, an osv-scanner dependency vulnerability scan with an expiring
  allowlist, one summarized JUnit scan report in the pipeline Tests tab, and
  Renovate dependency-update MRs. Use whenever the user wants to "add SAST",
  "add security scans to CI", "scan dependencies for vulnerabilities", "set up
  osv-scanner", "add Renovate", "show scan results in the Tests tab", or replicate
  an existing project's security pipeline into another GitLab repo. Also use to
  triage the first findings those jobs report. Not for GitHub Actions, not for a
  one-off local security review of a diff (use security-review), and not for
  cutting releases (use glab-release).
---

# GitLab Security Scan

Adds four jobs to a GitLab project's `.gitlab-ci.yml`, all in stage `test`, all
running only on pushes to the default branch (a merged MR), never in MR pipelines
or on tags, so a finding never blocks development:

| Job | Tool | Gates? | Report |
|---|---|---|---|
| `sast <pkg>` (one per package) | GitLab `Jobs/SAST` template, Semgrep analyzer | no | `scan-reports/sast-<pkg>.json` |
| `dependency scan` | osv-scanner v2, pinned | yes: unallowlisted, called advisory fails | `scan-reports/osv-scanner.json` |
| `scan report` | `.gitlab/scan-junit.py` | no | JUnit, pipeline Tests tab |
| `renovate` | `renovate-bot/renovate-runner` template | no | update MRs + Dependency Dashboard issue |

Reference implementation: `karambyte/chatto` on gitlab.com (`.gitlab-ci.yml`,
`.gitlab/`, `renovate.json`, `docs/dependency-scanning.md`).

## Assets

All under this skill's `assets/`. Copy, then fill `<placeholders>`.

| Asset | Goes to | Edit |
|---|---|---|
| `security-scan.gitlab-ci.yml` | merge into `.gitlab-ci.yml` | placeholders listed in its header |
| `scan-junit.py` | `.gitlab/scan-junit.py` | none: report names are CLI args |
| `osv-scanner.toml` | `.gitlab/osv-scanner.toml` | none until a finding needs an allowlist entry |
| `renovate.json` | `renovate.json` (repo root) | one `packageRules` group per package |

## Workflow

### 1. Survey the target

- Packages: each deployable dir and its language. One package = one `sast <pkg>` job.
  A single-package repo still gets one job (`sast app`) so report naming stays uniform.
- Lockfiles osv-scanner reads: `go.mod`, `package-lock.json`, `pnpm-lock.yaml`,
  `yarn.lock`, `requirements.txt`, `poetry.lock`, `Cargo.lock`, `composer.lock`,
  `Gemfile.lock`. List every one; each becomes a `-L`.
- Existing `.gitlab-ci.yml`: `include:` and `stages:` blocks (merge, don't duplicate;
  add `test` first), how existing jobs gate (`rules:` style), section comment style.
- Dirs that hold no runtime code (`openspec/`, `.notes/`, `deploy/`, `docs/`): add to
  `SAST_COMMON_EXCLUDED_PATHS`. Check no runtime code lives under an excluded dir
  name (`test`, `tests`, `__tests__`, `spec`, `tmp`, `e2e`, `docs`, `deploy`) or uses
  a `*.config.*` file name: it would go unscanned.
- Host: `renovate-bot/renovate-runner` is a gitlab.com project. Self-managed GitLab
  must mirror it or drop the `include`.

Ask the user only which of the four jobs to skip; default is all four.

### 2. SAST jobs

Per package, `SAST_EXCLUDED_PATHS` = common list + every other package dir. Nested
packages (e.g. `server/web/ui` inside a Go `server/`): the parent job excludes the
child dir, each child job excludes its siblings plus the parent's language by
extension (`*.go`). Chatto example:

```yaml
sast server:
  extends: .sast-package
  variables:
    SAST_EXCLUDED_PATHS: "$SAST_COMMON_EXCLUDED_PATHS, packages/chatto-server/web, packages/chatto-mobile"

sast web-ui:
  extends: .sast-package
  variables:
    SAST_EXCLUDED_PATHS: "$SAST_COMMON_EXCLUDED_PATHS, *.go, packages/chatto-server/web/widget, packages/chatto-mobile"
```

The report name is the job name slugged: `sast web-ui` -> `sast-web-ui`.

### 3. Dependency scan

One `-L` per lockfile, same list in the JSON pass and the gating pass. Keep the
`--all-vulns` informational line only for Go modules. Cache key: `go.sum` of the Go
module, else a fixed key `osv-scanner-v2.7.0`.

Run it locally before committing (osv-scanner from `go install
github.com/google/osv-scanner/v2/cmd/osv-scanner@v2.7.0`):

```bash
osv-scanner scan source --config .gitlab/osv-scanner.toml -L <lockfile-1> -L <lockfile-2>; echo "exit=$?"
```

Exit `0` clean, `1` findings, `127` a `-L` path that does not exist. Each finding
now is a finding on the first `main` pipeline: bump it or allowlist it (step 7) in
the same MR.

### 4. Scan report

`needs:` lists every `sast <pkg>` job plus `dependency scan`; the script args list
the same reports: `scan-report.xml sast-<pkg>... osv-scanner`. A name with no
`scan-reports/<name>.json` shows as an error case, so a forgotten arg is silent but a
forgotten job is not. Test it against local output:

```bash
mkdir -p scan-reports
osv-scanner scan source --config .gitlab/osv-scanner.toml --format json --output-file scan-reports/osv-scanner.json -L <lockfile-1>
python3 .gitlab/scan-junit.py scan-report.xml osv-scanner sast-<pkg>
# osv-scanner      findings=0 uncalled=0 missing=0
# sast-<pkg>       findings=0 uncalled=0 missing=1   <- no local SAST report: expected
rm -rf scan-reports scan-report.xml
```

### 5. Renovate

Fill `RENOVATE_REPOSITORIES` with the project path and one `packageRules` group per
package dir (`matchManagers` `gomod`, `npm`, `pip_requirements`, `cargo`, ...). Drop
`postUpdateOptions: ["gomodTidy"]` when there is no Go module. Keep the timezone the
user works in.

### 6. Lint

From the target repo (needs `glab auth login` to its host):

```bash
glab ci lint .gitlab-ci.yml
# ✓ CI/CD YAML is valid!
```

### 7. Docs and conventions

Write `docs/security-scanning.md` (chatto: `docs/dependency-scanning.md`) covering:
what each job scans and when, reading a failed `dependency scan` (bump vs allowlist),
the allowlist rules (`id` + `reason` + `ignoreUntil` <= 90 days, entries apply to every
lockfile), Renovate behaviour, and the maintainer steps below. Add one line to the
target's `CLAUDE.md`:

> **SAST**: per-package Semgrep runs on pushes to main. Scope lives in
> `SAST_COMMON_EXCLUDED_PATHS` in `.gitlab-ci.yml`. Project ruleset files need
> Ultimate, so false positives get `// nosemgrep: <rule-id> -- <reason>` on the line
> directly above. Runtime code must not use `*.config.*` file names or live under
> excluded dir names (unscanned).

### 8. Maintainer steps (tell the user; they need GitLab UI access)

1. Settings > Access tokens: project access token, role Developer, scopes `api` and
   `write_repository`, 1-year expiry.
2. Settings > CI/CD > Variables: `RENOVATE_TOKEN`, masked and protected.
3. CI/CD > Schedules: one daily schedule on the default branch (security MRs within
   24h; routine ones still batch weekly via `renovate.json`). Play it once and check
   the Dependency Dashboard issue appears.
4. Token expires after 365 days max: `renovate` fails with an auth error, rotate it.

### 9. Triage the first `main` pipeline

Open the Tests tab, or pull artifacts with glab (from the target repo):

```bash
glab ci list -P 5
glab api "projects/:id/pipelines/<pipeline-id>/jobs" | python3 -c 'import json,sys; [print(j["id"], j["name"], j["status"]) for j in json.load(sys.stdin)]'
glab api "projects/:id/jobs/<job-id>/artifacts" > a.zip && unzip -o -q a.zip && ls scan-reports
```

Per finding, in order of preference:
- **Real bug**: fix the code (chatto: gosec G601 `&loopVar` -> index the slice `&items[i]`).
- **False positive**: `// nosemgrep: <rule-id> -- <reason>` on the line directly above
  (rule id from the report's `identifiers`, e.g. `nodejs_scan.javascript-ssrf-rule-node_ssrf`
  on a browser-side `fetch`).
- **Test/tooling code flagged**: widen `SAST_COMMON_EXCLUDED_PATHS`.
- **Vulnerable dependency**: bump to the fixed version and run that package's tests;
  if no fix exists, allowlist in `.gitlab/osv-scanner.toml` with a reason naming the
  lockfile.

## Gotchas

- **`.gitlab/sast-ruleset.toml` is ignored below Ultimate.** The analyzer logs
  "ruleset customization not enabled" and the disabled rules keep firing. Suppress
  per line with `nosemgrep` instead.
- **The template's `semgrep-sast` and `gitlab-advanced-sast` jobs must be switched off**
  (`rules: - when: never`), or they run repo-wide next to the per-package jobs.
  `.sast-package` still `extends: semgrep-sast` and overrides its `rules`, so it
  inherits the analyzer image and script but not the "never".
- **`SAST_EXCLUDED_PATHS` replaces the template default**, it does not append. That
  is why every job carries `$SAST_COMMON_EXCLUDED_PATHS`.
- **Two osv passes, not one.** `--format json` writes JSON *instead of* the table, so
  the JSON pass is `|| true` and a second plain pass gates and prints the readable
  table in the log.
- **Go call analysis needs the Go toolchain**: the official osv-scanner image has none,
  hence `golang:*-alpine` + `go install`. Advisories in uncalled code exit `0`; they
  show as skipped cases in the Tests tab.
- **`go install` of osv-scanner can need a newer Go than the image** (chatto hit this on
  `golang:1.25-alpine`). Prefix the install with `GOTOOLCHAIN=auto` (install only; the
  scans keep the image's Go) or bump the image.
- **`dependency scan` artifacts need `when: always`**, otherwise a failing scan uploads
  nothing and the report shows it as missing.
- **`scan report` rule needs `when: always`** so it runs after a failed `dependency scan`.
- **Allowlist entries are global** across every `-L` lockfile. A "dev-only" reason
  written for one lockfile also hides the same advisory in a shipped one.
- **Renovate on main pushes** too, not only schedules, in stage `test` (the template
  defaults to `deploy`, which a failing scan would skip).
- **Main-only rules mean a Renovate MR is never scanned before merge.** A bump pulling
  in a vulnerable transitive dependency turns `main` red after merge. To scan MRs too,
  add `- if: '$CI_PIPELINE_SOURCE == "merge_request_event"'` to the scan jobs' rules.

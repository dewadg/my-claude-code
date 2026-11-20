---
name: defect-report
description: >
  Emit review findings and QA defects in the shared severity-tagged format: one line per finding
  (path:line, tier, problem, fix), ordered by severity, then totals and a one-sentence verdict.
  ALWAYS use when a code review or QA run ends with findings, or the user asks for findings in a
  standard format. Not for the review itself, notes (note-write), or RCAs (rca-write).
allowed-tools: Read
effort: low
---

Owns the finding-report format shared by the review and QA agents (and any agent emitting findings).
Every review or test run ends in this format so downstream agents and tooling can parse findings
deterministically. No prose preamble, no praise filler, no "looks good".

## Severity tiers

| Emoji | Tier | Use for |
|---|---|---|
| 🔴 | bug | Wrong output, crash, security hole, data loss |
| 🟡 | risk | Edge case, race, leak, perf cliff, missing guard, design flaw |
| 🔵 | nit | Style, naming, micro-perf — emit only when thorough review requested |
| ❓ | question | Need author intent before judging |

## Finding line

```
<path>:<line>: <emoji> <tier>: <problem>. <fix>.
```

One line per finding. Problem stated first, then the concrete fix — both tight. Order findings by
severity descending (🔴 → 🟡 → 🔵 → ❓) so critical issues surface first; within a tier, file order,
ascending line numbers.

For QA frontend defects with no clean code line, use `<view/route>: <step>:` as the locator and
attach the screenshot path.

## Totals + summary

Close every report with a totals line, then a one-sentence verdict:

```
<path>:<line>: 🔴 bug: token expiry uses `<` not `<=`. Off-by-one allows expired tokens 1 tick.
<path>:<line>: 🟡 risk: pool not closed on error path. Add `try/finally`.
<path>:<line>: ❓ question: why duplicate `.trim()` here? Need author intent.
totals: 1🔴 1🟡 1❓
summary: One real bug plus an error-path leak; small, focused fix set.
```

Zero findings → `No issues.` (no totals line, no summary).

## Rules

- Report only what is in the diff/files/tests in front of you. No "while we're here" scope creep.
- No big-refactor proposals inside the report — if one is unavoidable, raise it as a single ❓
  question, not a finding.
- Need more context to judge → append `(see L<n> in <file>)`. Never guess.
- Skip formatting nits unless they change meaning.
- No praise preamble. Findings only.

## Size

- One line per finding, no exceptions — no multi-line findings, no nested detail blocks.
- Past roughly 40 findings: totals and the severity breakdown in chat, the full finding list written
  to a file path.
- Screenshots as paths, never inline images.
- Quoted command output capped at ~10 lines; the full log is a path.

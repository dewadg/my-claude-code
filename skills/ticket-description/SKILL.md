---
name: ticket-description
description: Write the description body of a ticket/issue in a fixed section structure (Context/Background, Out of Scope, Pre-conditions, Requirements, Errors & Edge Cases, Acceptance Criteria with Given/When/Then), on any tracker — Jira, GitLab issues, GitHub issues, Linear, or no tracker at all. Use this whenever the user wants a ticket or issue description written, filled in, restructured, or fleshed out — including "fill USD-4300", "write the description for issue #42", "turn these notes into a story", "write the ACs", "this ticket is empty, flesh it out", "make this issue ready for dev", or when they paste rough requirements, a Slack thread, or a Figma link and expect a ticket out of it. Also use when reformatting an existing description that doesn't follow the structure. Produces the description text only; other fields (title, labels, points, assignee) are out of scope.
---

# Ticket description

You are writing the text a developer reads when the ticket lands in their sprint and nobody is around to explain it. They don't know the meeting it came from, the Figma frame, or which of the five similar screens this touches. The description has to survive that.

Two things follow from that. First, structure is fixed — the team scans for the same headings every time, so a heading in an unexpected place costs more than it adds. Second, **never invent requirements**. A fabricated acceptance criterion gets built. If you don't know something, ask or mark it, but do not fill the gap with a plausible-sounding bullet.

The structure below is tracker-agnostic. Only *how you read and write the ticket* changes per system.

## 1. Identify the tracker

Work out which system the ticket lives in before fetching anything. In order of evidence:

1. **An explicit URL** the user pasted — the host names the tracker.
2. **The identifier shape** — `PROJ-123` → Jira/Linear; `#42` or `group/project#42` → GitLab; `#42` in a GitHub repo → GitHub; `ENG-123` → Linear.
3. **The repo's remote** — a GitLab remote makes bare `#42` a GitLab issue, a GitHub remote makes it a GitHub issue.
4. **What's connected** — check which issue-tracker MCP servers or CLIs are actually available; that usually settles it.

If it's still ambiguous, ask which tracker rather than guessing — fetching the wrong `#42` wastes a round trip and writing to it is worse.

| Tracker | Read | Write description |
|---|---|---|
| Jira (MCP) | `jira_get_issue(issue_key=…, fields="summary,description,issuetype,labels,parent,status")` | `jira_update_issue(issue_key=…, fields={"description": "…"})` |
| GitLab (MCP) | `get_issue(project_id=…, issue_iid=…)` | `update_issue(project_id=…, issue_iid=…, description="…")` |
| GitLab (CLI) | `glab issue view 42` | `glab issue update 42 --description "…"` |
| GitHub (CLI) | `gh issue view 42 --json title,body,labels,state` | `gh issue edit 42 --body "…"` |
| Linear (MCP) | `get_issue` | `update_issue` |
| None | the user's pasted text is the whole input | print only; nothing to write back |

Tool names vary between MCP server builds — list what's connected and use the equivalent rather than calling a name from this table that doesn't exist.

## 2. Gather

**Given a ticket identifier** — fetch it before writing anything. The title tells you the intent, the existing description (however rough) tells you what's already known, and the parent link gives you the wider container: an epic or parent issue in Jira/Linear, an epic or milestone in GitLab, a tracking issue or project in GitHub. Read it too when the ticket's scope only makes sense inside it. If sibling tickets are referenced ("step 1.4", "as in USD-4289", "blocked by #38"), fetch those as well rather than guessing what they say.

Work type matters more than what the tracker calls it. Jira gives you an issue type field; GitLab and GitHub give you labels (`bug`, `feature`, `chore`) and sometimes nothing at all. Infer story vs bug vs task from the type field, then the labels, then the title — and when nothing says otherwise, write a story.

**Given raw notes / a pasted thread** — that text is your only source of truth. Everything in the output must trace back to a line in it, to a ticket you fetched, or to an answer the user gave you.

**Given a Figma link** — put it under Requirements → Design links. Fetch it only if you actually need to describe on-screen behaviour that the notes don't cover; a link the reader can click is usually enough, and the ticket should point at the design rather than transcribe it.

## 3. Close the gaps before writing

Read what you have and list what's genuinely missing — most often: which screens/entities are affected, what happens in the empty or single-item case, whether existing behaviour changes, and what "done" looks like for the edge cases.

Ask the user those questions in one batch, then write. Asking three questions once beats delivering a ticket padded with invented answers. But keep it proportional: if the notes are detailed and only one small thing is unclear, write the ticket and flag that one thing inline as `> TODO: confirm …` rather than blocking on it.

## 4. Write it

Pick the shape by work type. Story is the default; the other two are the same spirit with a different middle.

### Story

````markdown
# Context/Background

[2–4 sentences: where in the product this lives, who the actor is, what this story adds and why they need it. Bold the feature's proper name on first use.]

# Out of Scope

- [Things a reader would reasonably assume are included, but aren't — and who owns them if someone else does.]

# Pre-conditions

- [What must already exist for this to be buildable/testable.]

# Requirements

**Design links**

1. [Screen name](figma-url)

**[Sub-group, e.g. the component being added]**

- [Behaviour statement.]
- [Behaviour statement.]

**[Next sub-group, e.g. the modal]**

- Clicking the chip opens a modal:
    - Title: "[exact copy]"
    - **[Button label]** (primary) — [what it does]
    - **[Button label]** (secondary) — [what it does]

**States**

- [State: when visible/enabled vs not. Defer exact interaction detail to the design link.]

# Errors & Edge Cases

|Scenario|Expected Behavior|
|---|---|
|[Condition]|[What the system does]|

# Acceptance Criteria

## [Group heading when the ACs split across screens/flows]

**AC 1: [One-line title stating the outcome]**

  - **Given** [starting state]
      - **And** [additional condition]
  - **When** [the actor's action]
  - **Then** [observable result]
      - **And** [additional result]
````

### Bug

Same outer sections, but the middle becomes evidence instead of requirements:

```markdown
# Context/Background
# Steps to Reproduce      ← numbered, from a clean state, with the actual data used
# Expected vs Actual      ← two labelled paragraphs, or a two-column table
# Scope/Impact            ← environments affected, which users, how often, since when
# Errors & Edge Cases
# Acceptance Criteria     ← at minimum one AC asserting the fixed behaviour
```

Keep the reproduction literal — real entity names, real values, real URLs from the report. A dev who can't reproduce it can't fix it, and a sanitised "user does something" step is how tickets bounce back.

### Task / Chore

Strip it down; ceremony on a config change wastes everyone's time:

```markdown
# Context/Background
# Out of Scope            ← only if there's a real boundary worth stating
# Requirements
# Acceptance Criteria     ← usually 1–3 ACs, or a plain "Done when" bullet list if Given/When/Then reads absurd
```

## What good looks like in each section

**Context/Background** — the "why" the rest of the ticket can't carry. Name the screens, the actor ("admins", "recipients"), and the user-facing problem. A dev should finish this section knowing whether they're the right person to pick it up.

**Out of Scope** — this section earns its keep by preventing scope creep in review, so it should list the things a reader *would* assume are in. "Historical surveys", "the existing X callout (unchanged)", "dashboard rendering (Dashboard squad)". Omit the heading entirely rather than writing "N/A".

**Requirements** — statements about behaviour, not implementation. Say what appears, when, and what it says; leave component names, hooks, and API shapes to the dev unless a specific technical constraint is genuinely part of the ask. Quote user-facing copy exactly and in quotes — copy that gets paraphrased here gets paraphrased into the product.

Group with bold sub-headings when there's more than one moving part. Nested bullets carry the detail of the thing above them.

**Errors & Edge Cases** — a table, one row per scenario. This is where reviewers look first, so cover the boring ones: empty state, single item, English-only entity, permission denied, expired session. If you couldn't determine the expected behaviour for an edge case, put the row in with a `TODO: confirm` — a visible unknown is useful, a silently dropped row is not.

**Acceptance Criteria** — one AC per testable outcome, each with a title stating the outcome so the list reads as a checklist on its own. Then Given / When / Then, with `**And**` for extra conditions or results, indented under the clause it extends.

Write them from the outside: what a QA engineer can observe on screen or in a response. "Then the chip is shown" is testable; "Then the state updates correctly" is not. Cover both directions of every conditional — if AC 1 says the chip appears when translations are missing, there must be an AC saying it's hidden when they're complete. Missing negative cases are the most common defect in these tickets.

When the story spans two screens or flows, group the ACs under `##` headings and repeat the set per flow rather than writing "same as AC 1 but for X" — QA runs them one at a time.

## Formatting

Write the description in markdown: `#` for section headings, `**bold**` for AC titles, Given/When/Then keywords, and sub-group labels within Requirements, `##` for AC group headings, and pipe tables for edge cases. Never use italic for a heading, sub-group label, or bullet label — bold everywhere. Keep the section names and their order exactly as above — that fixed vocabulary is the whole point.

Markdown lands as-is on GitLab, GitHub, and Linear. Jira is the exception: older instances render wiki markup, and some Cloud APIs want ADF. Write markdown either way, and if the ticket you fetched came back in wiki markup or ADF, convert on the way in (so you can read it) and match the field's existing format on the way out (so you don't mangle it).

## 5. Deliver

Print the finished description in a fenced code block so it can be copied verbatim. Then, if you're working on a real ticket, offer to write it:

> Write this to USD-4300? (replaces the current description)

Only on an explicit yes, update it — and only the description field, using the write call for that tracker from the table in step 1.

Never write without that confirmation. Descriptions are frequently hand-edited by PMs, and overwriting one silently destroys work that isn't in your context. If the ticket already had a substantial description, say what you're replacing before asking. Where the tracker supports it, prefer an update that touches only the description over one that rewrites the whole issue.

## Before you print

Reread it as the developer who got assigned it on a Monday: is there anything here you'd have to ask about in standup? Then reread as QA: can you execute every AC without a conversation? Then check the honest question — is every statement traceable to something you were actually told, or did one of them come from you?

A worked example of a real story in this format is in `references/example-story.md`. Read it when you want to see the level of detail and phrasing the team expects; skip it if the shape above is already clear.

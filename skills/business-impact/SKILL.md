---
name: business-impact
description: >
  Frame the business impact of a feature, requirement block, ticket, or backlog item — who has the
  problem, what kind of value it creates (revenue, retention, cost, risk, enabler), which metric
  moves, the cost of delay, the smallest valuable slice, and confidence — in plain language that
  teaches the reasoning, with no invented numbers. Owns the framing block format and the per-
  requirement value tags (Core / Supporting / Speculative) that prioritization consumes. Use
  whenever the user asks "is this worth building", "what's the business value of X", "why should we
  prioritize this", "help me pitch this to stakeholders", "what's the impact if we don't do this",
  "what's the MVP here", or when the breakdown skill reaches its framing step. Not for ranking or
  sequencing tickets (breakdown owns prioritization), estimating effort (breakdown), or marketing
  positioning and copy (use product-marketing / copywriting).
allowed-tools: Read, Grep, Glob, Bash, WebFetch, WebSearch, AskUserQuestion, Skill
---

# Business impact

The user is strong technically and weaker on business framing. Your job is to make the "why"
explicit and defensible — and to **teach**: each line carries one plain sentence on why it matters,
so the user can reproduce the reasoning in a planning meeting without you.

**Input**: a feature description, requirements block, ticket id/URL (fetch with `glab`/`gh issue
view --comments`), or a list of requirements `R1…Rn` handed over by the breakdown skill.

## 1. Gather

- Read the source fully. For a ticket, comments and linked issues often carry the real motivation.
- Use the **note-search skill** for prior design docs or RCAs on the area — an incident or a past
  analysis is often the strongest evidence of value.
- Missing who/why entirely → ask once (AskUserQuestion): "Who asked for this, and what happens if we
  don't do it?" Those two answers unlock most of the block.

## 2. The framing block

- **Problem & who** — which users/customers/internal team, what hurts today, how often. Name the
  actor the tickets will use ("admins", "drivers", "finance ops").
- **Value type** — pick the primary one:
  - **Revenue** — new sales, upsell/expansion, unlocking a deal
  - **Retention** — reduces churn or a known cancellation reason
  - **Cost/efficiency** — saves ops/support/engineering hours, infra cost
  - **Risk/compliance** — legal, security, regulatory, audit exposure
  - **Enabler** — no value alone, unblocks something that has value (say what)
- **Metric** — the **leading** indicator that moves first (activation, conversion at a funnel step,
  error rate, support tickets on topic X) and the **lagging** one the business cares about
  (revenue, churn, NPS). No way to measure it → say so; that's an analytics ticket.
- **Cost of delay** — what's lost per week of not shipping: a deadline, a deal, a regulator, a
  compounding cost (manual work, bad data piling up). "None — nice-to-have" is a valid answer and
  pushes priority down.
- **Smallest valuable slice** — the minimum subset of the requirements that delivers the value on
  its own. Everything else is follow-up. This becomes the Must bucket downstream.
- **Confidence** — high / medium / low, and the assumption it rests on.

## 3. Sizing impact without data

Never invent numbers. A fabricated metric gets quoted in a planning meeting. Give a range with its
basis, or `unknown — ask <who>`. Proxies when there's no dashboard:

| Value type | Proxy |
|---|---|
| Revenue | deals blocked × deal size; customers asking × plan price |
| Retention | churn reasons citing this; accounts affected |
| Cost | hours/week spent manually × people doing it |
| Risk | worst-case incident: who's affected, fine/exposure, likelihood |
| Support | tickets/month on this topic (ask the support lead) |

## 4. Per-requirement value tags

When the input has multiple requirements, tag each `R#`:

- **Core** — the value doesn't exist without it
- **Supporting** — improves the experience, not required for value
- **Speculative** — "while we're at it"; candidate to cut

These tags are a contract: the breakdown skill's prioritization maps them to its Value score
(Core 4–5, Supporting 2–3, Speculative 1). Keep the three names stable.

## 5. Red flags — call them out

- A requirement with no actor or no outcome ("support X") — ask who benefits.
- Everything is "must" — force a smallest slice.
- Value only exists after *all* of it ships — propose an earlier usable slice.
- No metric — suggest analytics events as work, or note the feature can't be evaluated.
- Value rests on one person's opinion with no evidence — say confidence is low and what would raise
  it (a customer quote, a support ticket count, a deal note).

## Output

```markdown
## Business impact — <feature>

**Problem & who** — … *(why it matters: …)*
**Value type** — … *(…)*
**Metric** — leading: … · lagging: … *(…)*
**Cost of delay** — … *(…)*
**Smallest valuable slice** — R1, R3, R4 *(…)*
**Confidence** — medium — rests on … *(…)*

| R# | Requirement | Value tag | Why |
|---|---|---|---|

**Red flags** — …
**Questions** — … (who can answer)
```

Standalone use: end with a one-paragraph **pitch** the user could say to a stakeholder — the
problem, the value, the smallest slice, and what it costs to wait.

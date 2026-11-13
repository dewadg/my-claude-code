# Prioritization & sequencing

Rank the tickets (or features, or backlog items) and **show the reasoning** — the user needs to
defend the order to stakeholders.

## 1. Hard constraints first — override scoring

- Production incidents, security issues
- External deadlines: contract, regulation, launch/event date, app store cutoffs
- Dependencies: a ticket can't start before its `blocked by` finishes
- Spikes for unknowns that size or shape other tickets

## 2. Score the rest — WSJF-lite

Each 1–5:

| Factor | Ask |
|---|---|
| **Value** | How much of the business value does this carry? From the business-impact skill's tags: Core = 4–5, Supporting = 2–3, Speculative = 1 |
| **Urgency** | What does a week of delay cost? Deadline/deal = 5, none = 1 |
| **Risk/Enablement** | Does it retire a risk or unblock other tickets? Contract/groundwork/spike = high |
| **Effort** | XS=1, S=2, M=3, L=5 |

**Score = (Value + Urgency + Risk/Enablement) ÷ Effort.** Show the four numbers, not just the
result. Scores support a conversation; they aren't a verdict — say when the ranking flips on one
judgment call.

## 3. Bucket

- **Must** — the smallest valuable slice from business framing; release isn't useful without it
- **Should** — important, release works without it
- **Could** — cut first when capacity runs short
- **Won't (this time)** — explicitly out, with reason

Roadmap view, when the block spans releases: **Now / Next / Later**.

## 4. Sequence

- Respect `blocked by`; draw lanes per seam:
  ```
  BE     : #1 spike → #2 migration → #3 POST /refunds → #5 GET /refunds
  FE     :                     #4 Refund modal (mock on #3) → #6 Order detail
  Mobile :                     #7 Refund screen (mock on #3)
  ```
- **Critical path** — the longest chain of `blocked by` by summed effort. Name it; it sets the
  earliest finish. Shorten it by moving contract definition earlier or splitting the long ticket.
- Spikes and risky unknowns first; thin end-to-end slice before breadth; polish last.
- Balance lanes — if FE has 3× BE's effort, say so; it's a staffing question for the user.

## 5. Cut line

When capacity is known (people × weeks), draw the line on the ranked list and name what drops
first. When it isn't, show where the Must bucket ends and its effort total.

## Judgment calls — hand back to the user

Flag as **the user's decision to take to stakeholders**, with the trade-off spelled out, when:

- Two Musts compete for the same capacity
- Value depends on business context you don't have (deal status, strategy shift)
- Shipping a partial slice changes what users see (half a feature visible vs flag-hidden)

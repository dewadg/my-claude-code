# Grooming (`--groom`)

Target is an existing backlog — a label, milestone, epic, or list of ticket ids — not one block to
split. Fetch it via `tracker-cli.md` (JSON list with title, labels, body, updated date).

## 1. Hygiene

- **Duplicates** — same outcome in two tickets → propose merge, name the survivor
- **Stale** — untouched > 90 days → archive or decide; don't keep zombies
- **Stuck** — "ready" but skipped for 3+ sprints → ask why (wrong priority, hidden blocker, too big)
- **Drifted priority** — priority set for last quarter's goals → re-score

## 2. Refine the top candidates (top 5–10 by current priority)

For each, check against the **Definition of Ready**:

- [ ] Acceptance criteria testable (Given/When/Then, both directions of every conditional)
- [ ] Single seam (BE / FE / mobile / …) — crosses seams → split per `seams.md`
- [ ] Sized ≤ L in human effort (`estimation.md`) — XL → split
- [ ] Design linked (if UI) and final
- [ ] Dependencies known and scheduled or resolved
- [ ] Business value statable in one line

Failing tickets: propose the fix — split plan, missing ACs (rewrite via ticket-description on
approval), questions for the owner.

Too vague to refine → it's a big block; run the main breakdown workflow on it.

## 3. Re-prioritize

Tickets with no statable value → run the **business-impact skill** on them (batch, short form).
Then score and bucket per `prioritization.md`.

## 4. Categorize

- **Now** — ready, ranked
- **Next** — rough intent, next 2–3 sprints, needs refinement
- **Later** — directional
- **Icebox** — parked, revisit quarterly
- **Won't Do** — rejected, reason noted (close with a comment, on approval)

## Output

- Ready vs needs-work table: `# | Title | Seam | Size | DoR gaps | Proposed action`
- Merge / archive / split proposals
- Re-ranked Now list with scores
- Questions per ticket owner
- Agenda for the next refinement session (what still needs a human conversation)

No tracker changes (labels, closes, edits) without explicit approval per action group.

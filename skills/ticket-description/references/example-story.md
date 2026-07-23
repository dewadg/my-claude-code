# Worked example — Story

A real story ("Missing dimensions translation callout and modal on comms screens"), lightly cleaned. It came off a Jira board, but nothing in the body is tracker-specific — the same text is what you'd paste into a GitLab or GitHub issue. Note the proportions: Context is short, Requirements carry the exact copy, and every conditional in Requirements has both a positive and a negative AC.

---

# Context / Background

On the Dashboard access and Action suggestions comms screens, admins schedule emails to users. This story adds a **Missing dimensions translation** callout chip and a modal that lets the admin jump to add translations or proceed as-is, so they're aware some dimensions aren't translated before sending.

# Out of Scope

- Historical surveys
- The existing "Missing comms translation" callout and the comms language preview dropdown (existing behaviour)
- Dashboard rendering and fallback (Dashboard squad)

# Pre-conditions

- The Dashboard access and Action suggestions comms screens exist
- The Dimensions translations tab exists, so completeness can be evaluated

# Requirements

**Design links**

1. [Dashboard access — comms screen](https://www.figma.com)
2. [Action suggestions — comms screen](https://www.figma.com)
3. [Missing dimensions translation modal](https://www.figma.com)

**Callout chip**

- On the Dashboard access and Action suggestions comms screens, a **Missing dimensions translation** callout chip appears when one or more dimension translations are missing for the entity's languages.
- It appears alongside the existing **Missing comms translation** callout.
- The chip is informational only — admins can proceed without completing translations.

**Modal**

- Clicking the chip opens a modal:
    - Title: "Some dimensions aren't translated yet."
    - Body: "Users viewing the dashboard in those languages will see these dimensions in English until translations are added."
    - **Add translations** (primary) — navigates to the Dimensions translations tab (step 1.4).
    - **Keep current settings** (secondary) — closes the modal and stays on the comms screen.
    - Close (X) — closes the modal with no change.

**States**

- Chip: visible (translations missing), hidden (all complete). Modal: open, closed. For the exact interaction behaviour, refer to the [Figma design](https://www.figma.com).

# Errors & Edge Cases

|Scenario|Expected Behavior|
|---|---|
|Entity is English-only|No other languages exist, so no chip is shown.|

# Acceptance Criteria

## Dashboard access comms screen

**AC 1: Chip appears on Dashboard access when dimension translations are missing**

  - **Given** the admin is on the Dashboard access comms screen
      - **And** one or more dimension translations are missing for the entity's languages
  - **When** the admin views the screen
  - **Then** a "Missing dimensions translation" callout chip is shown

**AC 2: Chip is hidden on Dashboard access when all translations are complete**

  - **Given** all dimension translations are complete
  - **When** the admin views the Dashboard access comms screen
  - **Then** the "Missing dimensions translation" chip is not shown

**AC 3: Modal opens from the Dashboard access comms screen**

  - **Given** the "Missing dimensions translation" chip is visible on the Dashboard access comms screen
  - **When** the admin clicks it
  - **Then** a modal opens titled "Some dimensions aren't translated yet." with the English-fallback message
      - **And** "Add translations" navigates to the Dimensions translations tab (step 1.4)
      - **And** "Keep current settings" or closing the modal dismisses it with no change, keeping the admin on the Dashboard access screen

## Action suggestions comms screen

**AC 4: Chip appears on Action suggestions when dimension translations are missing**

  - **Given** the admin is on the Action suggestions comms screen
      - **And** one or more dimension translations are missing for the entity's languages
  - **When** the admin views the screen
  - **Then** a "Missing dimensions translation" callout chip is shown

**AC 5: Chip is hidden on Action suggestions when all translations are complete**

  - **Given** all dimension translations are complete
  - **When** the admin views the Action suggestions comms screen
  - **Then** the "Missing dimensions translation" chip is not shown

**AC 6: Modal opens from the Action suggestions comms screen**

  - **Given** the "Missing dimensions translation" chip is visible on the Action suggestions comms screen
  - **When** the admin clicks it
  - **Then** a modal opens titled "Some dimensions aren't translated yet." with the English-fallback message
      - **And** "Add translations" navigates to the Dimensions translations tab (step 1.4)
      - **And** "Keep current settings" or closing the modal dismisses it with no change, keeping the admin on the Action suggestions screen

---

## Why this one works

- **Context names the actor and the screens.** A dev knows in two sentences whether it's their area.
- **Out of Scope pre-empts the three things a reader would assume are included** — including one owned by another squad.
- **Copy is quoted exactly.** Button labels, modal title, body text. Nothing for the dev to invent.
- **Every conditional has both directions.** AC 1/2 and AC 4/5 are positive/negative pairs.
- **ACs are repeated per screen rather than cross-referenced.** QA runs one at a time; "same as AC 1 but on the other screen" is where coverage quietly drops.
- **Interaction detail is deferred to Figma** rather than transcribed, so the ticket doesn't drift from the design.

---
name: telegram-notify
description: >
  Send the user a Telegram message when a condition they asked about resolves — build finished,
  deploy done, CI green or failed. Use for "let me know when X finishes", "ping me when the tests
  pass". Reads TELEGRAM_BOT_TOKEN / TELEGRAM_CHAT_ID and sends MarkdownV2 via the bundled script.
---

# Telegram notify

Send the user a Telegram message when a criterion they specified is met. The
point is an out-of-band heads-up: the user has stepped away and wants to know an
outcome without watching the terminal.

## When to fire

The user describes a future condition plus a wish to be told about it:
"let me know when X", "ping me once Y finishes", "notify me if Z fails". Do the
work they asked for, watch for the condition, and **when it resolves**, send one
message reporting the outcome. Send on both success and failure unless they
scoped it to one (e.g. "only if it breaks").

If the trigger is a long-running command, run it to completion first, then send.
Don't send a "starting now" message unless asked — they want the result.

## How to send

Run the bundled script with the message as the first argument. A skill's paths resolve relative to
its own base directory, so the script is `<skill-dir>/scripts/send_telegram.sh`
(`~/.claude/skills/telegram-notify` when installed globally):

```bash
<skill-dir>/scripts/send_telegram.sh "$MESSAGE"
```

When the content is a single link, add a `--copy` flag with the raw URL so the
message carries a one-tap copy-to-clipboard button (see "Single link" below):

```bash
<skill-dir>/scripts/send_telegram.sh "$MESSAGE" --copy "$URL"
```

The script handles transport: it reads `TELEGRAM_BOT_TOKEN` and
`TELEGRAM_CHAT_ID` from the environment, posts with `parse_mode=MarkdownV2`, and
URL-encodes the body so newlines and special characters survive. The `--copy`
value is raw text (not MarkdownV2) — don't escape it.

To attach an image, add `--photo` with a URL, a local file path, or a Telegram
`file_id` (see "Photo / image" below):

```bash
<skill-dir>/scripts/send_telegram.sh "$CAPTION" --photo "$IMAGE"
```

To preview a message without sending, set `TELEGRAM_DRY_RUN=1`; the script prints
the assembled text and button payload and exits without needing credentials.

**Credentials are best-effort.** If either env var is unset, the script skips
silently and exits 0 — never treat a missing token as a task failure, and don't
nag the user to configure it unless they ask why no message arrived.

**Tmux session footer is automatic.** When the session runs inside tmux, the
script appends a `tmux: ` line with the session name so the user can tell which
session pinged them. Don't add this yourself — write only the outcome and
context, and let the script handle the footer.

**An exit code of 2 means the Telegram API rejected the message** — almost
always a MarkdownV2 escaping mistake. Read the error printed to stderr, fix the
escaping (see below), and retry once.

## Message structure

Every message has the same shape so it scans on a phone:

```
<bold title with emoji>

<content>

tmux: <session>
```

- **Title** — one line, `*bold*`, leading emoji for the outcome (✅ success, ❌ failure, ⚠️ warning,
  🚀 deploy).
- **Blank line** after it — Telegram collapses a single newline, so without it the headline runs
  into the body.
- **Content** — one or two lines that make the outcome actionable: a number, a path, a link. The
  reader is away from the screen, so a bare "done" forces a trip back to check.
- **`tmux:` footer** — the script appends it inside tmux; never write it yourself.

A notification, not a report: no log dumps, no preamble.

**Example — success:**
```
*✅ Tests passed*

142 passed, 0 failed in `auth` and `chat` packages\.
```

**Example — failure:**
```
*❌ Deploy failed*

`go build` errored: undefined `UsageSink` in `internal/llm/module\.go`\.
```

## Single link

When the content's payload is essentially one link (a deploy preview, a PR, a
report URL), make it effortless to act on:

1. **Render the link clickable in the body** with an inline link whose visible
   text is the URL itself: `[https://example\.com/x](https://example.com/x)`. The
   bracketed display half is MarkdownV2 — escape its `.` `-` etc. The parenthesized
   URL half is *not* parsed as markup, so leave it raw (only a literal `)` or `\`
   inside it needs escaping). Showing the URL as the link text keeps it readable
   even though it's tappable.
2. **Attach a copy button** by passing `--copy "<raw-url>"` to the script. This
   adds a 📋 Copy link button that drops the URL straight onto the clipboard on
   tap — handy when the user wants to paste it elsewhere rather than open it.

Only do this for a *single* link. If the content has several links or none, skip
the button and just format the prose normally.

**Example — single link:**
```
*🚀 Preview ready*

PR \#42 deployed: [https://app\.example\.com/pr/42](https://app.example.com/pr/42)
```
sent with `--copy "https://app.example.com/pr/42"`.

## Photo / image

When the outcome is best shown as an image — a screenshot, a chart, a generated
preview, a QR code — send it with `--photo`. The text argument becomes the
image's **caption**, so write it exactly like a normal message title + content;
the same MarkdownV2 escaping rules apply.

```bash
<skill-dir>/scripts/send_telegram.sh "$CAPTION" --photo "<url | local-path | file_id>"
```

The `--photo` value is one of:
- an **http(s) URL** — Telegram fetches and re-hosts it;
- a **local file path** — uploaded as multipart (`-F photo=@path`). The path is
  detected by existing on disk;
- a **Telegram `file_id`** — reuses an already-uploaded photo.

Notes:
- **Captions cap at 1024 characters** (a plain message allows 4096). Keep the
  caption tight; if you have a lot to say, send the photo with a short caption
  and the detail as a follow-up text message.
- `--copy` works alongside `--photo` — the copy button attaches to the photo
  message.
- A caption is required (it's the first argument); for a near-bare image use a
  one-line title like `*🖼 Screenshot*`.

**Example — screenshot of a passing run:**
```bash
<skill-dir>/scripts/send_telegram.sh "*✅ E2E green*

widget welcome \+ buttons verified" --photo "/tmp/welcome-e2e.png"
```

## MarkdownV2 escaping (read this — it's the one thing that breaks)

With `parse_mode=MarkdownV2`, Telegram rejects any *literal* unescaped instance of:

```
_ * [ ] ( ) ~ ` > # + - = | { } . !
```

- Use `*bold*` for the headline and backtick spans for values, paths, commands, identifiers. Inside
  a code span only a literal backtick or backslash needs escaping — which is why wrapping
  `internal/llm/module.go` beats escaping its dots.
- In prose, escape the punctuation: `.` `!` `-` `(` `)` are the common offenders (see the `\.` and
  `\!` in the examples above). Versions (`v1\.2\.0`) and decimals need it too.
- When in doubt, put the volatile token in backticks and escape the surrounding prose. Exit code 2
  means the API rejected the message; the stderr error names the byte offset — fix that character
  and resend once.

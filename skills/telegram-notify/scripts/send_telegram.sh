#!/usr/bin/env bash
# Send a Telegram message via the Bot API.
#
# Usage:  send_telegram.sh "<text>" [--copy "<text-to-copy>"] [--photo "<url|path|file_id>"]
#   - text is sent with parse_mode=MarkdownV2, so it MUST already have
#     MarkdownV2 reserved characters escaped (see the skill's SKILL.md).
#     With --photo, the text becomes the photo CAPTION (Telegram caps
#     captions at 1024 characters, vs 4096 for a plain message).
#   - --copy "<str>" attaches an inline-keyboard button that copies <str>
#     to the user's clipboard when tapped (Telegram CopyTextButton, Bot
#     API 7.11+). Used to make a single link one-tap copyable. <str> is
#     raw text, NOT MarkdownV2 — do not escape it.
#   - --photo "<url|path|file_id>" sends an image via sendPhoto with the
#     text as its caption. The value may be an http(s) URL, a local file
#     path (uploaded as multipart), or a Telegram file_id. Local paths are
#     detected by existing on disk; anything else is passed to Telegram as-is.
#   - Reads TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID from the environment.
#   - If either env var is unset/empty, the send is skipped silently
#     (exit 0) — notifications are best-effort and must never break the
#     task that triggered them.
#   - If TELEGRAM_DRY_RUN is set, the assembled payload (text + any
#     reply_markup) is printed to stdout and nothing is sent (exit 0).
#     Useful for previewing structure without a live chat.
#
# Exit codes:
#   0  sent, or skipped because env is not configured, or dry-run
#   1  bad usage (no text argument)
#   2  Telegram API returned ok=false (e.g. a MarkdownV2 escaping error) —
#      the API's description is printed so the caller can fix and retry.

set -euo pipefail

text="${1-}"
if [ -z "$text" ]; then
  echo "send_telegram.sh: missing text argument" >&2
  exit 1
fi
shift || true

copy_text=""
photo=""
while [ $# -gt 0 ]; do
  case "$1" in
    --copy)
      copy_text="${2-}"
      shift 2 || { echo "send_telegram.sh: --copy needs a value" >&2; exit 1; }
      ;;
    --photo)
      photo="${2-}"
      shift 2 || { echo "send_telegram.sh: --photo needs a value" >&2; exit 1; }
      [ -z "$photo" ] && { echo "send_telegram.sh: --photo value is empty" >&2; exit 1; }
      ;;
    *)
      echo "send_telegram.sh: unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

# When running inside tmux, append the session name so the user knows which
# session the notification came from. Wrapped in a code span, so only a literal
# backtick or backslash in the name needs MarkdownV2 escaping.
if [ -n "${TMUX:-}" ] && command -v tmux >/dev/null 2>&1; then
  session="$(tmux display-message -p '#S' 2>/dev/null || true)"
  if [ -n "$session" ]; then
    session="${session//\\/\\\\}"   # escape backslash first
    session="${session//\`/\\\`}"   # then backtick
    text="${text}"$'\n\n'"tmux: \`${session}\`"
  fi
fi

# Build a copy-to-clipboard inline keyboard when a --copy value was given.
# JSON string rules: escape backslash, double-quote, and control chars. A URL
# rarely contains any of these, but escape defensively so the JSON stays valid.
reply_markup=""
if [ -n "$copy_text" ]; then
  esc="${copy_text//\\/\\\\}"   # backslash first
  esc="${esc//\"/\\\"}"          # then double-quote
  esc="${esc//$'\n'/\\n}"        # newlines, just in case
  esc="${esc//$'\t'/\\t}"
  reply_markup="{\"inline_keyboard\":[[{\"text\":\"📋 Copy link\",\"copy_text\":{\"text\":\"${esc}\"}}]]}"
fi

# Pick the API method: a photo turns this into a sendPhoto call whose caption is
# the (already MarkdownV2-escaped) text; otherwise a plain sendMessage.
if [ -n "$photo" ]; then
  method="sendPhoto"
  text_field="caption"
else
  method="sendMessage"
  text_field="text"
fi

# Dry-run: show what would be sent and stop. No network, no credentials needed.
if [ -n "${TELEGRAM_DRY_RUN:-}" ]; then
  echo "--- method ---"
  printf '%s\n' "$method"
  if [ -n "$photo" ]; then
    echo "--- photo ---"
    if [ -f "$photo" ]; then printf 'upload: %s\n' "$photo"; else printf '%s\n' "$photo"; fi
  fi
  echo "--- ${text_field} ---"
  printf '%s\n' "$text"
  if [ -n "$reply_markup" ]; then
    echo "--- reply_markup ---"
    printf '%s\n' "$reply_markup"
  fi
  exit 0
fi

# Best-effort: if not configured, do nothing and succeed.
if [ -z "${TELEGRAM_BOT_TOKEN:-}" ] || [ -z "${TELEGRAM_CHAT_ID:-}" ]; then
  echo "send_telegram.sh: TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID not set, skipping" >&2
  exit 0
fi

endpoint="https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/${method}"

if [ -n "$photo" ] && [ -f "$photo" ]; then
  # Local file: multipart upload. --form-string sends field values literally
  # (no @/< filename interpretation), so a caption starting with '*' or '@'
  # and any newlines survive intact; only the actual file uses -F @path.
  args=(-s -X POST "$endpoint"
    --form-string "chat_id=${TELEGRAM_CHAT_ID}"
    -F "photo=@${photo}"
    --form-string "${text_field}=${text}"
    --form-string "parse_mode=MarkdownV2")
  if [ -n "$reply_markup" ]; then
    args+=(--form-string "reply_markup=${reply_markup}")
  fi
else
  # Plain message, or a photo passed by URL / file_id (Telegram fetches it).
  # --data-urlencode (not plain -d) so newlines and reserved URL characters
  # like & and = inside the message survive transport intact.
  args=(-s -X POST "$endpoint"
    --data-urlencode "chat_id=${TELEGRAM_CHAT_ID}"
    --data-urlencode "${text_field}=${text}"
    -d "parse_mode=MarkdownV2")
  if [ -n "$photo" ]; then
    args+=(--data-urlencode "photo=${photo}")
  fi
  if [ -n "$reply_markup" ]; then
    args+=(--data-urlencode "reply_markup=${reply_markup}")
  fi
fi

response="$(curl "${args[@]}")"

# Surface API failures so a MarkdownV2 escaping mistake is fixable, not silent.
case "$response" in
  *'"ok":true'*)
    exit 0
    ;;
  *)
    echo "send_telegram.sh: Telegram API error:" >&2
    echo "$response" >&2
    exit 2
    ;;
esac

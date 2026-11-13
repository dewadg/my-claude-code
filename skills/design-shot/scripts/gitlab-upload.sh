#!/usr/bin/env bash
# Upload an image to a GitLab project's uploads and print the markdown to embed in an issue.
#   gitlab-upload.sh <project-path-or-id> <file> [host]
# Token: $GITLAB_PAT, else $GITLAB_TOKEN, else `glab config get token --host <host>`.
# Host: arg 3, else $GITLAB_HOST, else gitlab.com.
set -euo pipefail

project="${1:?usage: gitlab-upload.sh <project-path-or-id> <file> [host]}"
file="${2:?usage: gitlab-upload.sh <project-path-or-id> <file> [host]}"
host="${3:-${GITLAB_HOST:-gitlab.com}}"
host="${host#https://}"; host="${host%/}"

[ -f "$file" ] || { echo "no such file: $file" >&2; exit 2; }

token="${GITLAB_PAT:-${GITLAB_TOKEN:-}}"
[ -n "$token" ] || token="$(glab config get token --host "$host" 2>/dev/null || true)"
[ -n "$token" ] || { echo "no GitLab token: set GITLAB_PAT or run 'glab auth login'" >&2; exit 3; }

if [[ "$project" =~ ^[0-9]+$ ]]; then pid="$project"; else pid="$(jq -rn --arg p "$project" '$p|@uri')"; fi

resp="$(curl -sS --fail-with-body -H "PRIVATE-TOKEN: $token" \
  -F "file=@${file}" "https://${host}/api/v4/projects/${pid}/uploads")" \
  || { echo "upload failed: $resp" >&2; exit 4; }

# markdown is project-relative (/uploads/<secret>/<name>) — valid inside that project's issues.
echo "$resp" | jq -r '.markdown'

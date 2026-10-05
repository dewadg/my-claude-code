#!/usr/bin/env bash
# Scaffold a new Go service from assets/scaffold.
# usage: scaffold.sh <dest-dir> <go-module-path> [binary-name]
set -euo pipefail

if [ $# -lt 2 ]; then
  echo "usage: $0 <dest-dir> <go-module-path> [binary-name]" >&2
  exit 2
fi

DEST=$1
MODULE=$2
BIN=${3:-${MODULE##*/}}
SRC="$(cd "$(dirname "$0")/.." && pwd)/assets/scaffold"

if [ -d "$DEST" ] && [ -n "$(ls -A "$DEST" 2>/dev/null | grep -v '^\.git$' || true)" ]; then
  echo "refusing: $DEST exists and is not empty (use Mode B/C of the skill instead)" >&2
  exit 1
fi

mkdir -p "$DEST"
cp -R "$SRC/." "$DEST/"
mv "$DEST/cmd/app" "$DEST/cmd/$BIN"

# Rewrite module path, binary name and service name. Targeted patterns only:
# a bare "app" also appears in SQL and prose.
find "$DEST" -type f \( -name '*.go' -o -name 'go.mod' -o -name 'Makefile' -o -name 'Dockerfile' \) -print0 |
  while IFS= read -r -d '' f; do
    sed -i.bak \
      -e "s#example\.com/app#${MODULE}#g" \
      -e "s#^const Name = \"app\"#const Name = \"${BIN}\"#" \
      -e "s#^BIN     := app#BIN     := ${BIN}#" \
      -e "s#/out/app#/out/${BIN}#g" \
      -e "s#\./cmd/app#./cmd/${BIN}#g" \
      -e "s#^COPY --from=go-builder /out/${BIN} /app#COPY --from=go-builder /out/${BIN} /${BIN}#" \
      -e "s#^ENTRYPOINT \[\"/app\"\]#ENTRYPOINT [\"/${BIN}\"]#" \
      "$f"
    rm -f "$f.bak"
  done

GOVER=$(go env GOVERSION | sed -E 's/^go([0-9]+\.[0-9]+).*/\1/')
sed -i.bak -e "s/^go .*/go ${GOVER}/" "$DEST/go.mod" && rm -f "$DEST/go.mod.bak"
sed -i.bak -e "s/^FROM golang:[0-9.]*-alpine/FROM golang:${GOVER}-alpine/" "$DEST/Dockerfile" && rm -f "$DEST/Dockerfile.bak"

cd "$DEST"
go mod tidy
# The module rename can change import sort order (e.g. gitlab.com sorts after github.com).
gofmt -w .
go build ./... && go vet ./...
[ -f .env ] || cp .env.example .env

echo "scaffolded $MODULE (binary: $BIN) into $DEST"

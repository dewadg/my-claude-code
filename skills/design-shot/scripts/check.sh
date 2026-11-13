#!/usr/bin/env bash
# design-shot dependency check.
# Exit 0: ready. Exit 3: something missing — prints what and the exact fix on stdout.
set -u

DS_HOME="${DESIGN_SHOT_HOME:-$HOME/.cache/design-shot}"
missing=0

if ! command -v node >/dev/null 2>&1; then
  echo "MISSING node — install Node.js >= 18 (e.g. via nvm), then rerun."
  missing=1
else
  major="$(node -p 'process.versions.node.split(".")[0]')"
  if [ "$major" -lt 18 ]; then
    echo "MISSING node >= 18 (found $(node -v)) — upgrade Node.js."
    missing=1
  else
    echo "OK node $(node -v)"
  fi
fi

chrome="${CHROME_PATH:-}"
if [ -z "$chrome" ]; then
  for c in \
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
    "/Applications/Chromium.app/Contents/MacOS/Chromium" \
    "$(command -v google-chrome 2>/dev/null)" \
    "$(command -v google-chrome-stable 2>/dev/null)" \
    "$(command -v chromium 2>/dev/null)" \
    "$(command -v chromium-browser 2>/dev/null)"; do
    if [ -n "$c" ] && [ -x "$c" ]; then chrome="$c"; break; fi
  done
fi
if [ -z "$chrome" ] || [ ! -x "$chrome" ]; then
  echo "MISSING Chrome/Chromium — install Google Chrome, or set CHROME_PATH to a Chromium binary."
  missing=1
else
  echo "OK chrome $chrome"
fi

if [ -d "$DS_HOME/node_modules/puppeteer-core" ]; then
  echo "OK puppeteer-core $(node -p "require('$DS_HOME/node_modules/puppeteer-core/package.json').version" 2>/dev/null)"
else
  echo "MISSING puppeteer-core in $DS_HOME"
  echo "INSTALL: npm install --prefix \"$DS_HOME\" puppeteer-core"
  missing=1
fi

[ "$missing" -eq 0 ] && exit 0 || exit 3

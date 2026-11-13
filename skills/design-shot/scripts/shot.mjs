#!/usr/bin/env node
// design-shot — render a Claude Design canvas artboard locally and screenshot it, whole or a portion.
//
// <dir> is the folder the Artifact tool's `read` saved the canvas files into; it must hold
// project/canvas.json, project/<Board>.dc.html (+ every board it <dc-import>s), and
// artifact-type/dc-runtime.js (served as the ./support.js the boards load).
//
//   node shot.mjs --list --dir <dir>
//   node shot.mjs --info --dir <dir> --board Login.dc.html
//   node shot.mjs --dir <dir> --board Login.dc.html --out login.png
//         [--state view=error]... [--text "Kata sandi" --up 3 | --selector "form"]
//         [--padding 16] [--width 1280 --height 800] [--full] [--scale 2]
//
// Exit: 0 ok · 2 bad args/missing files · 3 puppeteer-core missing (run scripts/check.sh) · 4 render failed

import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { createRequire } from 'node:module';
import { parseArgs } from 'node:util';

const DS_HOME = process.env.DESIGN_SHOT_HOME || path.join(os.homedir(), '.cache', 'design-shot');

const { values: a } = parseArgs({
  options: {
    dir: { type: 'string' },
    board: { type: 'string' },
    list: { type: 'boolean' },
    info: { type: 'boolean' },
    out: { type: 'string' },
    state: { type: 'string', multiple: true },
    selector: { type: 'string' },
    text: { type: 'string' },
    up: { type: 'string' },
    padding: { type: 'string' },
    width: { type: 'string' },
    height: { type: 'string' },
    full: { type: 'boolean' },
    scale: { type: 'string' },
  },
});

const die = (code, msg) => { console.error(msg); process.exit(code); };

if (!a.dir) die(2, 'missing --dir');
const DIR = path.resolve(a.dir);
const PROJECT = path.join(DIR, 'project');
const RUNTIME = path.join(DIR, 'artifact-type', 'dc-runtime.js');

function readCanvas() {
  const p = path.join(PROJECT, 'canvas.json');
  if (!fs.existsSync(p)) die(2, `missing ${p} — read project/canvas.json from the canvas first`);
  return JSON.parse(fs.readFileSync(p, 'utf8'));
}

function boardFile(name) {
  if (!name || name.includes('..') || name.startsWith('/') || name.includes('\\')) die(2, `bad board name: ${name}`);
  return name.endsWith('.dc.html') ? name : `${name}.dc.html`;
}

function readProps(html) {
  const m = html.match(/data-props='([^']*)'/);
  if (!m) return {};
  try { return JSON.parse(m[1]); } catch { return {}; }
}

function scanBoard(file) {
  const p = path.join(PROJECT, file);
  if (!fs.existsSync(p)) return null;
  const html = fs.readFileSync(p, 'utf8');
  const imports = [...html.matchAll(/<dc-import\s+[^>]*name="([^"]+)"/g)].map((m) => boardFile(m[1]));
  const blobs = [...html.matchAll(/\/_blob\/([A-Za-z0-9_-]+)/g)].map((m) => m[1]);
  return { html, imports, blobs, props: readProps(html) };
}

function findBlob(id) {
  const stack = [DIR];
  while (stack.length) {
    const d = stack.pop();
    for (const e of fs.readdirSync(d, { withFileTypes: true })) {
      const p = path.join(d, e.name);
      if (e.isDirectory()) stack.push(p);
      else if (e.name.startsWith(id)) return p;
    }
  }
  return null;
}

// ---------- --list ----------
if (a.list) {
  const c = readCanvas();
  const pages = Object.fromEntries((c.pages || []).map((p) => [p.id, p.name]));
  const firstPage = c.pages?.[0]?.name ?? '';
  const rows = Object.entries(c.boards || {}).map(([file, b]) => ({
    page: pages[b.page] ?? firstPage,
    title: b.title ?? file.replace(/\.dc\.html$/, ''),
    file,
    size: `${b.w}x${b.h}`,
    flags: [b.expand === 'fill' ? 'page' : '', b.is_interactive ? 'interactive' : '', b.w <= 480 ? 'mobile-size' : '']
      .filter(Boolean).join(','),
  }));
  rows.sort((x, y) => x.page.localeCompare(y.page) || x.title.localeCompare(y.title));
  console.log(`canvas: ${c.title ?? '(untitled)'} — ${rows.length} boards`);
  console.log('page\ttitle\tfile\tsize\tflags');
  for (const r of rows) console.log(`${r.page}\t${r.title}\t${r.file}\t${r.size}\t${r.flags}`);
  process.exit(0);
}

if (!a.board) die(2, 'missing --board');
const BOARD = boardFile(a.board);

// Walk the import graph from the board; report what's present, missing, and its props.
function walk(root) {
  const seen = new Set(); const missing = new Set(); const blobs = new Set(); const missingBlobs = new Set();
  const queue = [root];
  let rootProps = {};
  while (queue.length) {
    const f = queue.shift();
    if (seen.has(f)) continue;
    seen.add(f);
    const s = scanBoard(f);
    if (!s) { missing.add(f); continue; }
    if (f === root) rootProps = s.props;
    s.imports.forEach((i) => queue.push(i));
    s.blobs.forEach((b) => { blobs.add(b); if (!findBlob(b)) missingBlobs.add(b); });
  }
  return { seen, missing, blobs, missingBlobs, rootProps };
}

// ---------- --info ----------
if (a.info) {
  const w = walk(BOARD);
  const c = readCanvas();
  const b = c.boards?.[BOARD];
  console.log(`board: ${BOARD}${b ? ` (${b.title ?? ''}, ${b.w}x${b.h}${b.expand === 'fill' ? ', page' : ''})` : ' (not in canvas.json)'}`);
  console.log(`runtime: ${fs.existsSync(RUNTIME) ? 'present' : 'MISSING artifact-type/dc-runtime.js'}`);
  console.log(`imports: ${[...w.seen].filter((f) => f !== BOARD).join(', ') || '(none)'}`);
  console.log(`missing boards: ${[...w.missing].map((f) => `project/${f}`).join(', ') || '(none)'}`);
  console.log(`missing assets (read by id): ${[...w.missingBlobs].join(', ') || '(none)'}`);
  console.log('states/props:');
  for (const [k, v] of Object.entries(w.rootProps)) {
    if (k.startsWith('$')) continue;
    const opts = v.options ? ` options=${JSON.stringify(v.options)}` : '';
    console.log(`  ${k}\teditor=${v.editor ?? '?'}\tdefault=${JSON.stringify(v.default)}${opts}${v.section ? `\tsection=${v.section}` : ''}`);
  }
  process.exit(0);
}

// ---------- shoot ----------
if (!a.out) die(2, 'missing --out');
if (!fs.existsSync(RUNTIME)) die(2, `missing ${RUNTIME} — read artifact-type/dc-runtime.js from the canvas`);
const w = walk(BOARD);
if (w.missing.size) die(2, `missing boards: ${[...w.missing].map((f) => `project/${f}`).join(', ')} — read them first (--info lists them)`);
if (w.missingBlobs.size) console.error(`warn: missing assets render broken: ${[...w.missingBlobs].join(', ')}`);

// --state key=value → overrides the prop's default in every served board that declares it.
const coerce = (v) => (v === 'true' ? true : v === 'false' ? false : v !== '' && !isNaN(Number(v)) ? Number(v) : v);
const overrides = Object.fromEntries((a.state || []).map((s) => {
  const i = s.indexOf('=');
  if (i < 1) die(2, `bad --state ${s} (want key=value)`);
  return [s.slice(0, i), coerce(s.slice(i + 1))];
}));
const unknown = Object.keys(overrides).filter((k) => !(k in w.rootProps));
if (unknown.length) console.error(`warn: board does not declare props: ${unknown.join(', ')} (applied to imports only, if any declare them)`);

// The root board gets its props from data-props defaults; an imported board gets them only from
// attributes on its <dc-import> tag (kebab-case → camelCase). So: rewrite defaults in every board,
// and add the override as an attribute on each <dc-import> whose target declares that prop.
const kebab = (k) => k.replace(/[A-Z]/g, (c) => `-${c.toLowerCase()}`);
const attrVal = (v) => (typeof v === 'string' ? v.replace(/"/g, '&quot;') : `{{ ${JSON.stringify(v)} }}`);
const propsCache = new Map();
const propsOf = (file) => {
  if (!propsCache.has(file)) propsCache.set(file, scanBoard(file)?.props ?? {});
  return propsCache.get(file);
};

function applyState(html) {
  if (!Object.keys(overrides).length) return html;
  html = html.replace(/data-props='([^']*)'/, (m, json) => {
    let props; try { props = JSON.parse(json); } catch { return m; }
    for (const [k, v] of Object.entries(overrides)) if (props[k]) props[k].default = v;
    return `data-props='${JSON.stringify(props).replace(/'/g, '&#39;')}'`;
  });
  return html.replace(/<dc-import\b([^>]*)>/g, (m, attrs) => {
    const name = attrs.match(/\bname="([^"]+)"/)?.[1];
    if (!name) return m;
    const childProps = propsOf(boardFile(name));
    let extra = '';
    for (const [k, v] of Object.entries(overrides)) {
      const attr = kebab(k);
      if (childProps[k] && !new RegExp(`\\s${attr}=`).test(attrs)) extra += ` ${attr}="${attrVal(v)}"`;
    }
    return `<dc-import${attrs}${extra}>`;
  });
}

const TYPES = { '.html': 'text/html', '.js': 'text/javascript', '.json': 'application/json', '.css': 'text/css',
  '.png': 'image/png', '.jpg': 'image/jpeg', '.jpeg': 'image/jpeg', '.gif': 'image/gif', '.webp': 'image/webp',
  '.svg': 'image/svg+xml', '.woff2': 'font/woff2', '.woff': 'font/woff', '.ttf': 'font/ttf', '.otf': 'font/otf' };
const notFound = new Set();

const server = http.createServer((req, res) => {
  const url = decodeURIComponent(new URL(req.url, 'http://x').pathname);
  let file = null;
  if (url === '/support.js') file = RUNTIME;
  else if (url.startsWith('/_blob/')) file = findBlob(url.slice(7).split(/[/.]/)[0]);
  else {
    const p = path.normalize(path.join(PROJECT, url));
    if (p.startsWith(PROJECT) && fs.existsSync(p) && fs.statSync(p).isFile()) file = p;
  }
  if (!file) { notFound.add(url); res.writeHead(404); return res.end(); }
  let body = fs.readFileSync(file);
  if (file.endsWith('.dc.html')) body = applyState(body.toString('utf8'));
  res.writeHead(200, { 'content-type': TYPES[path.extname(file).toLowerCase()] || 'application/octet-stream' });
  res.end(body);
});

let puppeteer;
try {
  puppeteer = createRequire(path.join(DS_HOME, 'package.json'))('puppeteer-core');
} catch {
  die(3, `puppeteer-core not found in ${DS_HOME} — run scripts/check.sh for the install command`);
}

function chromePath() {
  if (process.env.CHROME_PATH) return process.env.CHROME_PATH;
  const c = ['/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    '/Applications/Chromium.app/Contents/MacOS/Chromium',
    '/usr/bin/google-chrome', '/usr/bin/google-chrome-stable', '/usr/bin/chromium', '/usr/bin/chromium-browser'];
  return c.find((p) => fs.existsSync(p)) || die(2, 'Chrome not found — set CHROME_PATH');
}

const canvas = readCanvas();
const meta = canvas.boards?.[BOARD] || {};
const preview = w.rootProps.$preview || {};
const width = Number(a.width || meta.w || preview.width || 1280);
const height = Number(a.height || meta.h || preview.height || 800);
const pad = Number(a.padding ?? 0);

await new Promise((r) => server.listen(0, '127.0.0.1', r));
const port = server.address().port;
const browser = await puppeteer.launch({ executablePath: chromePath(), headless: true, args: ['--hide-scrollbars'] });
try {
  const page = await browser.newPage();
  await page.setViewport({ width, height, deviceScaleFactor: Number(a.scale || 2) });
  await page.goto(`http://127.0.0.1:${port}/${encodeURIComponent(BOARD)}`, { waitUntil: 'networkidle0', timeout: 30000 });
  await page.evaluate(() => document.fonts.ready);
  await new Promise((r) => setTimeout(r, 300));

  let clip = null;
  if (a.selector || a.text) {
    const sel = a.selector || `::-p-text(${a.text})`;
    let el = await page.$(sel);
    if (!el) die(4, `no element matches ${a.selector ? `selector "${a.selector}"` : `text "${a.text}"`}`);
    const up = Number(a.up || 0);
    if (up > 0) {
      el = await el.evaluateHandle((n, k) => { let x = n; for (let i = 0; i < k && x.parentElement; i++) x = x.parentElement; return x; }, up);
    }
    const box = await el.boundingBox();
    if (!box) die(4, 'matched element is not visible');
    const doc = await page.evaluate(() => ({ w: document.documentElement.scrollWidth, h: document.documentElement.scrollHeight }));
    const x = Math.max(0, box.x - pad); const y = Math.max(0, box.y - pad);
    clip = { x, y, width: Math.min(doc.w - x, box.width + pad * 2), height: Math.min(doc.h - y, box.height + pad * 2) };
    await page.screenshot({ path: a.out, clip, captureBeyondViewport: true });
  } else {
    await page.screenshot({ path: a.out, fullPage: !!a.full });
  }
  const real404 = [...notFound].filter((u) => u !== '/favicon.ico');
  if (real404.length) console.error(`warn: not found while rendering: ${real404.join(', ')}`);
  console.log(JSON.stringify({ out: path.resolve(a.out), board: BOARD, viewport: { width, height }, state: overrides, clip }));
} catch (e) {
  die(4, `render failed: ${e.message}`);
} finally {
  await browser.close();
  server.close();
}

#!/usr/bin/env node
/**
 * validate-tokens.mjs — design-token contrast gate (WCAG 2.2 AA).
 *
 * Parses the Tailwind 4 @theme / :root token layer from a CSS file, resolves every
 * semantic foreground/background pairing declared in contrast.config.json (following
 * var() chains), and computes WCAG contrast in EVERY theme. Exits non-zero on any
 * pairing below its required ratio.
 *
 * Zero dependencies. Supports oklch(), #hex, rgb()/rgba(), hsl()/hsla().
 *
 * Usage:
 *   node validate-tokens.mjs [path/to/app.css] [path/to/contrast.config.json]
 * Defaults: src/app.css  and  <this-dir>/contrast.config.json
 */
import { readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const cssPath = resolve(process.argv[2] ?? 'src/app.css');
const configPath = resolve(process.argv[3] ?? join(__dirname, 'contrast.config.json'));

const THRESHOLD = { text: 4.5, large: 3.0, ui: 3.0 };

// ---------- CSS parsing ----------
function stripComments(css) {
  return css.replace(/\/\*[\s\S]*?\*\//g, '');
}
function parseDeclarations(body) {
  const map = new Map();
  for (const m of body.matchAll(/(--[\w-]+)\s*:\s*([^;]+);/g)) {
    map.set(m[1], m[2].trim());
  }
  return map;
}

function extractBalancedBlocks(css, startPattern) {
  const blocks = [];
  startPattern.lastIndex = 0;

  for (let match = startPattern.exec(css); match; match = startPattern.exec(css)) {
    const open = css.indexOf('{', match.index);
    if (open === -1) break;

    let depth = 1;
    let quote = null;
    let escaped = false;
    let i = open + 1;

    for (; i < css.length && depth > 0; i++) {
      const ch = css[i];
      if (quote) {
        if (escaped) {
          escaped = false;
        } else if (ch === '\\') {
          escaped = true;
        } else if (ch === quote) {
          quote = null;
        }
        continue;
      }
      if (ch === '"' || ch === "'") {
        quote = ch;
      } else if (ch === '{') {
        depth++;
      } else if (ch === '}') {
        depth--;
      }
    }

    if (depth !== 0) throw new Error(`unclosed CSS block beginning at offset ${match.index}`);
    blocks.push({ match, body: css.slice(open + 1, i - 1) });
    startPattern.lastIndex = i;
  }

  return blocks;
}

function parseCss(css) {
  css = stripComments(css);
  const base = new Map();
  // :root and @theme (including nested @keyframes) contribute to the base scope.
  for (const { body } of extractBalancedBlocks(css, /(?:^|[}\s])(?:@theme(?:\s+[^\s{]+)?|:root)\s*\{/gm)) {
    for (const [k, v] of parseDeclarations(body)) base.set(k, v);
  }

  // Support both data-attribute themes and the class-based `.dark` convention.
  const themes = new Map();
  const themePattern = /(?:^|[}\s])(?:\.dark|\[data-theme=["']?([\w-]+)["']?\])\s*\{/gm;
  for (const { match, body } of extractBalancedBlocks(css, themePattern)) {
    const name = match[1] ?? 'dark';
    const overlay = themes.get(name) ?? new Map();
    for (const [k, v] of parseDeclarations(body)) overlay.set(k, v);
    themes.set(name, overlay);
  }
  return { base, themes };
}

// ---------- var() resolution ----------
function resolveValue(value, lookup, seen = new Set(), depth = 0) {
  if (depth > 50) throw new Error('var() resolution too deep (cycle?)');
  return value.replace(/var\(\s*(--[\w-]+)\s*(?:,\s*([^()]*))?\)/g, (_, name, fallback) => {
    if (seen.has(name)) throw new Error(`var() cycle at ${name}`);
    const raw = lookup(name);
    if (raw === undefined) {
      if (fallback !== undefined) return resolveValue(fallback.trim(), lookup, seen, depth + 1);
      throw new Error(`unresolved ${name}`);
    }
    return resolveValue(raw, lookup, new Set([...seen, name]), depth + 1);
  }).trim();
}

// ---------- color -> WCAG relative luminance ----------
const clamp01 = (x) => Math.min(1, Math.max(0, x));
const srgbToLinear = (c) => (c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4));
const lumLinear = (r, g, b) => 0.2126 * r + 0.7152 * g + 0.0722 * b;

function num(s) { return s.trim().endsWith('%') ? parseFloat(s) / 100 : parseFloat(s); }

function hslToRgb(h, s, l) {
  h = ((h % 360) + 360) % 360;
  const c = (1 - Math.abs(2 * l - 1)) * s;
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
  const m = l - c / 2;
  const [r, g, b] = h < 60 ? [c, x, 0] : h < 120 ? [x, c, 0] : h < 180 ? [0, c, x]
    : h < 240 ? [0, x, c] : h < 300 ? [x, 0, c] : [c, 0, x];
  return [r + m, g + m, b + m];
}

function oklchToLinearRgb(L, C, Hdeg) {
  const h = (Hdeg * Math.PI) / 180;
  const a = C * Math.cos(h), b = C * Math.sin(h);
  const l_ = L + 0.3963377774 * a + 0.2158037573 * b;
  const m_ = L - 0.1055613458 * a - 0.0638541728 * b;
  const s_ = L - 0.0894841775 * a - 1.2914855480 * b;
  const l = l_ ** 3, m = m_ ** 3, s = s_ ** 3;
  return [
    clamp01(4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s),
    clamp01(-1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s),
    clamp01(-0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s),
  ];
}

function luminance(color) {
  const c = color.trim().toLowerCase();
  let m;
  if ((m = c.match(/^#([0-9a-f]{3,8})$/))) {
    let hex = m[1];
    if (hex.length === 3 || hex.length === 4) hex = [...hex].map((d) => d + d).join('');
    const r = parseInt(hex.slice(0, 2), 16) / 255;
    const g = parseInt(hex.slice(2, 4), 16) / 255;
    const b = parseInt(hex.slice(4, 6), 16) / 255;
    return lumLinear(srgbToLinear(r), srgbToLinear(g), srgbToLinear(b));
  }
  if ((m = c.match(/^rgba?\(([^)]+)\)$/))) {
    const p = m[1].split(/[ ,/]+/).filter(Boolean);
    const ch = p.slice(0, 3).map((v) => (v.endsWith('%') ? num(v) : parseFloat(v) / 255));
    return lumLinear(...ch.map(srgbToLinear));
  }
  if ((m = c.match(/^hsla?\(([^)]+)\)$/))) {
    const p = m[1].split(/[ ,/]+/).filter(Boolean);
    const [r, g, b] = hslToRgb(parseFloat(p[0]), num(p[1]), num(p[2]));
    return lumLinear(srgbToLinear(r), srgbToLinear(g), srgbToLinear(b));
  }
  if ((m = c.match(/^oklch\(([^)]+)\)$/))) {
    const p = m[1].split(/[ ,/]+/).filter(Boolean);
    const [r, g, b] = oklchToLinearRgb(num(p[0]), parseFloat(p[1]), parseFloat(p[2]) || 0);
    return lumLinear(r, g, b); // already linear-light
  }
  throw new Error(`unparseable color: ${color}`);
}

function contrast(c1, c2) {
  const l1 = luminance(c1), l2 = luminance(c2);
  const [hi, lo] = l1 >= l2 ? [l1, l2] : [l2, l1];
  return (hi + 0.05) / (lo + 0.05);
}

// ---------- main ----------
const { base, themes } = parseCss(readFileSync(cssPath, 'utf8'));
const config = JSON.parse(readFileSync(configPath, 'utf8'));
const themeNames = ['base', ...themes.keys()].filter((t) => !(config.ignoreThemes ?? []).includes(t));
const missingThemes = (config.requiredThemes ?? []).filter((theme) => !themes.has(theme));

const lookupFor = (theme) => (name) => {
  if (theme !== 'base') {
    const overlaid = themes.get(theme)?.get(name);
    if (overlaid !== undefined) return overlaid;
  }
  return base.get(name);
};

let failures = 0;
const rows = [];
for (const theme of missingThemes) {
  failures++;
  rows.push({
    theme,
    pair: 'required theme',
    level: '—',
    required: '—',
    ratio: '—',
    status: 'ERROR',
    detail: 'theme block not found',
  });
}
for (const theme of themeNames) {
  const lookup = lookupFor(theme);
  for (const pair of config.pairs) {
    const required = THRESHOLD[pair.level] ?? THRESHOLD.text;
    let ratio, status, detail = '';
    try {
      const fg = resolveValue(`var(--color-${pair.fg})`, lookup);
      const bg = resolveValue(`var(--color-${pair.bg})`, lookup);
      ratio = contrast(fg, bg);
      status = ratio >= required ? 'PASS' : 'FAIL';
    } catch (e) {
      status = 'ERROR';
      detail = e.message;
    }
    if (status !== 'PASS') failures++;
    rows.push({
      theme, pair: `${pair.fg} on ${pair.bg}`, level: pair.level,
      required, ratio: ratio ? ratio.toFixed(2) : '—', status, detail,
    });
  }
}

const pad = (s, n) => String(s).padEnd(n);
const themeWidth = Math.max(10, ...rows.map((row) => String(row.theme).length + 2));
const pairWidth = Math.max(26, ...rows.map((row) => String(row.pair).length + 2));
const ruleWidth = themeWidth + pairWidth + 7 + 6 + 8 + 8;
console.log(pad('THEME', themeWidth) + pad('PAIRING', pairWidth) + pad('LVL', 7) + pad('REQ', 6) + pad('RATIO', 8) + 'STATUS');
console.log('-'.repeat(ruleWidth));
for (const r of rows) {
  console.log(
    pad(r.theme, themeWidth) + pad(r.pair, pairWidth) + pad(r.level, 7) + pad(r.required, 6) +
    pad(r.ratio, 8) + r.status + (r.detail ? `  (${r.detail})` : ''),
  );
}
console.log('-'.repeat(ruleWidth));
if (failures) {
  console.error(`\n✖ ${failures} contrast gate failure(s).`);
  process.exit(1);
}
console.log('\n✓ All token contrast pairings meet WCAG 2.2 AA across all themes.');

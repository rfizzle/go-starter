import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import test from 'node:test';

const script = join(dirname(fileURLToPath(import.meta.url)), 'validate-tokens.mjs');

function run(css, pairs, options = {}) {
  const dir = mkdtempSync(join(tmpdir(), 'validate-tokens-'));
  const cssPath = join(dir, 'app.css');
  const configPath = join(dir, 'contrast.json');
  writeFileSync(cssPath, css);
  writeFileSync(configPath, JSON.stringify({ pairs, ...options }));
  return spawnSync(process.execPath, [script, cssPath, configPath], { encoding: 'utf8' });
}

test('parses nested @theme content and class-based dark overrides', () => {
  const result = run(`
    @theme {
      --color-fg: oklch(20% 0 0);
      --color-surface: oklch(100% 0 0);
      --animate-pulse: pulse 1s linear;
      @keyframes pulse { from { opacity: 0; } to { opacity: 1; } }
    }
    .dark {
      --color-fg: oklch(95% 0 0);
      --color-surface: oklch(15% 0 0);
    }
  `, [{ fg: 'fg', bg: 'surface', level: 'text' }]);

  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /base\s+fg on surface/);
  assert.match(result.stdout, /dark\s+fg on surface/);
});

test('fails when a dark-theme pairing does not meet its threshold', () => {
  const result = run(`
    @theme { --color-fg: #000; --color-surface: #fff; }
    .dark { --color-fg: #777; --color-surface: #666; }
  `, [{ fg: 'fg', bg: 'surface', level: 'text' }]);

  assert.equal(result.status, 1);
  assert.match(result.stdout, /dark\s+fg on surface.*FAIL/);
});

test('reports missing tokens as errors', () => {
  const result = run(
    '@theme { --color-surface: #fff; }',
    [{ fg: 'fg', bg: 'surface', level: 'text' }],
  );

  assert.equal(result.status, 1);
  assert.match(result.stdout, /ERROR/);
  assert.match(result.stdout, /unresolved --color-fg/);
});

test('fails when a required theme block is absent', () => {
  const result = run(
    '@theme { --color-fg: #000; --color-surface: #fff; }',
    [{ fg: 'fg', bg: 'surface', level: 'text' }],
    { requiredThemes: ['dark'] },
  );

  assert.equal(result.status, 1);
  assert.match(result.stdout, /dark\s+required theme.*ERROR/);
});

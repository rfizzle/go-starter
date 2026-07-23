import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import test from 'node:test';

const script = join(dirname(fileURLToPath(import.meta.url)), 'validate-spec.mjs');

function validSpec(extra = '') {
  return `# Button component spec

## Purpose
Submit an action.

## Anatomy
One native button element.

## Variants
Primary and quiet.

## States
Default, hover, focus-visible, disabled, and busy.

## Props API
Typed Svelte props and an onclick callback.

## Accessibility contract
APG/native pattern: \`button\`
Source: https://html.spec.whatwg.org/multipage/form-elements.html#the-button-element
Backing primitive: native element.

## Content
Concise action label.

## Responsive
No behavior change.

## Tokens used and configured scales
Semantic colors and configured spacing scales only.

${extra}`;
}

function run(spec) {
  const dir = mkdtempSync(join(tmpdir(), 'validate-component-spec-'));
  const file = join(dir, 'Button.spec.md');
  writeFileSync(file, spec);
  return spawnSync(process.execPath, [script, file], { encoding: 'utf8' });
}

test('accepts a complete native-element contract', () => {
  const result = run(validSpec());
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /Spec gate passed/);
});

test('rejects a missing governing accessibility source', () => {
  const result = run(validSpec().replace(/^Source:.*\n/m, ''));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /no governing accessibility source recorded/i);
});

test('rejects raw color literals', () => {
  const result = run(validSpec('Focus ring: #2563eb.'));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /raw color literal/i);
});

test('rejects blank accessibility metadata that only leaves labels', () => {
  const result = run(validSpec()
    .replace('APG/native pattern: `button`', 'APG/native pattern: · Source: · Backing primitive:')
    .replace(/^Source:.*\n/m, '')
    .replace(/^Backing primitive:.*\n/m, ''));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /no governing native or ARIA APG pattern named/i);
  assert.match(result.stderr, /no governing accessibility source recorded/i);
  assert.match(result.stderr, /no backing primitive declared/i);
});

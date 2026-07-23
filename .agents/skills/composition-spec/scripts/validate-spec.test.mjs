import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import { validateCompositionSpec } from './validate-spec.mjs';

const complete = `# FilterPanel Composition Spec

> Status: draft · Spec version: 2 · Owner: frontend · Feature: search

## Purpose
Combines the query input and filter controls used by searchable list pages.

## Children
| Component | Source | Props passed | Events consumed |
|---|---|---|---|
| Input | $lib/components | value | onchange |

## Props API
\`\`\`ts
type Props = {
  value: string;
  onchange: (value: string) => void;
};
\`\`\`

## Internal State
Owns only the locally expanded filter section; query data remains caller-owned.

## Interaction Flows
| User action | Internal effect | Emitted event or navigation |
|---|---|---|
| Types a query | None | onchange(value) |

## Layout
Uses the configured grid, gap, and breakpoint scales.

## Accessibility Delegation
Preserves DOM reading order; Input owns its accessible name and keyboard behavior.

## Testing Scenarios
- Query changes reach the callback.
- Focus order follows visual order.
`;

test('accepts a filled composition contract', () => {
  assert.deepEqual(validateCompositionSpec(complete), []);
});

test('rejects missing accessibility delegation and raw colors', () => {
  const invalid = complete
    .replace('## Accessibility Delegation', '## Coordination')
    .concat('\nBorder: #334455\n');
  const failures = validateCompositionSpec(invalid);
  assert.ok(failures.some((failure) => failure.includes('Accessibility Delegation')));
  assert.ok(failures.some((failure) => failure.includes('raw color literal')));
});

test('rejects an almost untouched template', () => {
  const template = readFileSync(join(import.meta.dirname, '..', 'references', 'spec-template.md'), 'utf8')
    .replace('<CompositionName>', 'FilterPanel');
  const failures = validateCompositionSpec(template);
  assert.ok(failures.some((failure) => failure.includes('Children table has no component row')));
  assert.ok(failures.some((failure) => failure.includes('template instruction remains')));
});

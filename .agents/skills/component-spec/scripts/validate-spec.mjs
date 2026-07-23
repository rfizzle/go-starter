#!/usr/bin/env node
/**
 * Portable component contract validator.
 *
 * Import `validateComponentSpec(text)` from a client or project harness, or run:
 *   node validate-spec.mjs path/to/Component.spec.md
 */
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const REQUIRED_SECTIONS = [
  'Purpose', 'Anatomy', 'Variants', 'States', 'Props API',
  'Accessibility contract', 'Content', 'Responsive', 'Tokens used',
];
const PLACEHOLDERS = [
  '<Component>', '<PascalName>', '<pattern-name>', '<component>', '<one-line brief>',
  '<official URL or native HTML specification>',
];
const COLOR_RE = /(#[0-9a-fA-F]{3,8}\b|\boklch\(|\brgba?\(|\bhsla?\()/g;

export function validateComponentSpec(text, {
  requireSource = true,
  requireFilledSections = true,
} = {}) {
  const lines = text.split('\n');
  const failures = [];
  const headings = lines
    .filter((line) => /^#{1,6}\s/.test(line))
    .map((line) => line.replace(/^#{1,6}\s+/, '').toLowerCase());

  for (const section of REQUIRED_SECTIONS) {
    const index = headings.findIndex((heading) => heading.includes(section.toLowerCase()));
    if (index === -1) {
      failures.push(`missing required section: "${section}"`);
    } else if (requireFilledSections) {
      const headingLines = lines
        .map((line, lineIndex) => (/^#{1,6}\s/.test(line) ? lineIndex : -1))
        .filter((lineIndex) => lineIndex >= 0);
      const start = headingLines[index] + 1;
      const end = headingLines[index + 1] ?? lines.length;
      if (!lines.slice(start, end).join('\n').trim()) {
        failures.push(`empty required section: "${section}"`);
      }
    }
  }
  for (const placeholder of PLACEHOLDERS) {
    if (text.includes(placeholder)) failures.push(`unfilled placeholder present: ${placeholder}`);
  }

  const model = text.match(/(?:Governing\s+)?(?:APG\/native|native\/APG|APG|native) pattern:\**[\t ]*(?:`([^`·\n]+)`|([^·\n]+))/i);
  const modelValue = (model?.[1] ?? model?.[2] ?? '').trim();
  if (!modelValue || /^[-–—\s]*$/.test(modelValue)) {
    failures.push('no governing native or ARIA APG pattern named');
  }
  if (requireSource && !/Source:\**[\t ]*(?:https?:\/\/[^\s·]+|native HTML[^·\n]*)/i.test(text)) {
    failures.push('no governing accessibility source recorded');
  }
  const primitive = text.match(/Backing primitive:\**[\t ]*(?:`([^`·\n]+)`|([^·\n]+))/i);
  const primitiveValue = (primitive?.[1] ?? primitive?.[2] ?? '').trim();
  if (!primitiveValue) {
    failures.push('no backing primitive declared (native / established primitive / justified manual)');
  }

  lines.forEach((line, index) => {
    const matches = line.match(COLOR_RE);
    if (matches) {
      failures.push(`raw color literal on line ${index + 1}: ${matches.join(', ')} — use a semantic token`);
    }
  });
  return failures;
}

function runCli() {
  if (!process.argv[2]) {
    console.error('Usage: node validate-spec.mjs <path-to-spec.md>');
    process.exitCode = 2;
    return;
  }

  const specPath = resolve(process.argv[2]);
  const failures = validateComponentSpec(readFileSync(specPath, 'utf8'));
  if (failures.length) {
    console.error(`✖ Spec gate failed for ${specPath}:\n`);
    for (const failure of failures) console.error(`  - ${failure}`);
    process.exitCode = 1;
    return;
  }
  console.log(`✓ Spec gate passed: ${specPath}`);
}

if (process.argv[1] && fileURLToPath(import.meta.url) === resolve(process.argv[1])) runCli();

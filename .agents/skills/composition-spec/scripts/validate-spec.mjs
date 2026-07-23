#!/usr/bin/env node
/**
 * Portable composition contract validator.
 *
 * Import `validateCompositionSpec(text)` from a client or project harness, or run:
 *   node validate-spec.mjs path/to/Composition.spec.md
 */
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const REQUIRED_SECTIONS = [
  'Purpose', 'Children', 'Props API', 'Internal State', 'Interaction Flows', 'Layout',
  'Accessibility Delegation', 'Testing Scenarios',
];
const PLACEHOLDERS = ['<CompositionName>', '<PascalName>', '<feature-name>'];
const TEMPLATE_INSTRUCTIONS = [
  'What the composition does, where it appears',
  'State the composition owns locally',
  'Describe semantic structure, responsive behavior',
  'Describe reading order, focus sequence',
  'List unresolved decisions and the person or artifact',
];
const COLOR_RE = /(#[0-9a-fA-F]{3,8}\b|\boklch\(|\brgba?\(|\bhsla?\()/g;

function sectionBody(lines, section) {
  const headingLines = lines
    .map((line, index) => (/^#{1,6}\s/.test(line) ? index : -1))
    .filter((index) => index >= 0);
  const headingIndex = headingLines.findIndex((index) => lines[index]
    .replace(/^#{1,6}\s+/, '')
    .toLowerCase()
    .includes(section.toLowerCase()));
  if (headingIndex === -1) return null;
  return lines.slice(headingLines[headingIndex] + 1, headingLines[headingIndex + 1] ?? lines.length).join('\n');
}

function hasDataRow(body) {
  const rows = body.split('\n').filter((line) => /^\s*\|/.test(line) && !/^\s*\|(?:\s*:?-+:?\s*\|)+\s*$/.test(line));
  return rows.length >= 2;
}

export function validateCompositionSpec(text, {
  requireAccessibilityDelegation = true,
  requireFilledSections = true,
} = {}) {
  const lines = text.split('\n');
  const failures = [];
  const headings = lines
    .filter((line) => /^#{1,6}\s/.test(line))
    .map((line) => line.replace(/^#{1,6}\s+/, '').toLowerCase());

  const requiredSections = requireAccessibilityDelegation
    ? REQUIRED_SECTIONS
    : REQUIRED_SECTIONS.filter((section) => section !== 'Accessibility Delegation');
  for (const section of requiredSections) {
    const body = sectionBody(lines, section);
    if (body === null) {
      failures.push(`missing required section: "${section}"`);
    } else if (requireFilledSections && !body.trim()) {
      failures.push(`empty required section: "${section}"`);
    }
  }
  for (const placeholder of PLACEHOLDERS) {
    if (text.includes(placeholder)) failures.push(`unfilled placeholder present: ${placeholder}`);
  }
  if (requireFilledSections) {
    for (const instruction of TEMPLATE_INSTRUCTIONS) {
      if (text.includes(instruction)) failures.push(`template instruction remains: ${instruction}`);
    }
    const children = sectionBody(lines, 'Children');
    if (children !== null && !hasDataRow(children)) failures.push('Children table has no component row');
    const flows = sectionBody(lines, 'Interaction Flows');
    if (flows !== null && !hasDataRow(flows)) failures.push('Interaction Flows table has no scenario row');
    const props = sectionBody(lines, 'Props API') ?? '';
    if (!/^\s*[A-Za-z_$][\w$]*\??\s*:/m.test(props)) failures.push('Props API has no typed property');
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
  const failures = validateCompositionSpec(readFileSync(specPath, 'utf8'));
  if (failures.length) {
    console.error(`✖ Spec gate failed for ${specPath}:\n`);
    for (const failure of failures) console.error(`  - ${failure}`);
    process.exitCode = 1;
    return;
  }
  console.log(`✓ Spec gate passed: ${specPath}`);
}

if (process.argv[1] && fileURLToPath(import.meta.url) === resolve(process.argv[1])) runCli();

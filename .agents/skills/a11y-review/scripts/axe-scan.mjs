#!/usr/bin/env node
/**
 * axe-scan.mjs — runtime accessibility gate.
 *
 * Enumerates every Storybook story via index.json, renders each in a headless browser,
 * runs axe-core against it, and fails on any returned WCAG A/AA violation. This is the
 * one gate that needs real dependencies (driving a browser is unavoidable).
 *
 * Prerequisites (install with the project's package manager):
 *   playwright, @axe-core/playwright, and a Playwright Chromium browser
 * And a Storybook reachable at the base URL (running `storybook dev`, or a static
 * `storybook build` served on a port).
 *
 * Usage:
 *   node axe-scan.mjs [baseUrl]          # default http://localhost:6006
 */
const baseUrl = (process.argv[2] ?? 'http://localhost:6006').replace(/\/$/, '');

let chromium, AxeBuilder;
try {
  ({ chromium } = await import('playwright'));
  ({ default: AxeBuilder } = await import('@axe-core/playwright'));
} catch {
  console.error('Missing playwright/@axe-core dependencies. Install them with the project package manager and install the Chromium browser.');
  process.exit(2);
}

async function getStories() {
  const res = await fetch(`${baseUrl}/index.json`).catch(() => null);
  if (!res || !res.ok) {
    console.error(`Could not fetch ${baseUrl}/index.json — is Storybook running at ${baseUrl}?`);
    process.exit(2);
  }
  const json = await res.json();
  const entries = Object.values(json.entries ?? json.stories ?? {});
  return entries.filter((e) => e.type !== 'docs').map((e) => ({ id: e.id, title: `${e.title} / ${e.name}` }));
}

const stories = await getStories();
if (!stories.length) {
  console.error('No stories found in index.json.');
  process.exit(2);
}

const browser = await chromium.launch();
const page = await browser.newPage();
let totalBlocking = 0;

for (const story of stories) {
  const url = `${baseUrl}/iframe.html?id=${encodeURIComponent(story.id)}&viewMode=story`;
  await page.goto(url, { waitUntil: 'networkidle' });
  await page.waitForFunction(() => {
    const root = document.querySelector('#storybook-root, #root');
    const preparing = document.querySelector('.sb-preparing-story, .sb-loader');
    return root?.childElementCount > 0 && !preparing;
  });
  await page.evaluate(async () => {
    await document.fonts?.ready;
    await new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve)));
  });
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
    .analyze();
  const blocking = results.violations;
  if (blocking.length) {
    totalBlocking += blocking.length;
    console.error(`\n✖ ${story.title}`);
    for (const v of blocking) {
      console.error(`   [${v.impact}] ${v.id} — ${v.help} (${v.nodes.length} node(s))`);
      console.error(`     ${v.helpUrl}`);
    }
  } else {
    console.log(`✓ ${story.title}`);
  }
}

await browser.close();

if (totalBlocking) {
  console.error(`\n✖ Accessibility gate failed: ${totalBlocking} WCAG A/AA violation(s).`);
  process.exit(1);
}
console.log(`\n✓ Accessibility gate passed across ${stories.length} stories (no WCAG A/AA violations).`);

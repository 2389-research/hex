// ABOUTME: Verifies the alpha page renders a visible download metadata error state.
// ABOUTME: Reads the built static HTML so the check matches deployable output.

import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';

const htmlPath = resolve('dist/alpha/index.html');
const html = await readFile(htmlPath, 'utf8');

const requiredText = [
  'DOWNLOAD MANIFEST UNAVAILABLE',
  'Refresh this page to retry',
];

for (const text of requiredText) {
  if (!html.includes(text)) {
    throw new Error(`Missing alpha metadata error text: ${text}`);
  }
}

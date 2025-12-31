#!/usr/bin/env node
// ABOUTME: Updates Homebrew formula in homebrew-tap to use public Firebase URLs
// ABOUTME: Replaces GitHub release URLs (from private repo) with Firebase Storage URLs

const https = require('https');
const { Octokit } = require('@octokit/rest');

const BUCKET_NAME = 'hex-code-daf69.firebasestorage.app';
const VERSIONS_PATH = 'releases/versions.json';
const TAP_OWNER = '2389-research';
const TAP_REPO = 'homebrew-tap';
const FORMULA_PATH = 'Formula/hex.rb';

function firebaseUrl(path) {
  return `https://firebasestorage.googleapis.com/v0/b/${BUCKET_NAME}/o/${encodeURIComponent(path)}?alt=media`;
}

function fetchJson(url) {
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          resolve(JSON.parse(data));
        } catch (e) {
          reject(new Error(`Failed to parse JSON from ${url}: ${e.message}`));
        }
      });
    }).on('error', reject);
  });
}

async function getVersionsJson() {
  const url = firebaseUrl(VERSIONS_PATH);
  console.log(`Fetching versions.json from ${url}`);
  return await fetchJson(url);
}

async function updateFormula(octokit, version, versionsData) {
  console.log(`\nUpdating Homebrew formula for v${version}`);

  // Get current formula
  console.log('[1/4] Fetching current formula from homebrew-tap...');
  const { data: currentFile } = await octokit.repos.getContent({
    owner: TAP_OWNER,
    repo: TAP_REPO,
    path: FORMULA_PATH,
  });

  const currentFormula = Buffer.from(currentFile.content, 'base64').toString('utf8');
  console.log('  Formula fetched');

  // Get release data from versions.json
  const release = versionsData.releases.find(r => r.version === version);
  if (!release) {
    throw new Error(`Version ${version} not found in versions.json`);
  }

  console.log('[2/4] Parsing release data...');
  console.log(`  Found ${Object.keys(release.downloads).length} downloads`);
  console.log(`  Found ${Object.keys(release.checksums).length} checksums`);

  // Update formula with Firebase URLs and checksums
  console.log('[3/4] Updating formula URLs and checksums...');
  let updatedFormula = currentFormula;

  // Build a map of filenames to Firebase URLs and checksums
  // e.g., "hex_1.9.0_Darwin_arm64.tar.gz" -> { url: "https://...", checksum: "..." }
  const filenameMap = {};
  for (const [key, url] of Object.entries(release.downloads)) {
    const filename = url.split('/').pop(); // Get filename from URL
    filenameMap[filename] = {
      url: url,
      checksum: release.checksums[key]?.replace('sha256:', '') || ''
    };
  }

  console.log(`  Built map for ${Object.keys(filenameMap).length} artifacts`);

  // Replace URLs and SHA256s together in a single pass
  // Match pattern: url "..." followed by sha256 "..." (with newlines/whitespace between)
  // The formula has format like:
  //   url "https://..."
  //         sha256 "..."
  const urlShaPattern = /url\s+"https:\/\/github\.com\/[^"]+\/(hex_[^"]+\.tar\.gz)"\s+sha256\s+"[a-f0-9]{64}"/gs;

  let replacementCount = 0;
  updatedFormula = updatedFormula.replace(urlShaPattern, (match, filename) => {
    if (filenameMap[filename]) {
      const { url, checksum } = filenameMap[filename];
      console.log(`  Replacing ${filename}`);
      replacementCount++;
      // Preserve the indentation style from the original
      return `url "${url}"\n      sha256 "${checksum}"`;
    }
    return match; // Keep original if not found
  });

  console.log(`  Replaced ${replacementCount} URL+SHA256 pairs`);

  // Verify we actually made changes
  if (updatedFormula === currentFormula) {
    throw new Error('Formula was not modified - pattern matching may have failed');
  }

  // Commit updated formula
  console.log('[4/4] Committing updated formula...');
  await octokit.repos.createOrUpdateFileContents({
    owner: TAP_OWNER,
    repo: TAP_REPO,
    path: FORMULA_PATH,
    message: `chore: update hex formula v${version} with Firebase URLs

Updated download URLs to use public Firebase Storage instead of private GitHub releases.

This ensures the formula works for public users via \`brew install 2389-research/tap/hex\`.`,
    content: Buffer.from(updatedFormula).toString('base64'),
    sha: currentFile.sha,
  });

  console.log('  Formula committed successfully');
}

async function main() {
  try {
    // Parse arguments
    const args = process.argv.slice(2);
    if (args.length < 1) {
      console.error('Usage: node update-homebrew-formula.js <version-tag>');
      console.error('Example: node update-homebrew-formula.js v1.7.0');
      process.exit(1);
    }

    const versionTag = args[0];
    const version = versionTag.replace(/^v/, '');

    // Check for GitHub token
    const token = process.env.GITHUB_TOKEN || process.env.HOMEBREW_TAP_TOKEN;
    if (!token) {
      throw new Error('GITHUB_TOKEN or HOMEBREW_TAP_TOKEN environment variable required');
    }

    // Initialize Octokit
    const octokit = new Octokit({ auth: token });

    // Get versions.json from Firebase
    const versionsData = await getVersionsJson();

    // Update the formula
    await updateFormula(octokit, version, versionsData);

    console.log('\n✅ Homebrew formula updated successfully!');
    console.log(`   Users can now install with: brew install ${TAP_OWNER}/tap/hex`);

  } catch (error) {
    console.error('\n❌ Failed to update Homebrew formula:', error.message);
    if (error.response) {
      console.error('   GitHub API error:', error.response.data);
    }
    process.exit(1);
  }
}

main();

#!/usr/bin/env node
// ABOUTME: Deploys hex releases to Firebase Storage and updates versions.json manifest
// ABOUTME: Uploads artifacts from GoReleaser dist/ directory to Firebase Storage

const admin = require('firebase-admin');
const fs = require('fs');
const path = require('path');

const BUCKET_NAME = 'hex-code-daf69.firebasestorage.app';
const VERSIONS_PATH = 'releases/versions.json';
const MAX_RELEASES = 5;

// Initialize Firebase Admin
admin.initializeApp({
  projectId: 'hex-code-daf69',
  storageBucket: BUCKET_NAME
});

const bucket = admin.storage().bucket();

// Parse command line arguments
const args = process.argv.slice(2);
if (args.length < 2) {
  console.error('Usage: node deploy-firebase.js <version-tag> <dist-dir>');
  console.error('Example: node deploy-firebase.js v1.7.0 dist/');
  process.exit(1);
}

const [versionTag, distDir] = args;
const version = versionTag.replace(/^v/, '');

function firebaseUrl(remotePath) {
  return `https://firebasestorage.googleapis.com/v0/b/${BUCKET_NAME}/o/${encodeURIComponent(remotePath)}?alt=media`;
}

async function uploadFile(localPath, remotePath) {
  await bucket.upload(localPath, {
    destination: remotePath,
    metadata: {
      contentType: 'application/gzip',
      cacheControl: 'public, max-age=31536000'
    }
  });
  return firebaseUrl(remotePath);
}

async function uploadBuffer(buffer, remotePath, contentType = 'application/json') {
  const file = bucket.file(remotePath);
  await file.save(buffer, {
    metadata: {
      contentType,
      cacheControl: 'public, max-age=300'
    }
  });
  return firebaseUrl(remotePath);
}

async function downloadVersionsJson() {
  try {
    const file = bucket.file(VERSIONS_PATH);
    const [exists] = await file.exists();
    if (!exists) {
      return { product: 'hex', latest: '', releases: [] };
    }
    const [content] = await file.download();
    return JSON.parse(content.toString());
  } catch (error) {
    console.log('No existing versions.json, starting fresh');
    return { product: 'hex', latest: '', releases: [] };
  }
}

function parseChecksums(distDir) {
  const checksumFile = path.join(distDir, 'checksums.txt');
  if (!fs.existsSync(checksumFile)) {
    throw new Error(`Checksums file not found: ${checksumFile}`);
  }

  const content = fs.readFileSync(checksumFile, 'utf8');
  const checksums = {};

  for (const line of content.trim().split('\n')) {
    const [hash, filename] = line.trim().split(/\s+/);
    if (filename && filename.endsWith('.tar.gz')) {
      // Parse filename like hex_1.7.0_Darwin_arm64.tar.gz
      const match = filename.match(/hex_[\d.]+_(\w+)_(\w+)\.tar\.gz/);
      if (match) {
        const [, os, arch] = match;
        const key = `${os.toLowerCase()}_${arch.toLowerCase()}`;
        checksums[key] = `sha256:${hash}`;
      }
    }
  }

  return checksums;
}

function findTarGzFiles(distDir) {
  const files = fs.readdirSync(distDir);
  return files.filter(f => f.endsWith('.tar.gz') && f.startsWith('hex_'));
}

async function main() {
  try {
    console.log(`\nDeploying hex v${version}`);
    console.log(`Distribution directory: ${distDir}\n`);

    // Step 1: Find and upload tar.gz files
    console.log('[1/3] Uploading release artifacts...');
    const tarFiles = findTarGzFiles(distDir);
    if (tarFiles.length === 0) {
      throw new Error(`No tar.gz files found in ${distDir}`);
    }

    const uploads = {};
    for (const file of tarFiles) {
      const localPath = path.join(distDir, file);
      const remotePath = `releases/v${version}/${file}`;
      console.log(`  Uploading ${file}...`);
      uploads[file] = await uploadFile(localPath, remotePath);
    }
    console.log(`  Uploaded ${tarFiles.length} files`);

    // Also upload checksums.txt
    const checksumFile = path.join(distDir, 'checksums.txt');
    if (fs.existsSync(checksumFile)) {
      await uploadFile(checksumFile, `releases/v${version}/checksums.txt`);
      console.log('  Uploaded checksums.txt');
    }

    // Step 2: Parse checksums
    console.log('[2/3] Parsing checksums...');
    const checksums = parseChecksums(distDir);
    console.log(`  Found ${Object.keys(checksums).length} checksums`);

    // Step 3: Update versions.json
    console.log('[3/3] Updating versions.json...');
    const versionsData = await downloadVersionsJson();

    const downloads = {};
    for (const [file, url] of Object.entries(uploads)) {
      const match = file.match(/hex_[\d.]+_(\w+)_(\w+)\.tar\.gz/);
      if (match) {
        const [, os, arch] = match;
        downloads[`${os.toLowerCase()}_${arch.toLowerCase()}`] = url;
      }
    }

    const newRelease = {
      version,
      date: new Date().toISOString(),
      downloads,
      checksums
    };

    // Prepend new release and trim to max
    versionsData.releases = [
      newRelease,
      ...versionsData.releases.filter(r => r.version !== version)
    ].slice(0, MAX_RELEASES);
    versionsData.latest = version;

    await uploadBuffer(
      Buffer.from(JSON.stringify(versionsData, null, 2)),
      VERSIONS_PATH
    );
    console.log('  versions.json updated');

    console.log('\nDeployment complete!');
    console.log(`  versions.json: ${firebaseUrl(VERSIONS_PATH)}`);

  } catch (error) {
    console.error('\nDeployment failed:', error.message);
    process.exit(1);
  }
}

main();

#!/usr/bin/env node
'use strict';

const https = require('https');
const http = require('http');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { execFileSync, spawnSync } = require('child_process');
const crypto = require('crypto');
const zlib = require('zlib');

// ─── helpers ──────────────────────────────────────────────────────────────────

const VERSION = require('../package.json').version;
const BASE_URL = `https://github.com/nathanramorim/forge-sdd/releases/download/v${VERSION}`;

function platformAsset() {
  const plat = process.platform;
  const arch = process.arch;

  const map = {
    'linux-x64':    { file: `forge-sdd_linux_amd64.tar.gz`,   ext: 'tar.gz' },
    'linux-arm64':  { file: `forge-sdd_linux_arm64.tar.gz`,   ext: 'tar.gz' },
    'darwin-x64':   { file: `forge-sdd_darwin_amd64.tar.gz`,  ext: 'tar.gz' },
    'darwin-arm64': { file: `forge-sdd_darwin_arm64.tar.gz`,  ext: 'tar.gz' },
    'win32-x64':    { file: `forge-sdd_windows_amd64.zip`,    ext: 'zip'    },
  };

  const key = `${plat}-${arch}`;
  const asset = map[key];
  if (!asset) {
    console.error(`forge-sdd: unsupported platform ${key}`);
    process.exit(1);
  }
  return asset;
}

function cacheDir() {
  const base = process.env.XDG_CACHE_HOME
    || (process.platform === 'win32'
        ? path.join(os.homedir(), 'AppData', 'Local')
        : path.join(os.homedir(), '.cache'));
  return path.join(base, 'forge-sdd', VERSION);
}

function binaryPath(dir) {
  const name = process.platform === 'win32' ? 'forge-sdd.exe' : 'forge-sdd';
  return path.join(dir, name);
}

// ─── download ─────────────────────────────────────────────────────────────────

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    const protocol = url.startsWith('https') ? https : http;
    const req = protocol.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        file.close();
        fs.unlinkSync(dest);
        return download(res.headers.location, dest).then(resolve).catch(reject);
      }
      if (res.statusCode !== 200) {
        file.close();
        fs.unlinkSync(dest);
        return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
      }
      res.pipe(file);
      file.on('finish', () => { file.close(); resolve(); });
    });
    req.on('error', (err) => { fs.unlinkSync(dest); reject(err); });
  });
}

function sha256(filePath) {
  const data = fs.readFileSync(filePath);
  return crypto.createHash('sha256').update(data).digest('hex');
}

async function verifyChecksum(assetFile, binDir) {
  const checksumsUrl = `${BASE_URL}/checksums.txt`;
  const checksumsPath = path.join(binDir, 'checksums.txt');
  await download(checksumsUrl, checksumsPath);

  const lines = fs.readFileSync(checksumsPath, 'utf8').split('\n');
  const entry = lines.find(l => l.includes(assetFile));
  if (!entry) {
    throw new Error(`checksum not found for ${assetFile}`);
  }
  return entry.split(/\s+/)[0].toLowerCase();
}

// ─── extract ──────────────────────────────────────────────────────────────────

function extractTarGz(archive, destDir) {
  // Use system tar — available on all platforms we target
  const result = spawnSync('tar', ['-xzf', archive, '-C', destDir, '--strip-components=0'], {
    stdio: 'inherit',
  });
  if (result.status !== 0) {
    throw new Error('tar extraction failed');
  }
}

function extractZip(archive, destDir) {
  // Use system unzip or PowerShell on Windows
  if (process.platform === 'win32') {
    const ps = spawnSync('powershell', [
      '-NoProfile', '-Command',
      `Expand-Archive -Path "${archive}" -DestinationPath "${destDir}" -Force`,
    ], { stdio: 'inherit' });
    if (ps.status !== 0) throw new Error('zip extraction failed');
  } else {
    const result = spawnSync('unzip', ['-o', archive, '-d', destDir], { stdio: 'inherit' });
    if (result.status !== 0) throw new Error('unzip extraction failed');
  }
}

// ─── main ─────────────────────────────────────────────────────────────────────

async function ensureBinary() {
  const dir = cacheDir();
  const bin = binaryPath(dir);

  if (fs.existsSync(bin)) {
    return bin;
  }

  fs.mkdirSync(dir, { recursive: true });

  const asset = platformAsset();
  const archivePath = path.join(dir, asset.file);

  process.stderr.write(`forge-sdd: downloading v${VERSION} for ${process.platform}/${process.arch}...\n`);

  const assetUrl = `${BASE_URL}/${asset.file}`;
  await download(assetUrl, archivePath);

  // Verify SHA256
  const expectedHash = await verifyChecksum(asset.file, dir);
  const actualHash = sha256(archivePath);
  if (actualHash !== expectedHash) {
    fs.unlinkSync(archivePath);
    throw new Error(`SHA256 mismatch: expected ${expectedHash}, got ${actualHash}`);
  }

  // Extract
  if (asset.ext === 'tar.gz') {
    extractTarGz(archivePath, dir);
  } else {
    extractZip(archivePath, dir);
  }

  // Ensure executable bit on Unix
  if (process.platform !== 'win32') {
    fs.chmodSync(bin, 0o755);
  }

  // Executa a tela de boas-vindas pós-instalação no primeiro download
  try {
    const postinstallPath = path.join(__dirname, 'postinstall.js');
    if (fs.existsSync(postinstallPath)) {
      execFileSync(process.execPath, [postinstallPath], { stdio: 'inherit' });
    }
  } catch (err) {
    // Silencia qualquer erro de renderização para não bloquear o fluxo do CLI
  }

  return bin;
}

ensureBinary()
  .then((bin) => {
    const args = process.argv.slice(2);
    const result = spawnSync(bin, args, { stdio: 'inherit' });
    process.exit(result.status ?? 1);
  })
  .catch((err) => {
    console.error(`forge-sdd error: ${err.message}`);
    process.exit(1);
  });

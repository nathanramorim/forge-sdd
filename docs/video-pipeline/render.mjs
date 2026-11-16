// Renderiza os quadros do vídeo: abre video.html no Chrome headless e captura render(t) a cada 1/fps.
// Uso: node render.mjs <saida-dir> [fps=30] [duracao=80] [t-unico]
//   com t-unico (segundos), salva só esse quadro como <saida-dir>/teste.jpg
import { spawn } from 'node:child_process';
import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const CHROME = process.env.CHROME || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const here = dirname(fileURLToPath(import.meta.url));
const [, , outDir = '/tmp/forge-video-frames', fpsArg = '30', durArg = '80', single] = process.argv;
const fps = Number(fpsArg), dur = Number(durArg), port = 9333;
mkdirSync(outDir, { recursive: true });

const chrome = spawn(CHROME, ['--headless=new', '--disable-gpu', '--no-first-run', '--user-data-dir=/tmp/chp-video',
  '--hide-scrollbars', `--remote-debugging-port=${port}`, '--window-size=1920,1080', 'about:blank'], { stdio: 'ignore' });
const stop = () => { try { chrome.kill(); } catch {} };
process.on('exit', stop);

const sleep = ms => new Promise(r => setTimeout(r, ms));
async function waitJson(path) {
  for (let i = 0; i < 80; i++) { try { const r = await fetch(`http://127.0.0.1:${port}${path}`); if (r.ok) return r.json(); } catch {} await sleep(250); }
  throw new Error('Chrome não respondeu no DevTools');
}
await waitJson('/json/version');
const targets = await waitJson('/json/list');
const page = targets.find(t => t.type === 'page');
const ws = new WebSocket(page.webSocketDebuggerUrl);
await new Promise((ok, no) => { ws.onopen = ok; ws.onerror = no; });
let id = 0; const pending = new Map();
ws.onmessage = ev => { const m = JSON.parse(ev.data); if (m.id && pending.has(m.id)) { pending.get(m.id)(m); pending.delete(m.id); } };
const send = (method, params = {}) => new Promise(res => { const i = ++id; pending.set(i, res); ws.send(JSON.stringify({ id: i, method, params })); });

await send('Page.enable');
await send('Emulation.setDeviceMetricsOverride', { width: 1920, height: 1080, deviceScaleFactor: 1, mobile: false });
await send('Page.navigate', { url: 'file://' + resolve(here, 'video.html') });
await sleep(1500);
await send('Runtime.evaluate', { expression: 'window.__ready', awaitPromise: true });
await sleep(500);

async function frame(t, file) {
  await send('Runtime.evaluate', { expression: `render(${t})` });
  const r = await send('Page.captureScreenshot', { format: 'jpeg', quality: 92 });
  writeFileSync(file, Buffer.from(r.result.data, 'base64'));
}

if (single !== undefined) {
  await frame(Number(single), `${outDir}/teste.jpg`);
  console.log(`quadro t=${single}s salvo em ${outDir}/teste.jpg`);
} else {
  const n = Math.round(fps * dur);
  for (let f = 0; f < n; f++) {
    await frame(f / fps, `${outDir}/${String(f).padStart(5, '0')}.jpg`);
    if (f % 150 === 0) console.log(`quadro ${f}/${n}`);
  }
  console.log(`${n} quadros em ${outDir}`);
}
stop();
process.exit(0);

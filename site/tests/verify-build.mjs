// site/tests/verify-build.mjs
// Suite de testes determinísticos de fidelidade 1:1, SEO, integridade de componentes e Cloudflare Workers
import { readFileSync, existsSync } from 'node:fs';
import { resolve, join } from 'node:path';
import assert from 'node:assert';

const SITE_ROOT = resolve(new URL('.', import.meta.url).pathname, '..');
const DIST_DIR = join(SITE_ROOT, 'dist');
const HTML_FILE = join(DIST_DIR, 'index.html');

console.log('--- Iniciando Verificação de Build do Forge-SDD Site (Astro 5 SSG) ---');

// 1. Verificação de arquivos essenciais gerados no build
console.log('1. Verificando presença dos artefatos estáticos em dist/...');
assert(existsSync(HTML_FILE), 'dist/index.html deve existir');
assert(existsSync(join(DIST_DIR, 'robots.txt')), 'dist/robots.txt deve existir');
assert(existsSync(join(DIST_DIR, 'sitemap-index.xml')), 'dist/sitemap-index.xml deve existir');
assert(existsSync(join(DIST_DIR, 'llms.txt')), 'dist/llms.txt deve existir');
assert(existsSync(join(DIST_DIR, 'favicon.svg')), 'dist/favicon.svg deve existir');
assert(existsSync(join(DIST_DIR, 'og-image.svg')), 'dist/og-image.svg deve existir');
assert(existsSync(join(DIST_DIR, '_headers')), 'dist/_headers deve existir');
assert(existsSync(join(DIST_DIR, 'video/forge-sdd-v3.mp4')), 'dist/video/forge-sdd-v3.mp4 deve existir');
assert(existsSync(join(DIST_DIR, 'video/forge-sdd-v3-poster.jpg')), 'dist/video/forge-sdd-v3-poster.jpg deve existir');

const html = readFileSync(HTML_FILE, 'utf-8');

// 2. SEO Técnico e Metadados (CA-01, CA-03)
console.log('2. Verificando SEO Técnico, Canonical, OpenGraph, Twitter e JSON-LD...');
assert(html.includes('<link rel="canonical" href="https://forge-sdd.org/">'), 'Canonical tag incorreta');
assert(html.includes('property="og:url" content="https://forge-sdd.org/"'), 'og:url incorreta');
assert(html.includes('property="og:title"'), 'og:title ausente');
assert(html.includes('property="og:description"'), 'og:description ausente');
assert(html.includes('property="og:image" content="https://forge-sdd.org/og-image.svg"'), 'og:image incorreta');
assert(html.includes('name="twitter:card" content="summary_large_image"'), 'twitter:card incorreta');
assert(html.includes('application/ld+json'), 'Script JSON-LD ausente');
assert(html.includes('"SoftwareApplication"'), 'Schema SoftwareApplication ausente no JSON-LD');
assert(html.includes('"WebSite"'), 'Schema WebSite ausente no JSON-LD');
assert(html.includes('"FAQPage"'), 'Schema FAQPage ausente no JSON-LD');

// 3. Estrutura das 5 Seções e De-Para 1:1 (CA-04)
console.log('3. Verificando presença e pré-renderização de todas as seções...');
assert(html.includes('id="estacoes"'), 'Seção #estacoes ausente');
assert(html.includes('id="emulador"'), 'Seção #emulador ausente');
assert(html.includes('id="instalacao"'), 'Seção #instalacao ausente');
assert(html.includes('id="governanca"'), 'Seção #governanca ausente');
assert(html.includes('id="evolucao"'), 'Seção #evolucao ausente');
assert(html.includes('id="video"'), 'Seção #video ausente');
assert(html.includes('src="/video/forge-sdd-v3.mp4"'), 'Vídeo /video/forge-sdd-v3.mp4 ausente');

// 4. Invariantes de Design System e Fix 61 (CA-05)
console.log('4. Verificando invariantes de UI, links oficiais e Fix 61...');
// Link do Instagram deve apontar para o perfil oficial do autor @nathan.ramorim
assert(html.includes('href="https://www.instagram.com/nathan.ramorim"'), 'Link oficial do Instagram deve ser https://www.instagram.com/nathan.ramorim');
// Não deve existir menções ao link antigo ou sem ponto
assert(!html.includes('href="https://www.instagram.com/nathanramorim"'), 'Link antigo sem ponto não deve existir');

// Barra viva de handoff e esteira processual
assert(html.includes('id="handoffBar"'), 'Barra #handoffBar ausente');
assert(html.includes('id="hoFrom"'), 'Campo #hoFrom ausente');
assert(html.includes('id="hoTo"'), 'Campo #hoTo ausente');
assert(html.includes('id="hoPayload"'), 'Campo #hoPayload ausente');
assert(html.includes('id="hoGateChip"'), 'Campo #hoGateChip ausente');

// SVG Fallbacks dos mascotes pré-renderizados no HTML inicial (evita CLS: 5 na esteira, 1 no emulador e 1 na governança)
const mascotMatches = html.match(/class="mascot-fallback-svg"/g);
assert(mascotMatches && mascotMatches.length === 7, `Esperado 7 mascotes SVG de fallback pré-renderizados (5 esteira + 1 emulador + 1 governança), encontrados: ${mascotMatches ? mascotMatches.length : 0}`);

// 5. Verificação de Scripts e Bundles (CA-06)
console.log('5. Verificando ausência de scripts CDN externos bloqueantes e presença de bundle Astro...');
assert(!html.includes('cdnjs.cloudflare.com/ajax/libs/three.js'), 'Three.js CDN bloqueante não deve existir');
assert(!html.includes('unpkg.com'), 'Scripts externos não versionados não devem existir');
assert(html.includes('/_astro/'), 'Bundles compilados pelo Vite em /_astro/ devem estar presentes');

// 6. Configuração Cloudflare Workers e _headers (CA-09)
console.log('6. Verificando configurações de borda da Cloudflare...');
const wranglerPath = join(SITE_ROOT, 'wrangler.jsonc');
assert(existsSync(wranglerPath), 'site/wrangler.jsonc deve existir');
const wranglerContent = readFileSync(wranglerPath, 'utf-8');
assert(wranglerContent.includes('"forge-sdd.org"'), 'wrangler.jsonc deve configurar rota forge-sdd.org');
assert(wranglerContent.includes('"directory": "./dist"'), 'wrangler.jsonc deve apontar assets para ./dist');

const headersContent = readFileSync(join(DIST_DIR, '_headers'), 'utf-8');
assert(headersContent.includes('max-age=31536000, immutable'), 'Headers de cache imutável para /_astro/* ausentes');
assert(headersContent.includes('X-Content-Type-Options: nosniff'), 'Header X-Content-Type-Options: nosniff ausente');
assert(headersContent.includes('X-Frame-Options: DENY'), 'Header X-Frame-Options: DENY ausente');

// 7. Verificação de Robots, Sitemap e LLMs.txt (CA-07)
console.log('7. Verificando sitemap, robots e llms.txt...');
const robotsContent = readFileSync(join(DIST_DIR, 'robots.txt'), 'utf-8');
assert(robotsContent.includes('Sitemap: https://forge-sdd.org/sitemap-index.xml'), 'robots.txt deve apontar para o sitemap');
const sitemapContent = readFileSync(join(DIST_DIR, 'sitemap-0.xml'), 'utf-8');
assert(sitemapContent.includes('https://forge-sdd.org'), 'sitemap deve incluir a URL do domínio oficial');
const llmsContent = readFileSync(join(DIST_DIR, 'llms.txt'), 'utf-8');
assert(llmsContent.includes('https://forge-sdd.org'), 'llms.txt deve conter referências ao projeto forge-sdd');

console.log('✓ Todos os testes de build, fidelidade, SEO e Cloudflare Workers passaram com sucesso!');

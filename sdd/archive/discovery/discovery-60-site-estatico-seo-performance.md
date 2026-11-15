# Discovery 60 — Site Estático Oficial (`https://forge-sdd.org`) em Astro com Otimização Máxima de SEO, GEO e Performance (Core Web Vitals)

Este discovery especifica a concepção de produto e arquitetura para transformar a Landing Page homologada (`docs/showcase-v3-lp.html`) em uma **aplicação web estática oficial 1:1 ("De-Para" estritamente fiel)** utilizando **Astro 5 (`output: 'static'`)** isolado na pasta `/site` do repositório, projetada para alcançar **pontuação máxima em SEO (Google Search + mecanismos de IA / GEO)** e **Performance (Core Web Vitals / Lighthouse 100)** em produção no domínio **`https://forge-sdd.org`** via **Cloudflare Workers (Static Assets)**.

> **Decisões Resolvidas no Clarify com o Usuário (`FLX-01` — Gate 1 Consolidado):**
> 1. **Stack SSG Isolada (`/site`):** **Astro 5 (`output: 'static'`)** dentro da subpasta `/site` (`site/package.json`, `site/astro.config.mjs`, `site/src/`). Zero impacto no `go.mod` e no binário Go do `forge-sdd` (`ARQ-04`).
> 2. **Domínio Oficial e Hospedagem:** **`https://forge-sdd.org`** servido na borda via **Cloudflare Workers com Static Assets** (`site/wrangler.jsonc` e `site/public/_headers` com cache imutável e cabeçalhos de segurança).
> 3. **Fidelidade Visual e Funcional 1:1 ("De-Para" Estrito):** Nenhuma alteração no design, cores (`--bg: #0d0f14`, `--bg2: #12151d`, `--bg3: #1a1e29`), tipografia (`Montserrat` + `JetBrains Mono`), cabeçalho sticky com ScrollSpy, ícones vetoriais em linha (`GitHub`, `NPM`, `Instagram`), Hero, Seção 01 (Trilho Processual + 5 Agentes 3D), Seção 02 (Emulador Realista de 5 passos), Seção 03 (Instalação e 8 Comandos em 2 colunas), Seção 04 (Governança em 3 Camadas com 4 cenários) e Seção 05 (Evolução v2.x vs v3.0-beta).
> 4. **Pré-Renderização Estática Integral (SSG) para SEO:** Em `docs/showcase-v3-lp.html`, o trilho (`#stepperTrack`), os 5 cards de agentes (`#agentsGrid`), o estado inicial do emulador e o terminal da governança eram injetados via `innerHTML` no cliente. Em `/site`, **100% desse conteúdo HTML é pré-renderizado estaticamente no build time pelo Astro**, permitindo que o Googlebot e crawlers de IA indexem todo o conteúdo sem depender de execução de JavaScript.
> 5. **Desacoplamento Não-Bloqueante do Three.js (Core Web Vitals):** Remoção do script bloqueante `<script src=".../three.min.js">` do `<head>`. Os módulos necessários de `three` são importados com *tree-shaking* pelo Vite/Astro e inicializados de forma assíncrona (`requestIdleCallback` + `IntersectionObserver`), exibindo um **fallback SVG estático idêntico no HTML inicial** (`LCP < 1.0s`, `TBT = 0ms`, `CLS = 0`) e pausando o loop `requestAnimationFrame` quando o canvas estiver fora da viewport ou com a aba oculta.

---

## 1. Contexto e Problema (Porquê)

- **Limitações de indexação e distribuição de um protótipo HTML único em `docs/`:** O arquivo `docs/showcase-v3-lp.html` consolidou com sucesso a identidade visual e interativa do Forge-SDD 3.0, porém está localizado em uma pasta interna de documentação e apresenta gargalos técnicos para ser o portal público oficial em `https://forge-sdd.org`:
  1. **Conteúdo invisível para crawlers sem JS:** Os 5 cards das estações (`01. O Líder` a `05. Governança`), seus contratos `Recebe (Entrada)` / `Entrega (Saída)` e o trilho processual são gerados exclusivamente em runtime via `document.createElement` e `innerHTML`.
  2. **Script bloqueante de terceiros no `<head>`:** A inclusão síncrona de `three.min.js` (~600 KB minificado via CDN externa) no `<head>` bloqueia a thread principal e atrasa o *First Contentful Paint (FCP)* e o *Largest Contentful Paint (LCP)* em conexões móveis.
  3. **Ausência de metadados de indexação e dados estruturados:** Faltam tags canônicas (`https://forge-sdd.org/`), `meta description`, OpenGraph, Twitter Cards, Schema.org (`JSON-LD` para `SoftwareApplication`, `WebSite` e `HowTo`/`FAQPage`), `sitemap.xml`, `robots.txt` e `llms.txt` (Generative Engine Optimization — GEO).
  4. **Custo contínuo de GPU/CPU:** O loop `requestAnimationFrame` atual renderiza simultaneamente 7 contextos WebGL (5 cards + 1 emulador + 1 governança) a 60fps mesmo quando os elementos estão fora da área visível da página.

---

## 2. Para Quem (Personas e Canais de Descoberta)

1. **Desenvolvedores, Tech Leads e Arquitetos de Software:** Buscam no Google ou em motores de resposta por IA (ChatGPT Search, Perplexity, Claude, Gemini) por termos como *"Spec-Driven Development"*, *"orquestração de agentes IA"*, *"governança de código para Claude Code / Gemini / Copilot"* e encontram `https://forge-sdd.org` com carregamento instantâneo e emulador interativo funcional.
2. **Crawlers de Motores de Busca (`Googlebot`, `Bingbot`) e Indexadores de IA (`OAI-SearchBot`, `ClaudeBot`, `PerplexityBot`):** Recebem HTML estático semântico completo na primeira resposta HTTP (200 OK na borda da Cloudflare), enriquecido por JSON-LD (`SoftwareApplication`) e pelo manifesto `/llms.txt`.
3. **Mantenedores do `forge-sdd`:** Mantêm o site oficial no mesmo repositório em uma pasta `/site` autossuficiente, sem poluir o `go.mod` ou o binário Go, com deploy contínuo para a Cloudflare.

---

## 3. O que já existe (Reuse Check — `DUP-01` / `DUP-02`)

| Ativo Existente | Onde reside | O que já entrega pronto | Decisão (`reusar`, `estender`, `criar`) |
| --- | --- | --- | --- |
| **Landing Page Homologada do Forge-SDD 3.0** | `docs/showcase-v3-lp.html` | Design System completo (`--bg` a `--r-capsule`), tipografia `Montserrat` + `JetBrains Mono`, SVGs inline (`#ico-github`, `#ico-npm`, `#ico-instagram`), textos das 5 seções, dados estruturados em JS (`STATIONS`, `EMU_STEPS`, `GOV_SCENARIOS`, `STATE_DEFS`) e motor 3D (`createForgeMascot3D`). | **`reusar` e `estender`** — portar 1:1 para componentes Astro pré-renderizados em `/site/src/`, mantendo `docs/showcase-v3-lp.html` intacto como referência histórica. |
| **Metadados do Pacote e Repositório** | `package.json` (raiz), `README.md`, `sdd/memory/constitution.md` | Nome oficial `@nathanramorim/forge-sdd`, licença `MIT`, links do GitHub, NPM e descrição canônica dos comandos CLI e prompts. | **`reusar`** — alimentar os schemas `JSON-LD`, meta tags OpenGraph e `/llms.txt`. |
| **Aplicação Estática Astro + Infra Cloudflare** | *Não existe pasta `/site` no repositório* | Build SSG com pré-renderização de HTML, *tree-shaking* do Three.js, geração de sitemap/headers e configuração Cloudflare Workers (`https://forge-sdd.org`). | **`criar` (`/site`)** — criar aplicação Astro isolada em `/site` sem tocar em `cmd/` ou `internal/`. |

---

## 4. Tabela "De-Para" 1:1 Estrita (`docs/showcase-v3-lp.html` → `/site`)

Nenhuma cor, espaçamento, fonte, texto, ícone ou comportamento interativo pode divergir do original. A evolução é estritamente arquitetural (pré-renderização estática + hidratação progressiva):

| Bloco / Elemento | Como está em `docs/showcase-v3-lp.html` (Antes) | Como fica em `/site` com Astro SSG (Depois — 1:1 Visual) | Ganho Direto de SEO & Core Web Vitals |
| --- | --- | --- | --- |
| **Document `<head>` e Fontes** | `<title>` isolado, Google Fonts e `<script src=".../three.min.js">` síncrono bloqueando o parser. | `SeoHead.astro` com `canonical` (`https://forge-sdd.org/`), `description`, OpenGraph, Twitter Card, `JSON-LD` (`SoftwareApplication`, `WebSite`, `FAQPage`), preload de fontes críticas e **zero `<script>` bloqueante**. | **LCP < 1.0s**, **TBT = 0ms**, Rich Snippets no Google e prévia rica em redes sociais. |
| **Defs SVG (`#ico-github`, `#ico-npm`, `#ico-instagram`)** | Bloco `<svg width="0" height="0">` no topo do `<body>`. | Componente `SvgSprite.astro` idêntico no topo do `<body>` (`stroke-width: 1.8`). | Zero requisições HTTP extras para ícones; fidelidade visual 1:1. |
| **Header Fixo (`.site-header`) + ScrollSpy** | HTML estático + `IntersectionObserver` atualizando `.nav-sec.active`. | Componente `Header.astro` com a mesma marcação, chips `v3.0.0-beta` e `MIT License`, links de seção e links sociais (`NPM`, `GitHub`, `Instagram`) + mesmo `IntersectionObserver`. | Navegação semântica indexável (`<header>`, `<nav aria-label="Navegação Principal">`). |
| **Hero (`.hero`) + Cópia `npx`** | `<h1>`, `.lead`, 4 botões CTA e `.install-pill` com botão `#copyHeroBtn`. | Componente `Hero.astro` com texto e classes 100% idênticos + handler leve de clipboard para `#copyHeroBtn`. | Único `<h1>` na página com palavra-chave primária renderizada no primeiro pacote TCP. |
| **Seção 01 (`#estacoes`) — Trilho Processual (`#stepperTrack`) e Handoff (`#handoffBar`)** | `#stepperTrack` vazio no HTML (`<div class="stepper-track" id="stepperTrack"></div>`), preenchido via JS no cliente. | `ProcessSection.astro` renderiza no build time os **5 botões `.step-node`** (`#step-node-0` a `#step-node-4`) e a `#handoffBar` já populada no estado inicial (`01. O Líder` → `02. PM & Arquiteto`). | Googlebot lê os 5 passos do fluxo e os artefatos de saída direto no HTML fonte. |
| **Seção 01 (`#estacoes`) — Grid dos 5 Agentes (`#agentsGrid`) + Palco 3D** | `#agentsGrid` vazio no HTML (`<div class="agents-grid" id="agentsGrid"></div>`), preenchido via JS no cliente. | `ProcessSection.astro` renderiza no build time os **5 `<article class="agent-card">`** completos com `<h3>`, `.trait`, `.agent-desc`, `.io-strip` (`Recebe` / `Entrega`), `.agent-phrase` e `.state-pills`. Cada `.avatar-stage` traz um **SVG estático idêntico ao personagem** que é substituído suavemente pelo `<canvas>` WebGL quando visível. | **100% dos textos dos 5 especialistas indexados pelo Google** e **CLS = 0.00** (dimensões fixas `78x86px`). |
| **Seção 02 (`#emulador`) — Emulador Interativo de 5 Passos** | `#emuLog`, `#emuFileList` e `#emuActionSlot` vazios no HTML, populados por `renderEmuStep(0)` no JS. | `EmulatorSection.astro` pré-renderiza no HTML o estado do Passo 1 (`1. $ forge init`) dentro de `#emuLog`, `#emuFileList` e `#emuActionSlot`, mantendo as 5 abas (`data-emu="0..4"`) e transições idênticas no cliente. | Indexação imediata do exemplo prático de terminal (`forge init`, `/discovery`, `/split-features`, `/proxima-feature`, `/revisar`). |
| **Seção 03 (`#instalacao`) — Instalação & 8 Comandos em 2 Colunas** | HTML estático com `.install-grid` (2 painéis) e `.cmd-ref-grid` (8 `.cmd-item` em 2 colunas). | `InstallSection.astro` preservando exatamente a mesma estrutura, classes e comandos (`npm install -g @nathanramorim/forge-sdd`, `npx @nathanramorim/forge-sdd@latest init`, `forge update` e os 8 cards de comandos). | Estrutura clara de comandos para extração de *Featured Snippets* e *AI Overviews*. |
| **Seção 04 (`#governanca`) — Governança em 3 Camadas (4 Cenários)** | 4 abas `.gov-tab` no HTML, mas `#govTerminal` vazio até `selectGov(0)` rodar no JS. | `GovernanceSection.astro` com as 4 abas `.gov-tab` e o `#govTerminal` já pré-renderizado no Cenário 1 (`SEG-01 · P-13`), além do avatar (`86x94px`) com fallback SVG e interações de correção (`#govBtnFix`, `#govBtnNext`, `#toggleGovCycleBtn`). | Indexação integral dos códigos de regra (`SEG-01`, `DUP-01`, `ARQ-01`, `TST-01`, `DUP-05`) no HTML estático. |
| **Seção 05 (`#evolucao`) + Footer** | Comparativo `Antes (v2.x)` vs `Agora (v3.0-beta)` + `<footer>` com links vetoriais. | `EvolutionSection.astro` e `Footer.astro` 100% idênticos. | Comparativo semântico indexável e links externos com `rel="noopener noreferrer"`. |

---

## 5. Estratégia de SEO Técnico, GEO (`llms.txt`) e Core Web Vitals

### 5.1 SEO On-Page e Dados Estruturados (`https://forge-sdd.org`)
- **URL Canônica e Robots:** `<link rel="canonical" href="https://forge-sdd.org/">` e `<meta name="robots" content="index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1">`.
- **Schema.org JSON-LD Triplo (`application/ld+json`):**
  1. `SoftwareApplication`: descreve o `Forge-SDD` (`applicationCategory: "DeveloperApplication"`, `operatingSystem: "macOS, Linux, Windows"`, `softwareVersion: "3.0.0-beta"`, `license: "https://opensource.org/licenses/MIT"`, `codeRepository: "https://github.com/nathanramorim/forge-sdd"`, `downloadUrl: "https://www.npmjs.com/package/@nathanramorim/forge-sdd"`, `offers: { "@type": "Offer", "price": "0", "priceCurrency": "USD" }`).
  2. `WebSite`: identifica o portal oficial `https://forge-sdd.org`.
  3. `FAQPage` / `HowTo`: estrutura as etapas de instalação (`npm install -g @nathanramorim/forge-sdd` → `forge init`) e o funcionamento das 5 estações para *Rich Results* no Google.
- **GEO (Generative Engine Optimization — `/llms.txt`):**
  - Arquivo estático `https://forge-sdd.org/llms.txt` seguindo o padrão `llms.txt` para que LLMs e buscadores de IA (Perplexity, ChatGPT Search, Claude, Gemini) consumam de forma limpa e concisa a arquitetura de 5 estações, comandos CLI, regras da Constituição e links oficiais do projeto.
- **Sitemap e Robots:**
  - `https://forge-sdd.org/robots.txt` liberando todos os crawlers e apontando para `https://forge-sdd.org/sitemap-index.xml` / `sitemap.xml` e `llms.txt`.

### 5.2 Engenharia de Performance (Metas Core Web Vitals)
- **LCP (Largest Contentful Paint) `< 1.0s`:** O `<h1>` do Hero e toda a dobra inicial dependem apenas do HTML estático (~15 KB gzip) e CSS inline/crítico servido pela CDN da Cloudflare.
- **TBT (Total Blocking Time) `= 0ms` e INP `< 50ms`:**
  - Handlers de clique (copiar comando, trocar estação, trocar aba do emulador, trocar cenário de governança) são registrados imediatamente em um módulo leve (`interactions.ts`) sem aguardar o carregamento do Three.js.
  - O módulo `mascot-3d.ts` (que importa apenas as classes necessárias de `three`: `WebGLRenderer`, `Scene`, `PerspectiveCamera`, `HemisphereLight`, `DirectionalLight`, `Group`, `Mesh`, `CircleGeometry`, `SphereGeometry`, `RingGeometry`, `PlaneGeometry`, `CylinderGeometry`, `BoxGeometry`, `TorusGeometry`, `MeshBasicMaterial`, `MeshStandardMaterial`, `MathUtils`, `Clock`) é carregado de forma assíncrona após o `first paint` (`requestIdleCallback` com fallback `setTimeout`).
- **Controle de Visibilidade WebGL (`IntersectionObserver`):**
  - Cada palco 3D (`.avatar-stage`, `.emu-avatar-mount`, `.gov-avatar-wrap`) possui um flag `isVisible` controlado por `IntersectionObserver`. O `renderer.render(scene, camera)` só é executado no `requestAnimationFrame` para os mascotes que estão efetivamente visíveis na tela no momento, reduzindo em até 70% o uso de GPU/bateria sem afetar a experiência visual.
- **CLS (Cumulative Layout Shift) `= 0.00`:**
  - Todos os containers de avatar possuem dimensões explícitas no CSS (`78x86px`, `70x78px`, `86x94px`) e contêm no HTML estático um placeholder SVG vetorial com as mesmas proporções e cores da estação.

---

## 6. Escopo (In / Out)

### In Scope (Dentro do Escopo)
- Criação do projeto Astro 5 (`output: 'static'`) isolado na pasta `/site` com TypeScript estrito e scripts de build/preview/check.
- Extração dos dados das 5 estações, emulador e governança para módulos tipados em `site/src/data/` consumidos tanto no build SSG (HTML pré-renderizado) quanto na hidratação cliente.
- Componentes Astro modulares reproduzindo 100% do HTML, CSS e interações de `docs/showcase-v3-lp.html`.
- Implementação de SEO técnico completo (`SeoHead.astro`, JSON-LD, OpenGraph, `robots.txt`, `sitemap.xml`, `llms.txt`, `favicon.svg` e `og-image.svg`).
- Otimização do motor 3D (`mascot-3d.ts`) com *tree-shaking* do Three.js, fallback SVG instantâneo e pausa automática fora da viewport.
- Configuração de deploy para Cloudflare Workers com Static Assets (`site/wrangler.jsonc` e `site/public/_headers` para `https://forge-sdd.org`).
- Suite de verificação automatizada do build estático (`npm --prefix site run test` / `astro check` + validação do HTML gerado em `site/dist/index.html`).

### Out of Scope (Fora do Escopo)
- Qualquer alteração visual, de layout, cores, fontes ou textos em relação a `docs/showcase-v3-lp.html`.
- Qualquer alteração no código Go (`cmd/`, `internal/`, `go.mod`) do binário `forge-sdd`.
- SSR dinâmico em servidor Node.js (a aplicação é 100% estática / SSG servida na borda).

---

## 7. Handoff para `/split-features`

Os artefatos completos deste Discovery são:
- [discovery-60-site-estatico-seo-performance.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/discovery-60-site-estatico-seo-performance.md)
- [criteria-60-site-estatico-seo-performance.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/criteria-60-site-estatico-seo-performance.md)
- [plan-60-site-estatico-seo-performance.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/plan-60-site-estatico-seo-performance.md)

**Próximo passo:** submeter ao Gate 1 (`FLX-01`) e executar `/split-features` para fatiar `plan-60` em `sdd/features/feat-60-site-estatico-seo-performance/` e enfileirar nos Builders.

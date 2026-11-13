# feat-60-01 — Scaffold Astro 5 SSG em `/site`, SEO Técnico (`SeoHead`, JSON-LD, Sitemap, `llms.txt`) e Extração de Dados Tipados

**Branch:** `feat/60-site-estatico-seo-performance`
**Fase:** 60-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `done`

## Objetivo

Inicializar o workspace isolado `/site` com **Astro 5 (`output: 'static'`)**, extrair os dados estruturados das 5 estações, emulador e governança de `docs/showcase-v3-lp.html` para módulos TypeScript tipados (`site/src/data/`), portar 100% do CSS global (`site/src/styles/global.css`) incluindo a correção do Fix 61 em `.handoff-bar`, e implementar a camada completa de SEO Técnico e GEO (`SeoHead.astro`, `BaseLayout.astro`, JSON-LD `SoftwareApplication`/`WebSite`/`FAQPage`, `sitemap-index.xml`, `robots.txt`, `llms.txt`, `favicon.svg` e `og-image.svg`) para o domínio oficial **`https://forge-sdd.org`**.

## Modelo do módulo
- **Modelo:** `enxuto` em relação ao binário Go do `forge-sdd` (zero alteração em `cmd/`, `internal/` ou `go.mod`, preservando `ARQ-01` e `ARQ-04`) e `completo` dentro da fronteira isolada `/site` (fundação SSG com dados tipados em `site/src/data/`, estilos globais em `site/src/styles/` e metadados/layouts de SEO em `site/src/components/SeoHead.astro` e `site/src/layouts/BaseLayout.astro`).
- **Justificativa:** Isolar o projeto Astro em `/site` garante que o binário Go permaneça sem dependências externas de runtime (`ARQ-04`), enquanto centraliza os dados (`STATIONS`, `EMU_STEPS`, `GOV_SCENARIOS`, `STATE_DEFS`) em módulos TypeScript únicos (`DUP-04`) que alimentarão tanto a pré-renderização SSG quanto a hidratação cliente.

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Design System (`:root`), estilos globais, correção de layout da `.handoff-bar` (Fix 61), link oficial `https://www.instagram.com/nathan.ramorim` e estruturas de dados (`STATE_DEFS`, `STATIONS`, `EMU_STEPS`, `GOV_SCENARIOS`) | `docs/showcase-v3-lp.html` (e `docs/index.html`) | CSS completo (incluindo o grid responsivo de `.handoff-bar` do Fix 61 sem sobreposição e links para `https://www.instagram.com/nathan.ramorim`) e objetos JS inline nas linhas de script | `reusar` e `estender` | Extrair 1:1 os tokens/regras CSS para `site/src/styles/global.css` e os dados para módulos tipados em `site/src/data/{stations,emulator,governance}.ts` (`DUP-02`, `DUP-04`) |
| Metadados canônicos do pacote NPM, licença e repositório | `package.json`, `README.md`, `sdd/memory/constitution.md` | Pacote `@nathanramorim/forge-sdd`, versão `3.0.0-beta`, licença `MIT`, repositório `https://github.com/nathanramorim/forge-sdd` | `reusar` | Alimentar `SeoHead.astro`, os 3 schemas `JSON-LD` (`SoftwareApplication`, `WebSite`, `FAQPage`) e o manifesto `/llms.txt` |
| Workspace Astro 5 SSG isolado em `/site` | `/site` | Não existe diretório `/site` no repositório | `criar` | Criar `site/package.json`, `site/tsconfig.json`, `site/astro.config.mjs`, `site/public/{robots.txt,llms.txt,favicon.svg,og-image.svg}` e atualizar `.gitignore` |

## Dependências novas
- **No módulo Go (`go.mod`):** Nenhuma (`ARQ-04` preservado).
- **Em `site/package.json` (workspace web estático isolado):**
  - `astro` (`^5.x`, MIT) — compilador SSG (`output: 'static'`).
  - `@astrojs/sitemap` (`^3.x`, MIT) — geração de `sitemap-index.xml` e `sitemap-0.xml` apontando para `https://forge-sdd.org`.
  - `three` (`^0.170.x`, MIT) e `@types/three` (`^0.170.x`, MIT) — declarados no `site/package.json` para bundling modular com *tree-shaking* pelo Vite.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`
- **IDs:** `ARQ-04`, `ARQ-11`, `SEG-04`, `SEG-07`, `DUP-01`, `DUP-02`, `DUP-04`, `COD-05`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `.gitignore`
- `site/package.json`
- `site/tsconfig.json`
- `site/astro.config.mjs`
- `site/src/styles/global.css`
- `site/src/data/stations.ts`
- `site/src/data/emulator.ts`
- `site/src/data/governance.ts`
- `site/src/components/SeoHead.astro`
- `site/src/layouts/BaseLayout.astro`
- `site/src/pages/index.astro`
- `site/public/robots.txt`
- `site/public/llms.txt`
- `site/public/favicon.svg`
- `site/public/og-image.svg`

## Critérios de Aceitação Executáveis

1. **CA-01 (Isolamento Total do Binário Go — `ARQ-04`):**
   ```bash
   git diff --exit-code go.mod go.sum && go test ./... && go vet ./...
   ```
2. **CA-02 (Build Estático Limpo em `/site`):**
   ```bash
   npm --prefix site run build && test -f site/dist/index.html && test -f site/dist/sitemap-index.xml && test -f site/dist/robots.txt && test -f site/dist/llms.txt
   ```
3. **CA-06 (SEO Técnico Completo para `https://forge-sdd.org` no `<head>`):**
   ```bash
   grep -q 'rel="canonical" href="https://forge-sdd.org/"' site/dist/index.html && \
   grep -q 'application/ld+json' site/dist/index.html && \
   grep -q '"SoftwareApplication"' site/dist/index.html && \
   grep -q '"FAQPage"' site/dist/index.html && \
   ! grep -q 'cdnjs.cloudflare.com/ajax/libs/three.js' site/dist/index.html
   ```
4. **CA-07 (Arquivos de Indexação e GEO — `robots.txt`, `sitemap-index.xml`, `llms.txt`):**
   ```bash
   grep -q 'https://forge-sdd.org/sitemap-index.xml' site/dist/robots.txt && \
   grep -q 'https://forge-sdd.org/' site/dist/sitemap-index.xml && \
   grep -q '@nathanramorim/forge-sdd' site/dist/llms.txt
   ```

## Handoff
Desbloqueia `feat-60-02` (Componentes Astro 1:1 pré-renderizados no build com Fallback SVG).

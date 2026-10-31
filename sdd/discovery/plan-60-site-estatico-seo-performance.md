# Plano de Discovery 60 — Site Estático Oficial (`https://forge-sdd.org`) em Astro com Otimização de SEO e Performance

Roadmap técnico e quebra em subtarefas para o `/split-features` dentro da subpasta única `sdd/features/feat-60-site-estatico-seo-performance/` (Regra 13), operando sobre uma **única branch** (`feat/60-site-estatico-seo-performance`, conforme Regra 15 e `FLX-02`).

Respeita a **Rule of 7** (4 subtarefas `P/M` ordenadas de baixo para cima), preserva **zero impacto no binário Go (`ARQ-04`)** e garante a conversão 1:1 ("De-Para" estrito) de `docs/showcase-v3-lp.html` para uma aplicação **Astro 5 (`output: 'static'`)** em `/site`, otimizada para **SEO (Google Search + GEO `/llms.txt`)** e **Core Web Vitals (`LCP < 1.0s`, `TBT = 0ms`, `CLS = 0`)** na **Cloudflare (`https://forge-sdd.org`)**.

---

## 1. Quebra de Tasks / Subfeatures (`P / M / G`)

### Fase 1 — Fundação Astro SSG (`/site`), SEO Técnico (`forge-sdd.org`) e Dados Tipados

- [ ] **`feat-60-01` — Scaffold Astro 5 SSG em `/site`, SEO Técnico (`SeoHead`, JSON-LD, Sitemap, `llms.txt`) e Extração de Dados Tipados** · **Tamanho: `M`**
  - **Depende de:** —
  - **Escopo:**
    1. Inicializar o workspace isolado `/site` (`site/package.json`, `site/tsconfig.json`, `site/astro.config.mjs` com `site: 'https://forge-sdd.org'`, `output: 'static'`, compressão HTML e integração `@astrojs/sitemap`) e adicionar `site/node_modules/`, `site/dist/` e `site/.astro/` ao `.gitignore`.
    2. Extrair de `docs/showcase-v3-lp.html` para módulos TypeScript tipados em `site/src/data/`:
       - `site/src/data/stations.ts` (`STATE_DEFS`, `STATIONS` com as 5 estações, cores `#f2a65a`, `#8aa4ff`, `#c9a7ff`, `#4fd1b5`, `#ff8a5c`, marcas `ring`, `diamond`, `bar`, `wave`, `tri`, contratos `inputArtifact`/`outputArtifact`, handoffs e estados);
       - `site/src/data/emulator.ts` (`EMU_STEPS` com os 5 passos reais do terminal/chat);
       - `site/src/data/governance.ts` (`GOV_SCENARIOS` com os 4 cenários da Governança em 3 Camadas).
    3. Criar `site/src/styles/global.css` contendo 100% dos tokens CSS (`:root`) e regras de estilo de `docs/showcase-v3-lp.html` sem alterar nenhum pixel.
    4. Criar `site/src/components/SeoHead.astro` e `site/src/layouts/BaseLayout.astro` com:
       - Meta tags canônicas (`https://forge-sdd.org/`), `description`, `robots`, `theme-color`, OpenGraph (`pt_BR`), Twitter Card e preload otimizado das fontes `Montserrat` e `JetBrains Mono` (sem nenhum `<script>` bloqueante no `<head>`);
       - Dados estruturados `JSON-LD` (`application/ld+json`) contendo `SoftwareApplication` (`Forge-SDD`, MIT, NPM, GitHub), `WebSite` e `FAQPage`.
    5. Criar os arquivos estáticos de indexação e identidade em `site/public/`: `robots.txt`, `llms.txt` (GEO estruturado para motores de IA), `favicon.svg` e `og-image.svg`.
  - **Skills Reutilizadas:** —
  - **Critério Executável:** `npm --prefix site run build && go test ./... && go vet ./...` (cobrindo `CA-01`, `CA-02`, `CA-06`, `CA-07`).

---

### Fase 2 — Componentes Pré-Renderizados 1:1 (SSG) e Fallbacks SVG Instantâneos (`CLS = 0`)

- [ ] **`feat-60-02` — Componentes Astro 1:1 Pré-Renderizados no Build (Header, Hero, 5 Estações, Emulador, Instalação, Governança, Evolução e Footer) com Fallback SVG** · **Tamanho: `M`**
  - **Depende de:** `feat-60-01`
  - **Escopo:**
    1. Criar `site/src/components/SvgSprite.astro` (`#ico-github`, `#ico-npm`, `#ico-instagram`), `Header.astro` (sticky + links de seção + links sociais), `Hero.astro` (H1, lead, 4 CTAs e barra `npx @nathanramorim/forge-sdd@latest init` com `#copyHeroBtn`) e `Footer.astro`.
    2. Criar `site/src/components/MascotFallbackSvg.astro` que renderiza no HTML estático um SVG vetorial idêntico à silhueta do robô de cada estação (cor da estação, antena com orbe do estado, visor, olhos e marca peitoral `ring`, `diamond`, `bar`, `wave` ou `tri`), garantindo visualização instantânea no `FCP` e `CLS = 0`.
    3. Criar `site/src/components/ProcessSection.astro` (`#estacoes`) **pré-renderizando estaticamente no build time**:
       - Os 5 nós `#step-node-0..4` dentro de `#stepperTrack`;
       - A barra `#handoffBar` populada com a transição inicial (`01. O Líder` → `02. PM & Arquiteto`);
       - Os 5 cards `<article class="agent-card">` dentro de `#agentsGrid` completos com `MascotFallbackSvg`, metadados, descrição, contratos `Recebe (Entrada)` / `Entrega (Saída)`, frase ativa e pílulas de estado (`.state-pill`).
    4. Criar `site/src/components/EmulatorSection.astro` (`#emulador`) pré-renderizando o estado inicial (`Passo 1 de 5` — `$ forge init`) em `#emuLog`, `#emuFileList`, `#emuActionSlot` e `#emuAvatarMount`.
    5. Criar `site/src/components/InstallSection.astro` (`#instalacao` com os 2 painéis de instalação e os 8 comandos em grid de 2 colunas), `GovernanceSection.astro` (`#governanca` com as 4 abas `.gov-tab` e `#govTerminal` pré-renderizado no Cenário 1 `SEG-01 · P-13`) e `EvolutionSection.astro` (`#evolucao` comparando `v2.x` vs `v3.0-beta`), compondo tudo em `site/src/pages/index.astro`.
  - **Skills Reutilizadas:** —
  - **Critério Executável:** `npm --prefix site run build` e verificação de que `site/dist/index.html` contém todos os 5 cards, trilho, emulador e governança já renderizados no HTML estático (`CA-03`, `CA-04`).

---

### Fase 3 — Hidratação Leve Não-Bloqueante e Motor 3D (`three` Tree-Shaken + `IntersectionObserver`)

- [ ] **`feat-60-03` — Hidratação Progressiva de Interações (`interactions.ts`) e Motor WebGL Otimizado (`mascot-3d.ts`) com Tree-Shaking e Pausa fora da Viewport** · **Tamanho: `M`**
  - **Depende de:** `feat-60-02`
  - **Escopo:**
    1. Implementar `site/src/scripts/interactions.ts` (carregado de forma não-bloqueante) acoplando todos os comportamentos interativos 1:1 sobre o DOM pré-renderizado:
       - Botão `#copyHeroBtn` (`navigator.clipboard`);
       - Ciclo automático de 4.8s e navegação manual (`#prevStepBtn`, `#nextStepBtn`, `#toggleCycleBtn`, clique nos `.step-node`, `.agent-card`, `.state-pill` e `.avatar-stage`) na Seção 01;
       - Troca dos 5 passos do Emulador (`.emu-cmd-btn`, `#emuNextStepBtn`, `#emuDynamicBtn`) na Seção 02;
       - Ciclo automático de 6.5s e simulação de correção (`.gov-tab`, `#govBtnFix`, `#govBtnNext`, `#toggleGovCycleBtn`) na Seção 04;
       - `IntersectionObserver` do ScrollSpy atualizando `.nav-sec.active` no Header.
    2. Implementar `site/src/scripts/mascot-3d.ts` importando apenas as classes necessárias de `three` (ESM com *tree-shaking* pelo Vite, sem CDN externa bloqueante):
       - Reproduzir 100% da geometria, iluminação, marcas peitorais (`ring`, `diamond`, `bar`, `wave`, `tri`), olhos (`cap`, `round`, `happy`, `x`), pulo (`triggerHop`), rastreamento de ponteiro (`pointermove` passivo) e animações de `createForgeMascot3D`;
       - Inicializar o WebGL via `requestIdleCallback` (com fallback `setTimeout`) após o *first paint*, substituindo suavemente o SVG de fallback pelo `<canvas>` sem causar *layout shift*;
       - Adicionar `IntersectionObserver` por container de mascote (`isVisible`) + checagem de `document.hidden`, executando `renderer.render(scene, camera)` apenas para os mascotes visíveis na viewport.
  - **Skills Reutilizadas:** —
  - **Critério Executável:** `npm --prefix site run build` gerando bundles JS otimizados em `site/dist/_astro/` sem script bloqueante no `<head>` (`CA-05`, `CA-08`).

---

### Fase 4 — Infraestrutura Cloudflare (`forge-sdd.org`), Testes Automatizados de SEO/De-Para e Governança

- [ ] **`feat-60-04` — Configuração Cloudflare Workers (`wrangler.jsonc` + `_headers`), Suite de Testes de Fidelidade/SEO (`npm test`) e Documentação** · **Tamanho: `P`**
  - **Depende de:** `feat-60-03`
  - **Escopo:**
    1. Criar `site/wrangler.jsonc` configurado para Cloudflare Workers com Static Assets (`directory: "./dist"`) e rotas customizadas para `forge-sdd.org` e `www.forge-sdd.org`.
    2. Criar `site/public/_headers` aplicando `Cache-Control: public, max-age=31536000, immutable` para `/_astro/*` e cabeçalhos de segurança (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy`).
    3. Criar o script de verificação automatizada `site/tests/verify-build.mjs` (executado via `npm --prefix site test`) que inspeciona `site/dist/` e valida programaticamente todos os critérios `CA-02` a `CA-09` (presença de tags SEO/JSON-LD, ausência de script bloqueante no `<head>`, pré-renderização dos 5 cards e trilho no HTML estático, presença de `sitemap-index.xml`, `robots.txt`, `llms.txt`, `_headers` e integridade dos tokens visuais).
    4. Atualizar `sdd/spec/concepts.md` e `README.md` documentando a estrutura de `/site` (`https://forge-sdd.org`) e validar com `go test ./...`, `go vet ./...` e `forge-sdd gate`.
  - **Skills Reutilizadas:** —
  - **Critério Executável:** `npm --prefix site run build && npm --prefix site test && go test ./... && go vet ./... && go run ./cmd/forge-sdd gate` com Exit 0 (`CA-09`, `CA-10`).

---

## 2. Janelas de Paralelismo e Dependências (para a Fila do Builder)

```text
feat-60-01 (Scaffold Astro + SEO Head/JSON-LD + Dados Tipados + global.css)
    │
    ▼
feat-60-02 (Componentes Astro 1:1 Pré-Renderizados no Build + Fallback SVG)
    │
    ▼
feat-60-03 (Hidratação Progressiva interactions.ts + Motor 3D Tree-Shaken + IntersectionObserver)
    │
    ▼
feat-60-04 (Cloudflare wrangler.jsonc/_headers + Suite de Verificação SEO/De-Para + Gate)
```

*(Nota de paralelismo: `feat-60-01` estabelece a estrutura de `/site` e os contratos de dados em `site/src/data/`; em seguida, a camada de marcação estática `feat-60-02` e os módulos de script `feat-60-03` operam sobre diretórios distintos `src/components/` vs `src/scripts/`, fechando com a validação end-to-end e infraestrutura Cloudflare em `feat-60-04`).*

---

## 3. Mapeamento de Rules, IDs da Constituição e Skills por Task

| Task | Tamanho | Arquivos `.agents/rules/` carregados pelo Builder e Governança | IDs da Constituição Verificados | Skills Reutilizadas (`.agents/skills/`) |
| --- | --- | --- | --- | --- |
| **`feat-60-01`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md` | `ARQ-04`, `ARQ-11`, `SEG-04`, `SEG-07`, `DUP-01`, `DUP-02`, `DUP-04`, `COD-05` | — |
| **`feat-60-02`** | `M` | `sdd-reuso-e-duplicacao.md`, `sdd-seguranca.md` | `DUP-02`, `DUP-04`, `SEG-04`, `COD-01`, `COD-04`, `COD-05` | — |
| **`feat-60-03`** | `M` | `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`, `sdd-seguranca.md` | `ARQ-04`, `SEG-04`, `COD-01`, `COD-04`, `COD-05`, `TST-01` | — |
| **`feat-60-04`** | `P` | `sdd-seguranca.md`, `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md` | `SEG-04`, `DUP-01`, `TST-01`, `TST-04`, `TST-05`, `FLX-01..06` | — |

---

## 4. Handoff para o Líder (`/split-features`)

O plano técnico (`plan-60`) está completo, alinhado às decisões confirmadas no Clarify com o usuário e mapeado contra a Constituição e `.agents/rules/`.
Próxima ação do **Líder**: submeter ao Gate 1 (`FLX-01`) e executar `/split-features` para gerar os arquivos de especificação em `sdd/features/feat-60-site-estatico-seo-performance/feat-60-01..04.md` e enfileirá-los para os Builders.

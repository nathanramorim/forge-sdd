# feat-60-02 — Componentes Astro 1:1 Pré-Renderizados no Build (Header, Hero, 5 Estações, Emulador, Instalação, Governança, Evolução e Footer) com Fallback SVG

**Branch:** `feat/60-site-estatico-seo-performance`
**Fase:** 60-02
**Tamanho:** `M`
**Depende de:** `feat-60-01-scaffold-astro-seo-dados-tipados`
**Status:** `done`

## Objetivo

Construir todos os componentes Astro de apresentação em `site/src/components/` reproduzindo 1:1 ("De-Para" estrito) a estrutura DOM, classes, IDs, cores, textos e links de `docs/showcase-v3-lp.html` (incluindo o Fix 61 da `.handoff-bar` e o link `https://www.instagram.com/nathan.ramorim`), garantindo que **100% do conteúdo antes injetado via JS (`#stepperTrack`, `#handoffBar`, `#agentsGrid`, `#emuLog`, `#emuFileList`, `#emuActionSlot` e `#govTerminal`) seja pré-renderizado estaticamente no HTML de build** e que cada palco de mascote exiba o componente `MascotFallbackSvg.astro` (`CLS = 0`).

## Modelo do módulo
- **Modelo:** `completo` dentro da camada de apresentação estática de `/site` (`site/src/components/*.astro` consumindo os contratos tipados de `site/src/data/` e compondo `site/src/pages/index.astro`), e `enxuto` em relação ao binário Go (`zero` alterações fora de `/site`).
- **Justificativa:** Ao pré-renderizar todas as seções em tempo de build pelo Astro, eliminamos os containers vazios que impediam crawlers sem JS (Googlebot, Bingbot, indexadores de IA) de lerem as 5 estações, os artefatos de entrada/saída, o passo a passo do emulador e os cenários da Governança em 3 Camadas.

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Marcação HTML, textos, classes CSS, SVGs inline (`#ico-github`, `#ico-npm`, `#ico-instagram`), `.handoff-bar` corrigida (Fix 61) e link oficial `https://www.instagram.com/nathan.ramorim` | `docs/showcase-v3-lp.html` | HTML estático das seções fixas + templates `innerHTML` dentro das funções `initProcessSection()`, `renderEmuStep()` e `selectGov()` | `reusar` e `estender` | Converter os templates JS de `docs/showcase-v3-lp.html` em expressões JSX/Astro pré-renderizadas no build time mantendo 100% dos mesmos `id`s, classes e textos (`DUP-02`) |
| Dados tipados das 5 estações, emulador e governança + Layout Base | `site/src/data/{stations,emulator,governance}.ts` e `site/src/layouts/BaseLayout.astro` (`feat-60-01`) | Módulos TypeScript exportando `STATE_DEFS`, `STATIONS`, `EMU_STEPS` e `GOV_SCENARIOS` | `reusar` | Importar diretamente nos componentes `.astro` para gerar o HTML inicial sem duplicar strings (`DUP-04`) |
| Fallback SVG vetorial do mascote por estação (`CLS = 0`) | `docs/showcase-v3-lp.html` (geometria `createForgeMascot3D`) | Silhueta do robô (antena com orbe colorido, cabeça, visor, olhos e marca peitoral `ring`, `diamond`, `bar`, `wave`, `tri`) existia apenas em WebGL | `criar` | Criar `site/src/components/MascotFallbackSvg.astro` para renderizar um `<svg class="mascot-fallback-svg">` estático com as cores e marcas exatas de cada estação antes da hidratação WebGL |

## Dependências novas
Nenhuma além das declaradas em `feat-60-01` (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-seguranca.md`
- **IDs:** `DUP-02`, `DUP-04`, `SEG-04`, `COD-01`, `COD-04`, `COD-05`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `site/src/components/SvgSprite.astro`
- `site/src/components/Header.astro`
- `site/src/components/Hero.astro`
- `site/src/components/MascotFallbackSvg.astro`
- `site/src/components/ProcessSection.astro`
- `site/src/components/EmulatorSection.astro`
- `site/src/components/InstallSection.astro`
- `site/src/components/GovernanceSection.astro`
- `site/src/components/EvolutionSection.astro`
- `site/src/components/Footer.astro`
- `site/src/pages/index.astro`

## Critérios de Aceitação Executáveis

1. **CA-03 (Fidelidade Visual e Estrutural 1:1 — "De-Para" Estrito + Fix 61):**
   ```bash
   npm --prefix site run build && \
   grep -q 'id="ico-github"' site/dist/index.html && \
   grep -q 'id="ico-npm"' site/dist/index.html && \
   grep -q 'id="ico-instagram"' site/dist/index.html && \
   grep -q 'https://www.instagram.com/nathan.ramorim' site/dist/index.html && \
   grep -q 'npx @nathanramorim/forge-sdd@latest init' site/dist/index.html
   ```
2. **CA-04 (Pré-Renderização Estática Integral no HTML para Crawlers sem JS):**
   ```bash
   grep -q 'id="step-node-0"' site/dist/index.html && \
   grep -q 'id="step-node-4"' site/dist/index.html && \
   grep -q 'O Líder (Orquestrador)' site/dist/index.html && \
   grep -q 'Analista de Governança' site/dist/index.html && \
   grep -q 'forge init . --agent claude,gemini --name checkout-api' site/dist/index.html && \
   grep -q 'SEG-01 · P-13' site/dist/index.html && \
   grep -q 'mascot-fallback-svg' site/dist/index.html
   ```

## Handoff
Desbloqueia `feat-60-03` (Hidratação progressiva de interações e motor WebGL 3D otimizado).

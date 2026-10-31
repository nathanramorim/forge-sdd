# feat-60-03 — Hidratação Progressiva de Interações (`interactions.ts`) e Motor WebGL Otimizado (`mascot-3d.ts`) com Tree-Shaking e Pausa fora da Viewport

**Branch:** `feat/60-site-estatico-seo-performance`
**Fase:** 60-03
**Tamanho:** `M`
**Depende de:** `feat-60-02-componentes-ssg-pre-renderizados-fallback-svg`
**Status:** `done`

## Objetivo

Desacoplar a interatividade cliente da página em dois módulos TypeScript otimizados em `site/src/scripts/`:
1. `interactions.ts`: acopla imediatamente todos os handlers de UI sobre o DOM pré-renderizado (cópia `npx` no Hero, ciclo automático e seleção das 5 estações/estados na Seção 01, transição dos 5 passos no Emulador da Seção 02, simulação e ciclo dos 4 cenários na Governança da Seção 04 e ScrollSpy no Header) sem aguardar o carregamento do Three.js (`TBT = 0ms`, `INP < 50ms`).
2. `mascot-3d.ts`: importa apenas as classes necessárias de `three` via ESM (com *tree-shaking* pelo Vite, eliminando o script CDN bloqueante do `<head>`), inicializa os mascotes 3D de forma assíncrona (`requestIdleCallback` com fallback `setTimeout`) substituindo suavemente o SVG de fallback (`CLS = 0`), e utiliza `IntersectionObserver` + `document.hidden` para renderizar no loop `requestAnimationFrame` exclusivamente os palcos visíveis na viewport.

## Modelo do módulo
- **Modelo:** `completo` dentro da camada de scripts cliente de `/site` (`site/src/scripts/interactions.ts` e `site/src/scripts/mascot-3d.ts` consumindo `site/src/data/`), e `enxuto` em relação ao binário Go (`ARQ-04`).
- **Justificativa:** Separar a interatividade leve de DOM (`interactions.ts`) do motor WebGL (`mascot-3d.ts`) garante que a página responda ao primeiro clique em milissegundos enquanto o bundle 3D é carregado em *idle time* e pausado automaticamente quando fora da tela.

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Lógica de eventos da LP e motor 3D `createForgeMascot3D` | `docs/showcase-v3-lp.html` (bloco `<script>` final + `.handoff-bar` com Fix 61 e link `https://www.instagram.com/nathan.ramorim`) | Funções `createForgeMascot3D`, `setActiveProcessStep`, `setAgentState`, `renderEmuStep`, `selectGov`, `applyGovFix`, handler de `#copyHeroBtn` e ScrollSpy | `reusar` e `estender` | Portar 1:1 toda a geometria 3D, expressões oculares (`cap`, `round`, `happy`, `x`), pulo (`triggerHop`), marcas peitorais e máquina de estados das seções, adicionando *tree-shaking* ESM de `three`, inicialização via `requestIdleCallback` e controle de visibilidade com `IntersectionObserver` (`DUP-02`) |
| Contratos de dados tipados e DOM pré-renderizado | `site/src/data/*.ts` (`feat-60-01`) e `site/src/components/*.astro` (`feat-60-02`) | Elementos DOM já presentes no HTML inicial com todos os IDs e fallbacks `.mascot-fallback-svg` | `reusar` | Hidratar os elementos já existentes em vez de recriar a grade do zero no carregamento inicial (`DUP-04`) |

## Dependências novas
Nenhuma além de `three` e `@types/three` já declarados em `site/package.json` (`feat-60-01`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-seguranca.md`
- **IDs:** `ARQ-04`, `SEG-04`, `COD-01`, `COD-04`, `COD-05`, `TST-01`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `site/src/scripts/mascot-3d.ts`
- `site/src/scripts/interactions.ts`
- `site/src/pages/index.astro`

## Critérios de Aceitação Executáveis

1. **CA-05 (Remoção de Script Bloqueante e Tree-Shaking do Three.js com Fallback SVG):**
   ```bash
   npm --prefix site run build && \
   ! grep -q 'cdnjs.cloudflare.com/ajax/libs/three.js' site/dist/index.html && \
   ls site/dist/_astro/*.js >/dev/null 2>&1 && \
   grep -q 'mascot-fallback-svg' site/dist/index.html
   ```
2. **CA-08 (Pausa Inteligente de WebGL fora da Viewport via `IntersectionObserver` e `document.hidden`):**
   ```bash
   grep -q 'IntersectionObserver' site/src/scripts/mascot-3d.ts && \
   grep -q 'document.hidden' site/src/scripts/mascot-3d.ts && \
   grep -q 'requestIdleCallback' site/src/scripts/interactions.ts
   ```

## Handoff
Desbloqueia `feat-60-04` (Configuração Cloudflare Workers, headers de cache/segurança e suite automatizada de verificação de SEO e fidelidade 1:1).

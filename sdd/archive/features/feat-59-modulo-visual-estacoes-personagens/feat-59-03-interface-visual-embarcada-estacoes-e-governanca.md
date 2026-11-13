# feat-59-03 — Interface Visual Embarcada (`embed.FS`) com os 5 Personagens das Estações, Trilho de Handoff, Filas Individuais e Inspetor de Governança

**Branch:** `feat/59-modulo-visual-estacoes-personagens`
**Fase:** 59-03
**Tamanho:** `M`
**Depende de:** `feat-59-02-agregador-read-only-e-servidor-http-sse`
**Status:** `done`

## Objetivo

Embutir em `internal/studio/assets/` via `embed.FS` a interface visual completa do **Forge Studio** para o navegador:
1. Renderizar os **5 personagens compactos das estações** (Three.js com fallback SVG 2D automático quando WebGL não estiver disponível):
   - `01. O Líder` (`#f2a65a`, marca `ring`)
   - `02. PM & Arquiteto` (`#8aa4ff`, marca `diamond`)
   - `03. O Planejador` (`#c9a7ff`, marca `bar`)
   - `04. Fila de Builders` (`#4fd1b5`, marca `wave`)
   - `05. Analista de Governança` (`#ff8a5c`, marca `tri`)
2. Exibir o **Trilho Processual de Handoff** com a rota viva e o pacote em trânsito, incluindo o loop de retrabalho `05 ↺ 04` quando há `changes_requested`.
3. Exibir os cards individuais das 5 estações com itens da fila (`sdd/.queue/stations/<station>.json`), estado de animação do mascote, contratos de `Recebe (Entrada)` e `Entrega (Saída)`.
4. Conectar em tempo real via `EventSource('/api/events')` com badge `LIVE` e chaveador para modo `DEMO`.
5. Apresentar o **Inspetor de Governança** (Camadas A, B e C) e a lista de Auto-Skills.
6. Garantir responsividade e boa navegação (`CA-10`) com cabeçalho sticky, empilhamento fluido e ausência de sobreposições de texto.

## Modelo do módulo
- **Modelo:** `completo` dentro dos assets estáticos de `internal/studio/assets/`.
- **Justificativa:** Projeta o estado vivo dos agentes de forma tangível sem qualquer dependência de pacotes npm externos em tempo de execução do binário Go (`ARQ-03`, `ARQ-04`).

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Palco 3D e mascotes | `docs/showcase-v3-lp.html` e `site/src/scripts/mascot-3d.ts` | Motor 3D de mascotes com fallback SVG | `reusar` e `estender` | Portar a lógica e o markup para o frontend do Studio (`DUP-02`) |
| CSS Design System v3 | `docs/showcase-v3-lp.html` | Cores `--bg`, `--bg2`, `--bg3`, tipografia e pílulas | `reusar` | Manter identidade 100% fiel |

## Dependências novas
Nenhuma no binário Go. Assets embutidos diretamente via `//go:embed assets/*` da stdlib (`ARQ-03`).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-seguranca.md`
- **IDs:** `ARQ-01..04`, `DUP-01`, `DUP-02`, `TST-01..05`, `FLX-01..06`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `internal/studio/assets/index.html`
- `internal/studio/assets/studio.css`
- `internal/studio/assets/studio.js`
- `internal/studio/assets.go`
- `internal/studio/assets_test.go`

## Critérios de Aceitação Executáveis
1. **CA-05 & CA-06 (5 Personagens e Trilho de Handoff):**
   ```bash
   go test -v -run TestStudioAssetsContent ./internal/studio/...
   ```
2. **CA-08 & CA-10 (Tempo Real SSE e Responsividade):**
   ```bash
   go test -v -run TestStudioAssetsResponsive ./internal/studio/... && go vet ./...
   ```

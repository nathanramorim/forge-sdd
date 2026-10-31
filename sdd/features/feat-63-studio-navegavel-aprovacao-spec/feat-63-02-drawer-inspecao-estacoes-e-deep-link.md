# feat-63-02 — Componente Station Drawer (`.station-drawer`), Cartões Clicáveis e Deep-Link para IDE Local

**Branch:** `feat/63-studio-navegavel-aprovacao-spec`
**Fase:** 63-02
**Tamanho:** `M`
**Depende de:** `feat-63-01-api-spec-e-sanitizacao`
**Status:** `todo`

## Objetivo

Transformar a exibição estática das estações no Forge Studio em uma experiência interativa e navegável por meio de uma gaveta lateral de inspeção (`.station-drawer`) e suporte a deep-links diretos para editores nativos (VS Code e Cursor):
1. **Navegabilidade de Estação:**
   - Tornar os 5 cartões de estação (`01. Líder`, `02. PM & Arquiteto`, `03. Planejador`, `04. Builder`, `05. Governança`) interativos no grid, adicionando classes de hover, cursor `pointer` e atributos de acessibilidade (`tabindex="0"`, `role="button"`, `aria-label`).
   - Ao clicar em qualquer cartão, abrir suavemente a gaveta lateral direita (`.station-drawer`), mantendo backdrop semitransparente escurecido.
2. **Conteúdo Estruturado da Gaveta Lateral (`.station-drawer`):**
   - **Cabeçalho:** Nome da estação, número, cor temática do mascote, ícone/marca (`ring`, `diamond`, `bar`, `wave`, `tri`) e resumo do papel canônico definido em `.agents/roles/`.
   - **Contratos:** Painel exibindo claramente os contratos `Recebe (Entrada)` e `Entrega (Saída)`.
   - **Fila de Tarefas da Estação:** Lista detalhada de itens carregados do snapshot (`sdd/.queue/stations/<station>.json`), apresentando status, tentativa, id, artefatos associados e as regras constitucionais aplicáveis mapeadas (`ARQ-*`, `SEG-*`, `DUP-*`, `FLX-*`).
   - **Telemetria & Ferramentas:** Lista filtrada das últimas ações e comandos executados pela estação no console de telemetria.
3. **Detalhes da Tarefa & Deep-Link de IDE:**
   - Ao selecionar uma tarefa específica na lista da gaveta, exibir visão detalhada do item com:
     - Botão primário de Deep-Link: `<a href="vscode://file/<abs_path>" class="btn-ide">Abrir no VS Code</a>` para carregamento imediato no editor nativo (`CA-04`).
     - Botão secundário de Deep-Link para Cursor (`cursor://file/<abs_path>`) e botão utilitário para copiar o comando CLI `code <path>`.
     - Botão de ação "Ver Especificação", que direciona a abertura do leitor de spec.
4. **Ergonomia e Responsividade (`CA-10`):**
   - Fechamento imediato da gaveta pelo botão `x`, clique no backdrop ou acionamento da tecla `ESC`.
   - Responsividade completa (largura fluida entre 360px e 540px em telas maiores, ocupando 100% da viewport em telas mobile/split-screen).
   - Preservação estrita das regras de estilo anti-deformação (`white-space: nowrap; flex-shrink: 0` em todos os chips, badges e botões).
5. **Testes Automatizados de Assets (`internal/studio/assets_test.go`):**
   - Validar presença do markup do drawer no `index.html`.
   - Validar classes CSS essenciais em `studio.css` (`.station-drawer`, `.drawer-backdrop`, `.btn-ide`).
   - Validar funções controladoras e handlers no `studio.js` (`openStationDrawer`, `closeStationDrawer`, `selectTaskInDrawer`).

## Modelo do módulo
- **Modelo:** `enxuto` nos assets estáticos de [`internal/studio/assets`](file:///Users/nathanramorim/git/forge-sdd/internal/studio/assets).
- **Justificativa:** Constrói os componentes web puramente em CSS e JS vanilla embutidos via `embed.FS`, sem introduzir frameworks npm adicionais (`ARQ-03`, `ARQ-04`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Cartões de Estação | `internal/studio/assets/studio.js` e `studio.css` | Renderização dinâmica de `.station-card` | `estender` | Adicionar listeners de clique, atributos ARIA e vinculação ao Drawer |
| Layout e Tokens Visuais | `internal/studio/assets/studio.css` | Design system dark theme, variáveis CSS e chips | `estender` | Criar regras do drawer (`.station-drawer`), animações de slide-in e botões de IDE (`.btn-ide`) |
| Ativos Embutidos | `internal/studio/assets.go` e `assets_test.go` | Embed de templates e teste de integridade | `estender` | Adicionar testes de validação dos novos componentes no DOM embutido |

## Dependências novas (`SEG-07`)
Nenhuma dependência externa. JS e CSS vanilla no navegador e Go stdlib (`embed.FS`).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`
- **IDs:** `ARQ-03`, `ARQ-04`, `DUP-02`, `COD-01`, `COD-05`, `TST-01`, `TST-04`, `TST-05`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `internal/studio/assets/index.html`
- `internal/studio/assets/studio.css`
- `internal/studio/assets/studio.js`
- `internal/studio/assets_test.go`

## Critérios de Aceitação Executáveis
1. **CA-01, CA-02 & CA-04 (Drawer de Inspeção, Detalhes de Tarefas e Deep-Link IDE):**
   ```bash
   go test -v -run TestStudioAssetsDrawerAndLinks ./internal/studio/... && go vet ./...
   ```
2. **Validação Geral de Integridade dos Assets e Layout Responsivo:**
   ```bash
   go test -v ./internal/studio/... && go vet ./...
   ```

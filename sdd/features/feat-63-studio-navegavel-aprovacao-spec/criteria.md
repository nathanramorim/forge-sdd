# Critérios Técnicos 63 — Studio Navegável, Visualizador de Spec com Aprovação Humana (FLX-01) e Deep-Link de IDE

Especificação técnica e critérios de aceitação executáveis da funcionalidade `feat-63-studio-navegavel-aprovacao-spec`, em conformidade estrita com a Constituição (`sdd/memory/constitution.md`) e os papéis canônicos de agente (`.agents/roles/`).

---

## 1. Governança de Arquitetura (`ARQ-11`, `DUP-02`, `SEG-07`)

### 1.1 Modelo do módulo (`ARQ-11`)
- **Modelo:** `enxuto`.
- **Justificativa:** A implementação estende os pacotes existentes [`internal/studio`](file:///Users/nathanramorim/git/forge-sdd/internal/studio) e [`internal/queue`](file:///Users/nathanramorim/git/forge-sdd/internal/queue), além dos arquivos web embutidos via `embed.FS` (`internal/studio/assets/`). Não introduz novos daemons, binários auxiliares ou bibliotecas de terceiros no Go (`ARQ-03`, `ARQ-04`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Servidor HTTP e Router | `internal/studio/server.go` | `Server` com ServeMux e assets `embed.FS` | `estender` | Adicionar handlers seguros para `GET /api/spec` e `POST /api/spec/approve` sem quebrar o isolamento de rede |
| Agregador de Estado e DTOs | `internal/studio/state.go` | `StudioSnapshot`, `StationDTO`, `ActiveTaskDTO` | `estender` | Enriquecer `StationItem` e snapshot com metadados de regras mapeadas e identificação de spec aguardando aprovação humana |
| Gestão de Filas e Handoffs | `internal/queue/stations.go` | `LoadStationQueue`, `SaveStationQueue`, `Handoff`, `StatusApproved` | `reusar` | Reusar o motor de transições idempotentes e escrita atômica para registrar a aprovação da spec |
| Portão Camada A e FLX-01 | `internal/gate/gate.go` | Verificação determinística `P-02` (spec existe) e `P-03` (tabela de reuso) | `reusar` | Garantir que a aprovação respeite os invariantes do Portão FLX-01 |
| Estilos e Componentes Web | `internal/studio/assets/studio.css` | Design system dark theme, chips, cards, mascotes | `estender` | Criar estilos para a Gaveta Lateral (`.station-drawer`), Modal/Viewer de Spec (`.spec-viewer-modal`) e botões de ação IDE/Aprovação |
| Lógica do Front-end e SSE | `internal/studio/assets/studio.js` | `initSSE`, `applySnapshot`, controllers 3D | `estender` | Implementar abertura de Drawer ao clicar em cards e tasks, renderizador de Markdown e chamadas `fetch` para leitura e aprovação |

## Dependências novas (`SEG-07`)
- **Nenhuma dependência externa.** Toda a funcionalidade utiliza 100% da biblioteca padrão do Go (`net/http`, `os`, `path/filepath`, `strings`, `encoding/json`, `embed`) e JavaScript vanilla no navegador (`ARQ-03`, `ARQ-04`).

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `enxuto`.
- **Justificativa:** A implementação estende os pacotes existentes [`internal/studio`](file:///Users/nathanramorim/git/forge-sdd/internal/studio) e [`internal/queue`](file:///Users/nathanramorim/git/forge-sdd/internal/queue), além dos arquivos web embutidos via `embed.FS` (`internal/studio/assets/`). Não introduz novos daemons, binários auxiliares ou bibliotecas de terceiros no Go (`ARQ-03`, `ARQ-04`).

---

## 2. Contratos Técnicos de API e Arquitetura

### 2.1 Endpoint `GET /api/spec?path=<path>`
- **Objetivo:** Retorna o conteúdo de uma Spec em formato seguro para renderização no Studio.
- **Validação de Entrada (`SEG-01`, `SEG-02`):**
  - O parâmetro `path` é obrigatório.
  - O caminho é limpo via `filepath.Clean`.
  - Rejeita tentativas de navegação com `..` ou caminhos absolutos fora do diretório do projeto.
  - Exige que o arquivo esteja contido dentro do diretório do projeto e sob pastas de documentação SDD (`sdd/` ou raiz de features).
  - Se o caminho for inválido ou tentar acessar fora do escopo, retorna `400 Bad Request` com erro explicativo.
  - Se o arquivo não existir no disco, retorna `404 Not Found`.
- **Formato de Resposta (JSON):**
  ```json
  {
    "path": "sdd/features/feat-63-studio-navegavel-aprovacao-spec/discovery.md",
    "absolute_path": "/Users/.../forge-sdd/sdd/features/feat-63-studio-navegavel-aprovacao-spec/discovery.md",
    "content": "# Discovery...",
    "status": "pending_review",
    "station": "pm-architect",
    "feature_id": "feat-63-studio-navegavel-aprovacao-spec"
  }
  ```

### 2.2 Endpoint `POST /api/spec/approve`
- **Objetivo:** Registra a aprovação humana explícita no Portão FLX-01 para uma Spec concebida.
- **Entrada (JSON Payload):**
  ```json
  {
    "spec_path": "sdd/features/feat-63-studio-navegavel-aprovacao-spec/discovery.md",
    "station": "pm-architect",
    "task_id": "feat-63"
  }
  ```
- **Processamento:**
  1. Valida o payload de entrada na borda (`SEG-01`).
  2. Verifica se a Spec existe e atende aos requisitos básicos do portão (Camada A `P-02` e `P-03`).
  3. Atualiza o status do item na fila da estação 02 (`pm-architect`) ou fila do Builder para `approved`.
  4. Dispara atualização no `SSEHub`, emitindo evento `event: state` para todos os clientes conectados.
- **Formato de Resposta (JSON):**
  ```json
  {
    "success": true,
    "message": "Spec aprovada com sucesso no Portão FLX-01",
    "gate": "FLX-01",
    "status": "approved",
    "updated_at": "2026-10-31T20:00:00Z"
  }
  ```

### 2.3 Contrato de UI & Deep-Link de IDE
- **Gaveta Lateral de Inspeção (`.station-drawer`):**
  - Painel fixado à direita com largura responsiva (min: 360px, max: 540px), animação suave de entrada/saída (`transform: translateX`).
  - Botão de fechamento acessível via clique no `x`, clique no backdrop escurecido ou tecla `ESC`.
  - Exibe os metadados da estação selecionada, contratos de entrada/saída, lista de ferramentas recentes e lista detalhada de itens da fila.
- **Visualizador de Spec Embutido (`.spec-viewer`):**
  - Carrega a Spec via requisição assíncrona ao endpoint `/api/spec`.
  - Converte Markdown em HTML semântico com títulos, parágrafos, blocos de código com destaque, tabelas alinhadas e checkboxes de checklist.
  - **Botão Deep-Link IDE:** `<a href="vscode://file/<abs_path>" class="btn-ide">Abrir no VS Code</a>` permitindo a abertura direta no editor nativo com um clique, com link alternativo para `cursor://file/<abs_path>`.
  - **Botão de Aprovação (Portão FLX-01):** Quando a spec estiver pendente de revisão, exibe botão destacado **"Aprovar Spec (Portão FLX-01)"** que invoca `POST /api/spec/approve`.
- **Destaque Visual no Handoff e Estação 02:**
  - Quando houver tarefa em `pm-architect` com status `pending_review` ou `approval`:
    - O card da Estação 02 exibe badge pulsante **"Aguardando Revisão Humana (FLX-01)"**.
    - O mascote 3D da Estação 02 assume o estado visual `approval`.
    - No Trilho de Handoff superior, o pacote de transição posiciona-se entre `02 (PM & Arquiteto)` e `03 (Planejador)` com indicação explícita de "Portão FLX-01 Pendente".

---

## 3. Critérios de Aceitação Executáveis (`CA-01` a `CA-06`)

### CA-01: Inspeção e Navegabilidade de Estação via Gaveta Lateral (Drawer)
- **Given** que o servidor `forge-sdd studio` está em execução e exibe os cards das 5 estações no navegador,
- **When** o usuário clica sobre o card de qualquer estação (ex: `02. PM & Arquiteto` ou `04. Fila de Builders`),
- **Then** a interface abre a gaveta lateral de inspeção (`.station-drawer`),
- **And** a gaveta exibe o cabeçalho com nome e papel da estação, os contratos de entrada e saída, as ferramentas ativas e a lista detalhada de tarefas da respectiva fila.

### CA-02: Navegação de Tarefas e Mapeamento de Regras da Constituição
- **Given** que a gaveta lateral de inspeção está aberta exibindo as tarefas da estação,
- **When** o usuário clica em qualquer item de tarefa na lista,
- **Then** a gaveta exibe o detalhamento completo da tarefa contendo: ID, Título, Status, Chave de Idempotência, Número da Tentativa e as regras constitucionais aplicáveis associadas (`ARQ-*`, `SEG-*`, `DUP-*`, `FLX-*`),
- **And** disponibiliza o botão "Ver Especificação" com o caminho do artefato correspondente.

### CA-03: Endpoint Seguro de Leitura de Spec e Prevenção de Path Traversal
- **Given** o servidor HTTP do Studio em execução,
- **When** uma requisição válida `GET /api/spec?path=sdd/features/feat-63-studio-navegavel-aprovacao-spec/discovery.md` é enviada,
- **Then** o endpoint retorna status HTTP `200 OK` e um payload JSON contendo os campos `content`, `path`, `absolute_path` e `status`;
- **And When** uma requisição maliciosa com tentativa de evasão de diretório como `GET /api/spec?path=../../etc/passwd` ou caminho externo é enviada,
- **Then** o endpoint rejeita a requisição retornando status HTTP `400 Bad Request` ou `403 Forbidden` (`SEG-01`, `SEG-02`, `SEG-04`).

### CA-04: Visualizador de Spec Embutido e Deep-Link Direto para IDE
- **Given** que o usuário aciona o botão "Ver Especificação" em uma tarefa ou na Estação 02,
- **When** o visualizador de Spec é carregado no Studio,
- **Then** o conteúdo Markdown da Spec é renderizado com formatação visual legível contendo cabeçalhos, tabelas de reuso, blocos de código e checklist executável,
- **And** a interface apresenta um botão de Deep-Link contendo a URI nativa `vscode://file/<abs_path>` (e suporte a `cursor://file/<abs_path>`), abrindo o arquivo diretamente na IDE local ao ser clicado.

### CA-05: Painel de Aprovação Humana da Spec (Portão FLX-01)
- **Given** que uma Spec foi elaborada e aguarda a revisão humana obrigatória do Portão FLX-01,
- **When** o usuário visualiza a Estação 02 ou o Trilho de Handoff no Studio,
- **Then** a Estação 02 exibe com destaque visual o badge pulsante "Aguardando Revisão Humana (FLX-01)" e o mascote 3D em estado `approval`;
- **And When** o usuário clica no botão "Aprovar Spec (FLX-01)" dentro do visualizador,
- **Then** uma requisição `POST /api/spec/approve` é disparada, a fila persiste o estado `approved`, o Portão FLX-01 é marcado como superado e a fila é liberada para a próxima estação.

### CA-06: Sincronização e Reflexo Instantâneo em Tempo Real via SSE
- **Given** que uma ou mais instâncias do navegador estão conectadas ao Studio através do endpoint `/api/events`,
- **When** uma Spec é aprovada, uma tarefa é atualizada ou um comando é registrado nas filas do filesystem,
- **Then** o SSE Hub detecta a modificação e transmite imediatamente o novo snapshot consolidado via SSE em menos de `500ms`,
- **And** a interface web atualiza o status das estações, o Drawer de inspeção e o Trilho de Handoff em tempo real, sem necessidade de recarregar a página (`F5`).

---

## 4. Estratégia de Testes Automatizados (`TST-01..05`)

1. **Testes Unitários da API de Spec (`internal/studio/server_test.go`):**
   - Teste de leitura de spec legítima (`sdd/features/...`) retornando `200 OK` e JSON com conteúdo íntegro.
   - Teste de segurança contra *Path Traversal* (`../`, `/etc/passwd`, caminhos fora do repositório) comprovando rejeição imediata com `400` ou `403` (`SEG-01`, `SEG-02`).
   - Teste do endpoint `POST /api/spec/approve` validando atualização de fila, idempotência de aprovação e retorno determinístico.
2. **Testes de Integração SSE:**
   - Teste validando que a aprovação via API dispara notificação no `SSEHub` e entrega o snapshot atualizado para clientes inscritos.
3. **Verificação de Regras da Constituição:**
   - Execução integral de `go test -v ./internal/studio/...` e `go test -v ./internal/queue/...`.
   - Execução de `go vet ./...` sem nenhum alerta (`COD-02`).

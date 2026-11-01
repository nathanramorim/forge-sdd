---
description: Como operar, estender e integrar o Protocolo ACP (Agent Client Protocol) para comunicação entre IDEs, Studio e as 5 estações SDD.
globs: ["internal/acp/**", "cmd/forge-sdd/acp.go", "docs/acp-*.md"]
regras: [ARQ-01, ARQ-06, SEG-01, SEG-02, SEG-04, COD-01, COD-03, FLX-01, FLX-02, FLX-03, GOV-04, TST-01, TST-05]
---

# Guia Prático — Protocolo ACP (Agent Client Protocol)

Procedimentos operacionais para implementar, operar e estender o Protocolo ACP no forge-sdd, atendendo aos IDs `ARQ-01`, `ARQ-06`, `SEG-01`, `SEG-02`, `SEG-04`, `COD-01`, `COD-03`, `FLX-01`, `FLX-02`, `FLX-03`, `GOV-04`, `TST-01` e `TST-05` de `sdd/memory/constitution.md`.

## 1. Fronteiras Arquiteturais e Responsabilidades (`ARQ-01`, `ARQ-06`, `COD-01`)

- **Separação `cmd/` vs `internal/` (`ARQ-01`):**
  - O ponto de entrada CLI para o protocolo reside estritamente em `cmd/forge-sdd/acp.go`. Ele inicializa o despachador, orquestra os sinais de sistema operacional (`os.Interrupt`, `SIGTERM`) e delega a execução ao transporte.
  - Toda a lógica de protocolo, despacho de chamadas, controle de sessões, validações de workspace e regras de negócio reside no pacote `internal/acp`.
  - O pacote `internal/acp` jamais importa `cmd/` ou depende de bibliotecas externas, operando exclusivamente sobre a `stdlib` do Go.
- **Superfície da CLI (`ARQ-06`):**
  - O subcomando `forge-sdd acp` expõe o servidor ACP sobre `stdio` (linha de comando / IPC local).
  - Flags expostas (como `--dir`) devem ser validadas na inicialização e repassadas ao `SessionManager` como raiz segura de trabalho.

## 2. Envelopes JSON-RPC 2.0 e Tratamento de Erros (`COD-01`, `COD-03`)

- **Estruturas Canônicas:**
  - Todas as mensagens trocadas seguem o padrão JSON-RPC 2.0 (`jsonrpc: "2.0"`).
  - `Request`: deve conter campos `jsonrpc`, `id` (inteiro ou string) e `method`. Parâmetros opcionais em `params`.
  - `Response`: conformidade estrita com exclusão mútua (`MarshalJSON` personalizado). Quando `error` estiver presente, `result` não deve ser emitido; quando for bem-sucedido, `result` é emitido e `error` omitido.
  - `Notification`: mensagens unidirecionais sem campo `id` (`method: "session/notification"`).
- **Códigos Padronizados de Erro:**
  - `-32700` (`CodeParseError`): falha de sintaxe ou parsing do JSON recebido.
  - `-32600` (`CodeInvalidRequest`): envelope JSON inválido (campo `jsonrpc` incorreto ou ausência de método).
  - `-32601` (`CodeMethodNotFound`): método invocado não está registrado no despachador.
  - `-32602` (`CodeInvalidParams`): argumentos ausentes ou inválidos (ex: workspace inexistente, `sessionId` vazio, `specPath` nulo).
  - `-32603` (`CodeInternalError`): falhas inesperadas internas de runtime ou I/O.
  - `-32001` (`CodeServerError`): erros de domínio e rejeição de regras de governança (ex: violação de critérios do Portão FLX-01 ou tentativa de operar em sessão cancelada).

## 3. Ciclo de Vida de Sessões e Isolamento de Workspace (`SEG-01`, `SEG-02`, `SEG-04`)

- **Handshake Inicial (`initialize`):**
  - Todo cliente deve iniciar a comunicação enviando a chamada `initialize`.
  - O servidor responde informando a versão de protocolo (`protocolVersion: "1.0.0"`), dados do servidor (`ServerInfo`) e capacidades suportadas (`Capabilities: { sessions: true, gates: true, prompts: true }`).
- **Criação de Sessão (`session/new`):**
  - Cria uma nova instância de `Session` com identificador único determinístico (`session-<timestamp>`).
  - O caminho do workspace informado em `workspaceDir` é sanitizado via `filepath.Clean` e validado contra path traversal (`SEG-01`, `SEG-02`):
    - Rejeição imediata de strings contendo caracteres nulos (`\0`) ou de controle.
    - Se uma raiz permitida (`allowedRoot`) estiver configurada, o caminho deve estar contido sob ela.
    - O diretório físico deve existir e ser confirmado via `os.Stat`.
- **Recuperação e Cancelamento (`session/load`, `session/cancel`):**
  - `session/load`: recupera metadados de uma sessão ativa existente.
  - `session/cancel`: marca o estado da sessão como `cancelled`, impedindo novos despachos ou comandos no pipeline.
- **Proteção de Segredos (`SEG-04`):**
  - Nenhum token, credencial, segredo de API ou chave privada deve ser transmitido em payloads de sessão ou parâmetros de método.

## 4. Envio de Prompts e Integração com Filas SDD (`FLX-01`, `COD-01`)

- **Despacho via `session/prompt`:**
  - Recebe `sessionId`, `prompt`, `station` (opcional, padrão: `concepcao` / `StationPMArchitect`), `operation` e metadados.
  - Valida a existência da sessão ativa e o nome da estação contra o enum canônico de estações (`queue.IsValidStation`).
  - Enfileira a demanda diretamente na fila determinística correspondente via `queue.EnqueueStationItem`, garantindo atomicidade na gravação do arquivo `sdd/.queue/stations/<station>.json`.
  - Emite automaticamente uma notificação `prompt_enqueued` para os assinantes conectados.

## 5. Notificações de Streaming Assíncrono (`session/notification`)

- **Canal Unidirecional de Streaming:**
  - O despachador mantém inscrições thread-safe (`SubscribeNotifications`, `UnsubscribeNotifications`, `BroadcastSessionNotification`).
  - As notificações são transmitidas sem bloqueio para todos os canais de saída ativos (linha stdio ou conexão HTTP SSE no Studio).
  - Tipos canônicos de eventos emitidos:
    - `prompt_enqueued`: confirmação de enfileiramento de item.
    - `station_state`: alteração de status operacional de uma estação.
    - `mascot_state`: atualização de estado visual e animação do personagem 3D correspondente.
    - `tool_call`: telemetria de ferramenta executada por agente.
    - `gate_approved`: homologação e handoff de portão de governança.

## 6. Homologação do Portão FLX-01 e Handoff Determinístico (`FLX-01`, `FLX-02`, `FLX-03`)

- **Execução do Portão via `gate/approve`:**
  - O cliente externo (IDE ou Studio) solicita a aprovação do Portão FLX-01 fornecendo `specPath`, `sessionId` ou `workspaceDir`, e a branch associada.
  - O método aciona a verificação canônica em `internal/gate` (`gate.ValidateFLX01WithBranch`), reaproveitando integralmente os checks `P-01`, `P-02` e `P-03`.
  - Se algum critério falhar (ex: branch apontando para `main`, arquivo de spec inexistente, ou tabela de reuso ausente), o servidor responde com erro `-32001` acompanhado do detalhamento das evidências violadas.
  - Em caso de aprovação:
    1. A fila da Estação 02 (`concepcao`) é atualizada para `approved`.
    2. O handoff atômico é executado via `queue.Handoff`, transferindo o artefato para a Estação 03 (`planejamento`).
    3. Notificação assíncrona `gate_approved` é propagada pelo despachador.

## 7. Garantias de Testes, Concorrência e Qualidade (`TST-01`, `TST-05`)

- **Testes Unitários e E2E:**
  - Toda a suíte em `internal/acp/` opera em isolamento com diretórios temporários (`t.TempDir()`), sem tocar em arquivos globais do usuário.
  - A integridade sob concorrência e ausência de deadlocks deve ser validada executando `go test -v -race ./internal/acp/...`.
  - O teste E2E (`e2e_test.go`) valida a jornada completa de uma IDE real através de pipes de I/O (`io.Pipe`).
- **Verificação Contínua:**
  - Execução mandatória de `go vet ./...` sem advertências após qualquer alteração no protocolo ACP.

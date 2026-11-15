# feat-64-02 — Integração de Sessões ACP com Motor de Filas (`internal/queue`) e Aprovação de Portão (`gate/approve`)

**Branch:** `feat/64-protocolo-acp-comunicacao-estacoes-studio`
**Fase:** 64-02
**Tamanho:** `M`
**Depende de:** `feat-64-01`
**Status:** `ready`

## Objetivo

Conectar o protocolo ACP ao ciclo de vida das 5 estações do `forge-sdd`, permitindo controle de sessões e resolução de portão de forma determinística:
1. Implementar o [`SessionManager`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/session.go) em `internal/acp/session.go` para gerenciar sessões ativas ancoradas no workspace local com controle de concorrência (`sync.RWMutex`).
2. Implementar o método RPC `session/prompt`: receber demandas textuais ou comandos de clientes externos (IDEs ou Studio), validar parâmetros (`SEG-01`) e despachar deterministicamente para a fila da estação ativa via [`internal/queue/stations.go`](file:///Users/nathanramorim/git/forge-sdd/internal/queue/stations.go) (`EnqueueStationItem`).
3. Implementar o método RPC `gate/approve`: viabilizar a aprovação do Portão de Governança FLX-01 via ACP. O método invoca o validador canônico em [`internal/gate/gate.go`](file:///Users/nathanramorim/git/forge-sdd/internal/gate/gate.go) (`P-02`, `P-03`), rejeita com erro `-32001` caso critérios sejam violados, e em caso de sucesso realiza escrita atômica transicionando a fila da Estação 02 para `approved` e acionando o handoff determinístico (`FLX-01`, `DUP-01`).
4. Desenvolver o canal de streaming de notificações unidirecionais (`session/notification`) propagando alterações de estado das estações, mascotes ativos e chamadas de ferramentas.
5. Criar testes de integração cobrindo fluxos de sucesso, tentativas de acesso com caminhos inválidos fora do workspace (`SEG-02`) e rejeições de aprovação com erro semântico `-32001`.

## Modelo do módulo
- **Modelo:** `enxuto` dentro de [`internal/acp`](file:///Users/nathanramorim/git/forge-sdd/internal/acp).
- **Justificativa:** Age como camada orquestradora entre o envelope ACP e os módulos canônicos `internal/queue` e `internal/gate`. Não recria lógica de persistência nem duplica estruturas de fila (`ARQ-01`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Motor de Filas das 5 Estações | [`internal/queue/stations.go`](file:///Users/nathanramorim/git/forge-sdd/internal/queue/stations.go) | `EnqueueStationItem`, `Handoff`, escrita atômica | `reusar` | O ACP não altera os arquivos de fila diretamente; ele despacha chamadas exclusivamente através das funções públicas de `internal/queue` (`DUP-01`). |
| Validador de Portão Camada A | [`internal/gate/gate.go`](file:///Users/nathanramorim/git/forge-sdd/internal/gate/gate.go) | Verificações de portão `P-01` a `P-05` | `reusar` | `gate/approve` reutiliza o validador determinístico já homologado no forge-sdd. |
| Memória e Progresso | [`sdd/memory/progress.md`](file:///Users/nathanramorim/git/forge-sdd/sdd/memory/progress.md) | Arquivo de estado ativo do projeto | `reusar` | O status ativo da sessão reflete e sincroniza com o progresso do projeto. |

## Dependências novas (`SEG-07`)

- **Nenhuma dependência externa.**
- Implementação estrita com 100% Go stdlib (`path/filepath`, `strings`, `sync`, `time`, `os`, `fmt`, `context`).

## Rules e IDs da Constituição Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-fluxo-e-portao.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-02`, `ARQ-04`, `ARQ-11`, `SEG-01`, `SEG-02`, `SEG-04`, `DUP-01`, `DUP-02`, `FLX-01`, `FLX-04`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-02`, `TST-03`, `TST-04`, `TST-05`.
- **Skills Reutilizadas:** —

## Arquivos Tocados

- `internal/acp/session.go` (criar)
- `internal/acp/gate.go` (criar)
- `internal/acp/session_test.go` (criar)
- `internal/acp/dispatcher.go` (estender com handlers de sessão e portão)

## Critérios de Aceitação Executáveis

1. **CA-01 (Ciclo de Vida de Sessões e Isolamento de Workspace):**
   ```bash
   go test -v -race -run "TestACPSessionLifecycle" ./internal/acp/... && go vet ./...
   ```
2. **CA-02 (Aprovação de Portão FLX-01 e Rejeição com Código -32001):**
   ```bash
   go test -v -race -run "TestACPGateApproval" ./internal/acp/... ./internal/queue/... && go vet ./...
   ```
3. **CA-03 (Validação Integrada de Concorrência e Filas):**
   ```bash
   go test -v -race -run "TestACPSessionAndGate" ./internal/acp/... ./internal/queue/... && go vet ./...
   ```

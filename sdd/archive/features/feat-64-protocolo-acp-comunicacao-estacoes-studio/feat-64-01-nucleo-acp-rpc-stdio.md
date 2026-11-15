# feat-64-01 — Núcleo do Protocolo ACP (`internal/acp`), Tipos JSON-RPC 2.0 e Subcomando CLI `forge-sdd acp`

**Branch:** `feat/64-protocolo-acp-comunicacao-estacoes-studio`
**Fase:** 64-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `ready`

## Objetivo

Implementar a fundação do **Protocolo ACP (Agent Client Protocol)** em Go stdlib puro no repositório `forge-sdd`:
1. Definir o envelope canônico e os tipos JSON-RPC 2.0 em [`internal/acp/types.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/types.go): `Request`, `Response`, `Error`, `Notification` e os códigos de erro padronizados (`-32700`, `-32600`, `-32601`, `-32602`, `-32603`, `-32001`).
2. Implementar o [`RPCDispatcher`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/dispatcher.go) para registro e despacho thread-safe de métodos (`initialize`, `session/new`, `session/load`, `session/cancel`), suportando handlers síncronos e assíncronos.
3. Desenvolver o [`StdioTransport`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/transport_stdio.go) para comunicação via stream de entrada/saída padrão (`os.Stdin`/`os.Stdout`), utilizando `bufio.Scanner` e serialização protegida por `sync.Mutex` (`COD-01`).
4. Criar o subcomando CLI [`cmd/forge-sdd/acp.go`](file:///Users/nathanramorim/git/forge-sdd/cmd/forge-sdd/acp.go) (`forge-sdd acp`) conforme o procedimento padrão da skill `novo-subcomando-cli-sdd` (`DUP-05`), mantendo as fronteiras arquiteturais intactas (`ARQ-01`, `ARQ-06`).
5. Criar suíte de testes unitários com concorrência segura (`go test -race`) em [`internal/acp/acp_test.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/acp_test.go) cobrindo parsing de mensagens JSON-RPC, handshake de inicialização e tratamento de entradas malformadas.

## Modelo do módulo
- **Modelo:** `enxuto` dentro de [`internal/acp`](file:///Users/nathanramorim/git/forge-sdd/internal/acp).
- **Justificativa:** O módulo atua exclusivamente como adaptador de comunicação e roteamento JSON-RPC 2.0. Não introduz daemons pesados, subprocessos em background ou persistência paralela, comunicando-se diretamente com os pacotes internos já existentes (`ARQ-03`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Comandos da CLI | [`cmd/forge-sdd/`](file:///Users/nathanramorim/git/forge-sdd/cmd/forge-sdd) | Subcomandos Cobra (`gate.go`, `studio.go`, `queue.go`) | `estender` | Adicionar `cmd/forge-sdd/acp.go` seguindo o padrão estabelecido na skill `novo-subcomando-cli-sdd` (`DUP-05`). |
| Streams de I/O e JSON | Go Standard Library | `encoding/json`, `bufio`, `io`, `os`, `sync` | `reusar` | Implementação do transporte stdio e serialização de RPC 100% nativa em Go stdlib sem bibliotecas externas (`ARQ-04`, `SEG-07`). |
| Protocolo ACP e Tipos RPC | Repositório forge-sdd | Inexistente | `criar` | Criar o pacote `internal/acp` para abrigar a especificação canônica do protocolo ACP no projeto. |

## Dependências novas (`SEG-07`)

- **Nenhuma dependência externa.**
- Implementação estrita com 100% Go stdlib (`bufio`, `context`, `encoding/json`, `fmt`, `io`, `os`, `sync`).

## Rules e IDs da Constituição Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`.
- **IDs:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-06`, `ARQ-11`, `SEG-01`, `SEG-04`, `SEG-07`, `DUP-02`, `DUP-05`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-03`, `TST-04`, `TST-05`.
- **Skills Reutilizadas:** `novo-subcomando-cli-sdd`.

## Arquivos Tocados

- `internal/acp/types.go` (criar)
- `internal/acp/dispatcher.go` (criar)
- `internal/acp/transport_stdio.go` (criar)
- `internal/acp/acp_test.go` (criar)
- `cmd/forge-sdd/acp.go` (criar)
- `cmd/forge-sdd/main.go` (estender para registrar subcomando `acpCmd`)

## Critérios de Aceitação Executáveis

1. **CA-01 (Handshake `initialize` e Parsing JSON-RPC 2.0 via Stdio):**
   ```bash
   go test -v -race -run "TestACPInitialize" ./internal/acp/... && go vet ./...
   ```
2. **CA-02 (Tratamento de Mensagens Malformadas e Códigos de Erro):**
   ```bash
   go test -v -race -run "TestACPErrors" ./internal/acp/... && go vet ./...
   ```
3. **CA-03 (Execução CLI sem Pânico e Validação Estática):**
   ```bash
   go test -v -race ./internal/acp/... ./cmd/forge-sdd/... && go vet ./...
   ```

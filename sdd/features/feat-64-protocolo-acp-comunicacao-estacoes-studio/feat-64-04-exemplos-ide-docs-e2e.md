# feat-64-04 — Exemplos de Integração com IDEs (Zed, Cursor, VS Code), Documentação e Testes E2E

**Branch:** `feat/64-protocolo-acp-comunicacao-estacoes-studio`
**Fase:** 64-04
**Tamanho:** `P`
**Depende de:** `feat-64-01`, `feat-64-02`, `feat-64-03`
**Status:** `done`

## Objetivo

Documentar a integração do Protocolo ACP, fornecer arquivos de exemplo para IDEs populares de mercado e validar a esteira completa através de teste end-to-end automatizado:
1. Criar a documentação canônica de regras em [`.agents/rules/sdd-acp.md`](file:///Users/nathanramorim/git/forge-sdd/.agents/rules/sdd-acp.md) detalhando a especificação do protocolo, lifecycle das sessões, envelopes JSON-RPC 2.0 e tratamento de erros.
2. Criar guia prático de configuração para desenvolvedores em [`docs/acp-ide-setup.md`](file:///Users/nathanramorim/git/forge-sdd/docs/acp-ide-setup.md) contendo:
   - Configuração para o **Zed Editor** (`agent_servers` / ACP adapter em `settings.json`).
   - Snippet de configuração para extensões ACP no **Cursor** e **VS Code** via `stdio` bridge (`forge-sdd acp`).
3. Desenvolver teste E2E determinístico em [`internal/acp/e2e_test.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/e2e_test.go) simulando a jornada completa de um cliente IDE externo operando sobre stdio:
   - Handshake `initialize`
   - Abertura de workspace com `session/new`
   - Envio de demanda com `session/prompt`
   - Recepção de notificações assíncronas de atividade
   - Resolução e aprovação do Portão FLX-01 via `gate/approve`
   - Finalização da sessão
4. Realizar verificação integral de qualidade: zero avisos em `go vet ./...`, execução completa dos testes com flag de detecção de corrida (`-race`) e conformidade com `sdd/memory/constitution.md`.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Produção de documentação técnica, regras para agentes e suíte de testes E2E executada sob demanda sem introduzir dependências de runtime (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Regras de Agentes | [`.agents/rules/`](file:///Users/nathanramorim/git/forge-sdd/.agents/rules) | `sdd-arquitetura.md`, `sdd-fluxo-e-portao.md` | `estender` | Adicionar `.agents/rules/sdd-acp.md` como guia canônico de conformidade com o protocolo ACP. |
| Documentação e Guias | [`docs/`](file:///Users/nathanramorim/git/forge-sdd/docs) | Guias de desenvolvimento e arquitetura | `estender` | Adicionar `docs/acp-ide-setup.md` com instruções detalhadas para configuração de IDEs. |
| Testes Unitários ACP | [`internal/acp/`](file:///Users/nathanramorim/git/forge-sdd/internal/acp) | Pacote criado nas fases 64-01 e 64-02 | `estender` | Adicionar teste E2E integrando transporte stdio, filas e portões. |

## Dependências novas (`SEG-07`)

- **Nenhuma dependência externa.**
- Implementação estrita com 100% Go stdlib.

## Rules e IDs da Constituição Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-governanca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-06`, `SEG-04`, `COD-01`, `COD-05`, `FLX-02`, `FLX-03`, `GOV-04`, `TST-01`, `TST-05`.
- **Skills Reutilizadas:** —

## Arquivos Tocados

- `.agents/rules/sdd-acp.md` (criar)
- `docs/acp-ide-setup.md` (criar)
- `internal/acp/e2e_test.go` (criar)
- `README.md` (estender com referência ao subcomando `forge-sdd acp`)

## Critérios de Aceitação Executáveis

1. **CA-01 (Validação do Teste E2E do Protocolo ACP):**
   ```bash
   go test -v -race -run "TestACPEndToEnd" ./internal/acp/... && go vet ./...
   ```
2. **CA-02 (Conformidade Global do Repositório e Zero Dependências):**
   ```bash
   go vet ./... && go test -v -race ./internal/acp/... ./internal/studio/... ./internal/queue/...
   ```

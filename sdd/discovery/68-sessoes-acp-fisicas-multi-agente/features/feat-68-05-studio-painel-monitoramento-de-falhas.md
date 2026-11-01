# feat-68-05 — Studio: Painel de Monitoramento de Falhas das Estações

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-05
**Tamanho:** `M`
**Depende de:** `feat-68-03b`, `feat-68-04`
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Com restart automático (68-03), o operador precisa entender por que uma estação caiu. Esta task expõe o histórico de falhas por estação em endpoint de leitura no Studio e um painel em pt-BR, sem emojis e sem dependências.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** um handler de leitura em `internal/studio` e um bloco de UI nos assets embutidos; a regra de negócio (histórico) já vive no pool (`ARQ-10`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Histórico de falhas | `internal/acp` (68-03) | `FailureHistory` por estação | reusar | Única fonte; Studio só lê |
| Endpoints de leitura | `internal/studio/server.go`, `state.go` | Handlers GET e agregador read-only | estender | Novo `GET /api/acp/failures` seguindo o padrão existente |
| Atualização em tempo real | `internal/studio/sse.go` | Canal SSE de estado das estações | reusar | Painel atualiza ao mudar estado (reiniciando/erro) |
| UI do Studio | `internal/studio/assets/index.html`, `studio.js`, css | Painéis e chips | estender | Novo painel; textos pt-BR; sem emojis |
| Testes do Studio | `internal/studio/server_test.go` | Testes de handlers | estender | Casos com pool falso |

## Dependências novas
- Nenhuma. Apenas stdlib Go e assets embutidos (`SEG-07`, `ARQ-03`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-03`, `ARQ-10`, `SEG-05` (sem args/env/segredos no payload), `COD-05` (pt-BR, sem emojis), `TST-02`, `TST-03`, `TST-04`.
- Rules: `sdd-acp.md`, `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.

## Contrato
- `GET /api/acp/failures` (somente leitura; método diferente de GET retorna 405) retorna, por estação: `station`, `adapter`, `currentState`, `healthMode`, e `failures[]` com `timestamp`, `cause`, `exitCode`, `attempt`, `state`.
- Painel: tabela por estação com colunas Horário, Causa, Código de saída, Tentativas, Estado atual; estado vazio "Nenhuma falha registrada".

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race -run "TestStudioFailures" ./internal/studio/... && go test -race ./...
```
- Endpoint devolve histórico do pool falso, na ordem cronológica, sem args nem env.
- POST/PUT/DELETE retornam 405.
- Asset embutido contém o painel e não contém emojis; textos em pt-BR.
- Pool nulo (modo leitura) retorna lista vazia com 200.

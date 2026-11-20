# feat-104-05 — Modelo por papel no forge.json e nos agentes

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** nenhuma

## Problema
Subagentes e papéis usam `model: inherit`; não há como pôr modelo menor nos papéis mecânicos nem no Codex.

## Comportamento
1. `forge.json` ganha `agent_models` por papel (`discovery`, `builder`, `reviewer`, subagentes A/B herdam do papel).
2. `forge update` escreve o `model:` no cabeçalho de `.claude/agents/*.md` e o `-m` no perfil do Codex; valor validado contra lista conhecida (`SEG-01`).
3. Sem a chave, nada muda (`inherit`). Gemini só aceita modelo; Copilot sem flag confirmada.

## Modelo do módulo
**Enxuto**: estende `internal/config` e `internal/scaffold`.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0005, ADR-0014 (feat-90-07).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Modelo/esforço por sessão | `internal/config/forge_profiles.go`, feat-90-07 | `Models`, `Efforts`, `{model}` | **Reusar**; estender ao cabeçalho dos agentes | A 85 foi absorvida pela 90; `feat/85-...` local é antiga e não traz pasta de discovery |
| Cabeçalho dos agentes | `.claude/agents/*.md` | `model: inherit` | **Estender** | Orçamento 420 bytes |
| Perfil Codex | `cli_profiles` | Sem `-m` confirmado | **Estender**, marcado como não verificado | `codex` não instalado aqui |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `SEG-01`, `DUP-06`, `COD-03`, `TST-01`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestAgentModelsEscreveCabecalho|TestAgentModelsRecusaValorInvalido' ./internal/scaffold | grep -F 'TestAgentModelsEscreveCabecalho' >/dev/null
go test ./internal/scaffold -run 'TestAgentModelsEscreveCabecalho|TestAgentModelsRecusaValorInvalido'
# negativo: modelo fora da lista é recusado
go test ./internal/scaffold -run 'TestAgentModelsRecusaValorInvalido'
go vet ./... && go test ./...
```
- Teste negativo: modelo fora da lista é recusado; sem chave o cabeçalho segue `inherit`.
- Definição de agente continua até 420 bytes (P-24).

## A confirmar no Gate 1
Valores aceitos por agente (Codex e Agy a verificar).

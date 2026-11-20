# feat-104-09 — Camada de entrada multiagente: matriz de capacidades e adaptadores

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** draft (aguarda Gate 1)
**Depende de:** feat-104-08, feat-104-04

## Problema
O diagrama tem 4 entradas (CLI, Claude, Codex, Agy) e o dono quer alcançar também o app desktop. Cada agente tem recursos diferentes (hooks, MCP, subagentes, `-m`).

## Comportamento
1. `internal/config` ganha `Capabilities` por agente: `hooks`, `mcp`, `subagentes`, `modelo_flag`, `ferramentas_por_papel`, cada uma `confirmado|suposto|nao`.
2. Um adaptador por agente (Claude Code terminal e desktop, Codex, Agy, e a CLI `forge` como entrada de texto) nivelado pelo menor denominador comum: MCP `forge_doctor` e comandos `forge` são o ferramental comum.
3. `forge doctor --harness` imprime a matriz; capacidade `suposto` aparece como aviso.
4. Preferência de transporte: MCP stdio; `acp` (já em `gemini --acp`, `agy acp`) como alternativa.

## Modelo do módulo
**Enxuto**: estende `internal/config` e `internal/scaffold` (adaptadores); sem pacote novo.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0005 (agente por sessão).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Perfis por agente | `internal/config/forge_profiles.go` (`CLIProfile`) | claude, gemini, copilot, codex; Agy parcial | **Estender** com capacidades | `DUP-01` |
| ACP | `internal/acp` | Servidor JSON-RPC | **Reusar** | Sem dependência |
| Adaptadores | `internal/scaffold/templates` | Por agente | **Reusar** | — |
| Tabela de agentes | `docs/pipeline.md` | Matriz informal | **Atualizar** | Fonte única |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-06`, `COD-01`, `DUP-01`, `TST-01`. Rules: `sdd-acp`, `sdd-arquitetura`.

## Critérios de Aceitação Executáveis
```bash
go test -list 'TestMatrizCapacidades|TestCapacidadeSupostaGeraAviso' ./internal/config | grep -q 'TestMatrizCapacidades'
go test ./internal/config -run 'TestMatrizCapacidades|TestCapacidadeSupostaGeraAviso'
go vet ./... && go test ./...
```
- Teste negativo: agente sem capacidade `subagentes` cai para o modo `sessions` com aviso.

## A confirmar no Gate 1
Quem é a caixa "CLI" do diagrama: o `forge` ou um terminal qualquer? Padrão conservador: o `forge` como entrada de texto.

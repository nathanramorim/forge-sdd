# feat-104-05 — Modelo por papel no forge.json e nos agentes

**Iniciativa:** 104-pipeline-em-subagentes
**Branch:** feat/104-pipeline-em-subagentes
**Status:** implementada em 2026-11-20 (escrita no Claude Code concluída; Codex e Agy pendentes por falta de flag confirmada)
**Depende de:** nenhuma

## Problema
Subagentes e papéis usam `model: inherit`; não há como pôr modelo menor nos papéis mecânicos nem no Codex.

## Comportamento
1. `forge.json` ganha `agent_models` por papel (`discovery`, `builder`, `reviewer`, subagentes A/B herdam do papel).
2. `forge update` (e `forge` na instalação) escreve o `model:` no cabeçalho de `.claude/agents/*.md`; valor validado na borda (`SEG-01`): não vazio, sem caractere de controle, `[A-Za-z0-9._:-]`, até 80, começando por letra, sem terminar em `:` ou `-` e fora dos escalares YAML (`null`, `true`, `false`, `yes`, `no`, `on`, `off`, `y`, `n`, `~`). O `-m` do Codex fica fora (ver Decisões).
3. Sem a chave, nada muda (`inherit`). Gemini só aceita modelo; Copilot sem flag confirmada.

## Modelo do módulo
**Enxuto**: estende `internal/config` e `internal/scaffold`.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo), ADR-0008 (modo `sessions` convive), ADR-0015 (proposta desta iniciativa), ADR-0003 (orçamento), mais: ADR-0005, ADR-0014 (feat-90-07).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Modelo/esforço por sessão | `internal/config/forge_profiles.go`, feat-90-07 | `Models`, `Efforts`, `{model}` | **Reusar**; estender ao cabeçalho dos agentes | A 85 foi absorvida pela 90; `feat/85-...` local é antiga e não traz pasta de discovery |
| Validação de modelo | `config.ValidModel` (forge_profiles.go) | regex SEG-01 | **Reusar** | Mesmo formato de `models` |
| Cabeçalho dos agentes | `.claude/agents/*.md` | `model: inherit` | **Estender** (`scaffold.SyncAgentModels` (também chamado por `scaffold.Run`)) | Orçamento 420 bytes |
| Perfil Codex | `cli_profiles` | Sem `-m` confirmado | **Adiado**: flag não confirmada | `codex` não instalado aqui |
| Teto de 420 bytes | `gate.MaxRoleAdapterBytes` | Constante do gate | **Reusar** (importada em `scaffold`) | DUP-06; sem cópia local |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `SEG-01`, `DUP-06`, `COD-03`, `TST-01`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestAgentModelsEscreveCabecalho|TestAgentModelsRecusaValorInvalido' ./internal/scaffold | grep -F 'TestAgentModelsEscreveCabecalho' >/dev/null
go test ./internal/scaffold -run 'TestAgentModelsEscreveCabecalho|TestAgentModelsRecusaValorInvalido'
go test ./cmd/forge-sdd -run 'TestUpdateAgentModels_AplicaSemAtualizacaoDeVersao|TestUpdateAgentModels_DryRunNaoGrava'
# negativo: modelo inválido é recusado
go test -list 'TestAgentModelsRecusaValorInvalido' ./internal/scaffold | grep -F 'TestAgentModelsRecusaValorInvalido' >/dev/null
go test ./internal/scaffold -run 'TestAgentModelsRecusaValorInvalido'
# negativo: update pelo comando recusa valor inválido sem gravar
go test -list 'TestUpdateAgentModels_ValorInvalidoErraSemGravar' ./cmd/forge-sdd | grep -F 'TestUpdateAgentModels_ValorInvalidoErraSemGravar' >/dev/null
go test ./cmd/forge-sdd -run 'TestUpdateAgentModels_ValorInvalidoErraSemGravar'
# negativo: escalares YAML (null, true, a:, -1) são recusados
go test -list 'TestAgentModelsRecusaEscalarYAML' ./internal/scaffold | grep -F 'TestAgentModelsRecusaEscalarYAML' >/dev/null
go test ./internal/scaffold -run 'TestAgentModelsRecusaEscalarYAML'
go vet ./... && go test ./...
```
- Teste negativo: modelo fora da lista é recusado; sem chave o cabeçalho segue `inherit`.
- Definição de agente continua até 420 bytes (P-24).

## Como usar
Em `.forge/config/forge.json`:
```json
{ "agent_models": { "discovery": "opus", "builder": "sonnet", "reviewer": "haiku" } }
```
Rode `forge update` (`--dry-run` mostra sem gravar). O `model:` vai para o cabeçalho de `.claude/agents/*.md`: discovery para pm-architect e planner, builder para builder, reviewer para reviewer e governance; os demais seguem `inherit`. Diferente de `models`/`efforts`, que escolhem o `--model`/`--effort` da sessão do pipeline.

## Decisões da implementação
- Chave nova `agent_models` (papel -> modelo), separada de `models` (modelo da sessão por `--model`): uma escolhe a CLI da sessão, a outra o cabeçalho do agente.
- Papel -> arquivos: `discovery` = pm-architect e planner; `builder` = builder; `reviewer` = reviewer e governance. Demais agentes (lider, migrator, archivist) seguem `inherit`.
- Valores: o formato de `ValidModel`, não lista fixa (o Claude Code aceita alias e ID completo; lista fixa envelheceria). Recusa com erro `agent_models.<papel>`.
- O `model:` que passaria de 420 bytes (P-24) é recusado sem gravar nada (ex.: governance com nome de 80 caracteres).
- Codex (`-m`) e Agy: não escritos. `codex` não instalado, flag não confirmada, e `modelEffortByCommand` só emite flag confirmada. Fica como follow-up com verificação.
- Gravação em `scaffold.Run` e `RunAgents` fora de `--dry-run`; sem `main.go` tocado. Sem a chave, nada muda.
- `forge update` aplica e valida `agent_models` antes de qualquer outro caminho (`reportAgentModels` em `cmd/forge-sdd/agent_models.go`, uma chamada em `runUpdateFlow`), inclusive sem atualização de versão; `--dry-run` lista sem gravar.
- Codex e Agy: pendentes. Nota para o `doctor` futuro: checar se `agent_models` combina com os cabeçalhos gravados e, quando a flag `-m` do Codex for confirmada, o perfil dele. Não implementado aqui.

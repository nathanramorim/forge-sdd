# feat-58-03 — Fluxo Líder + PM/Arquiteto + Planejador com Reuse Check (`DUP-02`) e Gate de Spec (`FLX-01`)

**Branch:** `feat/58-orquestracao-autonoma-governanca-skills`
**Fase:** 58-03
**Depende de:** `feat-58-01`, `feat-58-02`
**Status:** `done`

## Objetivo

Atualizar os comandos canônicos de concepção e planejamento (`.agents/commands/{discovery,split-features,nova-feature,novo-fix}.md`) para materializar o fluxo **Líder + PM/Arquiteto → Gate Obrigatório de Revisão de Spec (`FLX-01`) → Planejador (`Plan`) → Decisão Humana de Quebra e Enfileiramento**, exigindo nas Specs as seções **"Modelo do módulo" (`ARQ-11`)**, **"O que já existe" (`DUP-02`)** e **"Dependências novas" (`SEG-07`)**.

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `enxuto` (atualização de instruções canônicas em `.agents/commands/` e templates do scaffold).
- **Justificativa:** Padroniza o contrato de entrada/saída dos comandos de concepção e planejamento sem adicionar estado em código Go.

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Comandos canônicos | `.agents/commands/{discovery,split-features,nova-feature,novo-fix}.md` | Comandos unificados na `feat-02-03` com Clarify (`feat-03-04`) e Telemetria (`feat-55`) | `estender` | Incorporar a separação PM/Arquiteto vs Planejador, as 3 seções de `spec-o-que-ja-existe.md` e o Gate de Revisão de Spec antes do enfileiramento |
| Template `spec-o-que-ja-existe.md` | `busize-solutions/planning-brain-governanca` (`templates/spec-o-que-ja-existe.md`) | Seções `Modelo do módulo`, `O que já existe` e `Dependências novas` | `reusar` | Embutir como seções obrigatórias nos artefatos gerados por `/discovery`, `/split-features` e `/nova-feature` |

## Dependências novas (`SEG-07`)
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-fluxo-e-portao.md`
- **IDs:** `DUP-02`, `SEG-07`, `FLX-01`, `ARQ-11`, Regra 14, Regra 15.

## Critérios de Aceitação Executáveis

1. `.agents/commands/discovery.md` (e seu `.md.tmpl`) instrui:
   - **Líder + PM + Arquiteto** (`pm-architect`) a gerarem `discovery-ID-<nome>.md` e `criteria-ID-<nome>.md` contendo obrigatoriamente as seções `Modelo do módulo` (`ARQ-11`), `O que já existe` (`DUP-02`) e `Dependências novas` (`SEG-07`);
   - **Planejador** (`planner`) a consultar `sdd/memory/constitution.md` e `.agents/rules/` para gerar `plan-ID-<nome>.md` mapeando as rules/IDs por task;
   - **Gate 1 (`FLX-01`):** submeter a Spec à revisão humana antes de avançar para `/split-features` ou enfileiramento.
2. `.agents/commands/split-features.md` e `.agents/commands/nova-feature.md` exigem que cada arquivo `feat-XX-*.md` criado contenha as seções `Modelo do módulo`, `O que já existe`, `Dependências novas` e `Rules e IDs da Constituição Aplicáveis`, além de consultar o humano (via Líder) para enfileirar as tasks no Builder (`forge-sdd queue enqueue`).
3. Golden tests em `internal/scaffold/testdata/golden/.agents/commands/` atualizados; `go test ./internal/scaffold/...` e `go vet ./...` passam sem erros.

## Handoff
Depende de `feat-58-01` e `feat-58-02`. Conecta-se diretamente ao motor de fila e portão da `feat-58-04`.

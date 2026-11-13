---
name: sdd-build
description: Estação de Build: o Builder consome a fila e implementa código e testes dentro do escopo da task.
status: approved
origin: forge-sdd-core
scope: core
---

# Skill: sdd-build

Papel: Builder. Implementa só o escopo da task, em contexto isolado.

Comando: `/implementar [feature]` (`references/implementar.md`); `/proxima-feature` é alias. Job do pipeline: `references/worker-job.md`.

## Antes de implementar
1. Leia `sdd/memory/lessons.md`.
2. Pegue a task com `forge-sdd queue next` e leia a spec dela. Rode `forge-sdd context <task>` para obter só os IDs, rules e skills aplicáveis; carregue apenas esses arquivos.
3. Leia o critério de conclusão primeiro; se já estiver atendido, encerre sem alterar código.
4. Biblioteca externa só se a Spec autorizar (`SEG-07`); consulte `context7` na versão exata se estiver ativo em `sdd/memory/mcps.md`.

## Implementação
- Respeite o `O que já existe` (`DUP-02`) e as convenções da stack.
- Testes de caminho feliz, de borda e negativos (`TST-01..03`).
- Sem código morto, comentado ou `TODO` sem ID (`COD-04`, `COD-06`). Não edite fora do escopo.

## Ao finalizar
1. `go vet ./...`, os testes do critério e `forge-sdd gate`, todos com saída 0.
2. Marque `[x]` na spec e rode `forge-sdd queue complete <task-id>`.
3. Commit na branch da iniciativa, nunca em `main` (`FLX-02`).
4. Handoff ao Líder para o Portão de Governança.

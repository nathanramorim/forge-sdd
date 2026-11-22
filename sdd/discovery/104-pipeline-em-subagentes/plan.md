# Plano — Iniciativa 104

Ordem: contratos e verificador, diagnóstico e MCP, modelo por papel, pipeline em subagentes; a avaliação acompanha. Até 7 tasks por onda.

## Onda 1 — Contratos e verificador
- **feat-104-01** Contratos entre etapas. **Depende de:** nenhuma. Rules: ARQ-02, COD-03, DUP-01.
- **feat-104-02** Verificador do verificador. **Depende de:** nenhuma. Rules: TST-01, TST-05.
- **feat-104-11** Conjunto de avaliação (cria a base cedo). **Depende de:** feat-104-02. Rules: TST-01.

## Onda 2 — Diagnóstico e MCP
- **feat-104-03** `forge doctor --harness --json`. **Depende de:** feat-104-01. Rules: ARQ-01, ARQ-05.
- **feat-104-04** Servidor MCP `forge_doctor`, hook `SessionStart` e instrução. **Depende de:** feat-104-03. Rules: ARQ-04, ARQ-06, SEG-04.

## Onda 3 — Modelo por papel
- **feat-104-05** Modelo por papel no `forge.json` e nos agentes. **Depende de:** nenhuma. Rules: SEG-01, DUP-06.

## Onda 4 — Pipeline em subagentes (paralelo entre 06 e 12 após 01)
- **feat-104-06** Laço no Go, lock por iniciativa, retorno estruturado e spike de aninhamento. **Depende de:** feat-104-01. Rules: ARQ-01, FLX-08.
- **feat-104-07** Isolamento: worktree por subagente e junção. **Depende de:** feat-104-06.
- **feat-104-08** Papéis no chat (Discovery, Builder, Reviewer) com permissão pelo harness. **Depende de:** feat-104-07, feat-104-05. Rules: SEG-06.
- **feat-104-09** Camada de entrada multiagente (matriz e adaptadores). **Depende de:** feat-104-08, feat-104-04.
- **feat-104-12** Rastreio por nó. **Depende de:** feat-104-06.
- **feat-104-10** Migração e convivência com o modo de 3 sessões; `docs/pipeline.md`. **Depende de:** feat-104-09, feat-104-12, feat-104-11.

Paralelo: 02 com 01; 05 com tudo da onda 1; 07 e 12 após 06.

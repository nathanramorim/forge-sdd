# Comando: novo-fix

**Uso:** `/novo-fix <descrição do bug>`.

1. **Brief primeiro:** rode `forge-sdd brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de qualquer spec.
2. **Clarify:** `sdd/memory/clarify.md`.
3. **Branch primeiro:** pergunte a branch de partida (padrão `main`); `git branch --list fix/*` para retomar; `git checkout -b fix/<id>-<nome>`.
4. **Origem:** identifique (ou pergunte) a feature ou iniciativa que gerou o bug.
5. **Local:** `sdd/discovery/<id>-<nome>/fixes/fix-<id>-<nn>-<nome>.md`, na pasta da iniciativa de origem.
6. **Conteúdo:** as seis seções obrigatórias da skill `sdd-concepcao` (incluindo `## Decisões do projeto aplicadas`) e a seção `## Feature Correlacionada` (arquivo de origem, causa-raiz e evidência do defeito).
7. **Telemetria:** `forge-sdd session record --feature "<arquivo criado>" --outcome blocked --criterio-atendido=false`.

Handoff: `/implementar`.

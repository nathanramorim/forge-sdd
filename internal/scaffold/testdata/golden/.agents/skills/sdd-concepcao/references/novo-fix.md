# Comando: novo-fix

**Uso:** `/novo-fix <descrição do bug>`.

1. **Clarify:** `sdd/memory/clarify.md`.
2. **Branch primeiro:** pergunte a branch de partida (padrão `main`); `git branch --list fix/*` para retomar; `git checkout -b fix/<id>-<nome>`.
3. **Origem:** identifique (ou pergunte) a feature ou iniciativa que gerou o bug.
4. **Local:** `sdd/discovery/<id>-<nome>/fixes/fix-<id>-<nn>-<nome>.md`, na pasta da iniciativa de origem.
5. **Conteúdo:** as cinco seções obrigatórias da skill `sdd-concepcao` e a seção `## Feature Correlacionada` (arquivo de origem, causa-raiz e evidência do defeito).
6. **Telemetria:** `forge-sdd session record --feature "<arquivo criado>" --outcome blocked --criterio-atendido=false`.

Handoff: `/implementar`.

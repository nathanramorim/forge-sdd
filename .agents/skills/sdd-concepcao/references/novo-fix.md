# Comando: novo-fix

**Uso:** `/novo-fix <descrição do bug>`.

1. **Brief primeiro:** rode `forge brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de qualquer spec.
2. **Clarify:** `sdd/memory/clarify.md`.
3. **Branch primeiro:** pergunte a branch de partida (padrão `main`); `git branch --list fix/*` para retomar; `git checkout -b fix/<id>-<nome>`.
4. **Origem:** identifique (ou pergunte) a feature ou iniciativa que gerou o bug.
5. **Local:** `sdd/discovery/<id>-<nome>/fixes/fix-<id>-<nn>-<nome>.md`, na pasta da iniciativa de origem.
6. **Conteúdo:** as seis seções obrigatórias da skill `sdd-concepcao` (incluindo `## Decisões do projeto aplicadas`) e a seção `## Feature Correlacionada` (arquivo de origem, causa-raiz e evidência do defeito).
7. **Telemetria:** `forge session record --sessao spec --feature "<arquivo criado>" --tokens-input <N> --tokens-output <N> --outcome blocked --criterio-atendido=false --files-touched <lista>`. Nunca passe zeros: se não souber os tokens, omita as flags (o CLI vai estimar).

**Visualizar artefatos (antes do Gate 1)**

Use AskUserQuestion com as opções:
- "Abrir pasta no Finder" → `open sdd/discovery/<id>-<nome>/`
- "Abrir no VS Code" → verifique `command -v code && code sdd/discovery/<id>-<nome>/` (fallback: Finder se VS Code ausente)
- "Ver lista dos arquivos" → liste os arquivos em `sdd/discovery/<id>-<nome>/`
- "Pular"

Execute a ação escolhida. Consulte `sdd/memory/clarify.md` → seção "Visualização pós-spec" para comportamento detalhado (macOS-only, fallbacks).

Handoff: `/implementar`.

# Comando: nova-feature

**Uso:** `/nova-feature <descrição>`; com `fix: <descrição>` vira correção (prefixo `fix/`, fluxo de `novo-fix`).

1. **Brief primeiro:** rode `forge brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de qualquer spec.
2. **Clarify:** `sdd/memory/clarify.md`.
3. **Branch primeiro (`FLX-02`, Regra 15):** pergunte a branch de partida (padrão `main`); rode `git branch --list feat/*` e pergunte se retoma uma existente; só então `git checkout -b feat/<id>-<nome>`. Uma branch cobre toda a pasta da iniciativa.
4. **ID:** `forge next-id`.
5. **Local:** dentro de uma iniciativa existente, `sdd/discovery/<id>-<nome>/features/feat-<id>-<nn>-<nome>.md`. Sem discovery, crie uma iniciativa enxuta `sdd/discovery/<id>-<nome>/` com `status.md` e `features/`.
6. **Conteúdo:** as seis seções obrigatórias da skill `sdd-concepcao` (incluindo `## Decisões do projeto aplicadas`).

**Visualizar artefatos (antes do Gate 1)**

Use AskUserQuestion com as opções:
- "Abrir pasta no Finder" → `open sdd/discovery/<id>-<nome>/`
- "Abrir no VS Code" → verifique `command -v code && code sdd/discovery/<id>-<nome>/` (fallback: Finder se VS Code ausente)
- "Ver lista dos arquivos" → liste os arquivos em `sdd/discovery/<id>-<nome>/`
- "Pular"

Execute a ação escolhida. Consulte `sdd/memory/clarify.md` → seção "Visualização pós-spec" para comportamento detalhado (macOS-only, fallbacks).

7. **Gate e fila:** submeta ao Líder (`FLX-01`); aprovada, enfileire com `forge queue enqueue`.
8. **Telemetria:** `forge session record --feature "<arquivo criado>" --tokens-input <N> --tokens-output <N> --outcome blocked --criterio-atendido=false --files-touched <lista>`. Nunca passe zeros: se não souber os tokens, omita as flags (o CLI vai estimar).

Handoff: `/implementar`.

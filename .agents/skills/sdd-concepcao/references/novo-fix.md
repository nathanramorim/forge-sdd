# Comando: novo-fix

**Uso:** `/novo-fix <descrição do bug>`.

1. **Brief primeiro:** rode `forge brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de qualquer spec.
2. **Clarify:** `sdd/memory/clarify.md`.
3. **Branch da spec primeiro (ADR-0022):** pergunte a branch de partida (padrão `main`); `git branch --list fix/*` para retomar; `git checkout -b fix/<id>-<nome>`, **a partir da principal**. Esta branch carrega **só a spec do fix**; a correção de código vai numa branch por feature do split, aberta pelo Builder.
4. **Origem:** identifique (ou pergunte) a feature ou iniciativa que gerou o bug.
5. **Local:** `sdd/discovery/<id>-<nome>/fixes/fix-<id>-<nn>-<nome>.md`, na pasta da iniciativa de origem.
6. **Conteúdo:** as seis seções obrigatórias da skill `sdd-concepcao` (incluindo `## Decisões do projeto aplicadas`) e a seção `## Feature Correlacionada` (arquivo de origem, causa-raiz e evidência do defeito).
7. **Telemetria:** `forge session record --sessao spec --feature "<arquivo criado>" --tokens-input <N> --tokens-output <N> --outcome blocked --criterio-atendido=false --files-touched <lista>`. Nunca passe zeros: se não souber os tokens, omita as flags (o CLI vai estimar).

**Issue/card (antes do Gate 1):** rode `forge queue card --feature fix-<id>-<nn>-<nome>` na raiz do projeto. Gestor desligado: nada a fazer, em silêncio. GitHub: cria ou reaproveita a issue (`card-evidence.json`). MCP ativo em `sdd/memory/mcps.md`: crie o card e registre com `--provider/--id/--url`. Falha vira aviso e não bloqueia a spec.

**Visualizar artefatos (antes do Gate 1)**

Use AskUserQuestion com as opções:
- "Abrir pasta no Finder" → `open sdd/discovery/<id>-<nome>/`
- "Abrir no VS Code" → verifique `command -v code && code sdd/discovery/<id>-<nome>/` (fallback: Finder se VS Code ausente)
- "Ver lista dos arquivos" → liste os arquivos em `sdd/discovery/<id>-<nome>/`
- "Pular"

Execute a ação escolhida. Consulte `sdd/memory/clarify.md` → seção "Visualização pós-spec" para comportamento detalhado (macOS-only, fallbacks).

## PR da spec do fix em draft (RN-1, RN-7)

```bash
git add sdd/discovery/<id>-<nome>/fixes && git commit -m "docs(<id>): spec do fix <nome>"
git push -u origin fix/<id>-<nome>
gh pr create --draft --base main --head fix/<id>-<nome> --fill
```

Grave o sinal `spec-para-revisao` no `status.md` e deixe a **Conferência 1** conferir o julgamento do fix (causa-raiz e escopo contra as rules). Com o **[Seguir]** do humano, grave `Gate 1: aprovado`, marque o PR pronto e **mescle a spec na principal**. A correção começa depois, a partir da principal, em **uma branch por feature** do split.

Handoff: `/implementar`.

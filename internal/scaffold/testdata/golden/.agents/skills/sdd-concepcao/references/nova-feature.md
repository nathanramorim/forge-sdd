# Comando: nova-feature

**Uso:** `/nova-feature <descrição>`; com `fix: <descrição>` vira correção (prefixo `fix/`, fluxo de `novo-fix`).

1. **Brief primeiro:** rode `forge brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de qualquer spec.
2. **Clarify:** `sdd/memory/clarify.md`.
3. **Branch da spec primeiro (`FLX-02`, Regra 15, ADR-0022):** pergunte a branch de partida (padrão `main`); rode `git branch --list feat/*` e pergunte se retoma uma existente; só então `git checkout -b feat/<id>-<nome>`, **a partir da principal**. Esta branch carrega **só a spec**. A construção terá outra branch, uma por feature do split, aberta pelo Builder.
4. **ID:** `forge next-id`.
5. **Local:** dentro de uma iniciativa existente, `sdd/discovery/<id>-<nome>/features/feat-<id>-<nn>-<nome>.md`. Sem discovery, crie uma iniciativa enxuta `sdd/discovery/<id>-<nome>/` com `status.md` e `features/`.
6. **Conteúdo:** as seis seções obrigatórias da skill `sdd-concepcao` (incluindo `## Decisões do projeto aplicadas`).

**Issue/card (antes do Gate 1):** rode `forge queue card --feature feat-<id>-<nn>-<nome> --dir sdd/discovery/<id>-<nome>`. Gestor desligado: nada a fazer, em silêncio. GitHub: cria ou reaproveita a issue (`card-evidence.json`). MCP ativo em `sdd/memory/mcps.md`: crie o card e registre com `--provider/--id/--url`. Falha vira aviso e não bloqueia a spec.

**Visualizar artefatos (antes do Gate 1)**

Use AskUserQuestion com as opções:
- "Abrir pasta no Finder" → `open sdd/discovery/<id>-<nome>/`
- "Abrir no VS Code" → verifique `command -v code && code sdd/discovery/<id>-<nome>/` (fallback: Finder se VS Code ausente)
- "Ver lista dos arquivos" → liste os arquivos em `sdd/discovery/<id>-<nome>/`
- "Pular"

Execute a ação escolhida. Consulte `sdd/memory/clarify.md` → seção "Visualização pós-spec" para comportamento detalhado (macOS-only, fallbacks).

7. **Gate e fila:** submeta ao Líder (`FLX-01`); aprovada, enfileire com `forge queue enqueue`.
8. **Telemetria:** `forge session record --sessao spec --feature "<arquivo criado>" --tokens-input <N> --tokens-output <N> --outcome blocked --criterio-atendido=false --files-touched <lista>`. Nunca passe zeros: se não souber os tokens, omita as flags (o CLI vai estimar).

## PR da spec em draft (RN-1, RN-7)

A spec viaja em **PR próprio, em draft**, separada do código:

```bash
git add sdd/discovery/<id>-<nome> && git commit -m "docs(<id>): spec de <nome>"
git push -u origin feat/<id>-<nome>
gh pr create --draft --base main --head feat/<id>-<nome> --fill
```

- Grave o sinal `spec-para-revisao` no `status.md`; a **Conferência 1** confere só o julgamento.
- Com o **[Seguir]** do humano, grave `Gate 1: aprovado`, marque o PR pronto e **mescle a spec na principal**.
- Depois da mescla, o código começa a partir da principal, em **uma branch por feature** (`feat/<id>-<nome>-<sid>`) com o seu PR em draft.

**Nunca** reutilize a branch da spec para implementar: ela volta à principal e some.

Handoff: `/implementar`.

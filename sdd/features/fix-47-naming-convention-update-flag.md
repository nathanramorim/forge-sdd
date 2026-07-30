# fix/naming-convention-not-applied-on-update

**Branch:** `fix/naming-convention-not-applied-on-update`
**Fase:** 47
**Depende de:** `fix-5ae2-08-naming-telemetry-tokens`
**Status:** `done`

## Objetivo

Corrigir a impossibilidade de alterar `naming_convention` em um projeto Forge-SDD
já inicializado. A flag `--naming-convention` só existia em `initCmd`, mas todo
projeto existente é redirecionado para `runUpdateFlow`, que ignorava
completamente essa flag — a mudança nunca era persistida em `sdd/.sddrc`, e os
prompts (`/discovery`, `/nova-feature`, `/novo-fix`) continuavam operando com a
convenção antiga.

## Causa raiz

1. `updateCmd` nunca registrava a flag `--naming-convention`.
2. `runUpdateFlow` nunca lia essa flag, mesmo quando presente (via `initCmd`
   redirecionando para update em projeto já existente).
3. Mesmo corrigindo a leitura, dois retornos antecipados na função
   (`"Nenhuma atualização necessária"` e `"Nenhuma modificação realizada"`)
   abortavam a execução antes de persistir a mudança quando nenhuma outra
   flag (`--agent`, `--upgrade`, `--version`) era combinada.

## Critérios de Aceitação Executáveis

1. `updateCmd` registra a flag `--naming-convention` (`sequencial`, `hash` ou `workitem`).
2. `runUpdateFlow` lê a flag (via `cmd.Flags().Lookup`) e, se alterada, sobrescreve
   `rc.NamingConvention` antes de montar a config final.
3. Passar apenas `--yes --naming-convention <valor>` (sem `--agent`/`--upgrade`/`--version`)
   não deve mais cair no early-return "Nenhuma atualização necessária" — deve
   persistir a nova convenção em `sdd/.sddrc` e confirmar no terminal.
4. Testado manualmente: `forge-sdd update --yes --naming-convention workitem`
   em um projeto com `naming_convention: sequencial` resulta em
   `sdd/.sddrc` com `"naming_convention": "workitem"`.

## Handoff

Fix pontual em `cmd/forge-sdd/main.go` (`runUpdateFlow` e registro de flags de
`updateCmd`). Sem impacto em `scaffold`, `config` ou nos prompts dos agentes
— estes já liam corretamente `naming_convention` de `sdd/.sddrc` em tempo de
execução; o problema era exclusivamente a persistência da mudança via CLI.

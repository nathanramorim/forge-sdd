# feat/upgrade-preserves-domain

**Branch:** `feat/upgrade-preserves-domain`
**Fase:** 31
**Depende de:** `feat-30-fix-branch-pr-flow`
**Status:** `todo`

## Objetivo
Implementar proteção no CLI para que, ao inicializar/atualizar o Forge-SDD em um projeto existente (seja rodando `init` novamente ou por meio do `/upgrade-sdd`), os arquivos de domínio do usuário não sejam sobrescritos, enquanto os arquivos de instrução dos agentes e a versão sejam atualizados.

## Critério de conclusão
```
- Arquivos estruturais e de agentes (.gemini/, .claude/, .github/, .vscode/mcp.json, GEMINI.md, CLAUDE.md) são sobrescritos/atualizados.
- O arquivo sdd/.sdd-version e sdd/.sddrc são sempre atualizados com a versão mais nova do CLI.
- Todos os outros arquivos dentro de sdd/ (memory/progress.md, memory/constitution.md, spec/*, features/*, README.md, HOWTO.md, plan.md) não são sobrescritos se já existirem.
- Testes unitários/integração validam a proteção contra sobreposição.
```

## Tarefas
- [x] **31-1** Criar especificação da feature em `sdd/features/feat-31-upgrade-preserves-domain.md`
- [x] **31-2** Modificar `internal/scaffold/scaffold.go` para implementar a lógica de preservação condicional
- [x] **31-3** Escrever testes em `internal/scaffold/scaffold_test.go` para validar que arquivos de domínio não são sobrescritos ao re-executar scaffold
- [x] **31-4** Rodar testes e atualizar golden files
- [x] **31-5** Atualizar logs de progresso e index.md

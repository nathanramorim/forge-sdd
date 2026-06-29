# feat/remove-install-skill

**Branch:** `feat/remove-install-skill`
**Fase:** 35
**Depende de:** `feat-34-update-agent-prompts`
**Status:** `done`

## Objetivo

Remover completamente a funcionalidade de instalação de skill (`/install-skill`) que estava implementada de forma incompleta e não definida. A remoção envolve a deleção de todos os prompts associados e referências na documentação e instruções nos três agentes suportados (Gemini, Claude e Copilot).

## Critério de conclusão

```bash
# 1. Nenhum arquivo deve conter o termo "install-skill.prompt" na árvore de arquivos
! find . -name "*install-skill.prompt*" | grep -q .

# 2. Referências ao comando "/install-skill" ou "instalar skill" devem ser removidas
! grep -r -i -E "/install-skill|instalar skill de" README.md docs/ GEMINI.md CLAUDE.md internal/scaffold/templates/

# 3. Os testes unitários devem passar sem quebras devido à ausência dos templates
go test ./...
go vet ./...
```

## Tarefas

- [x] **35-1** Criar especificação da feature em `sdd/features/feat-35-remove-install-skill.md`
- [x] **35-2** Atualizar `sdd/memory/progress.md` e `sdd/features/index.md`
- [x] **35-3** Deletar arquivos de prompt `install-skill` de todos os diretórios (workspace, templates, golden)
- [x] **35-4** Remover referências a `/install-skill` nas documentações (`README.md`, `docs/comandos.md`, `docs/skills.md`, `npm/README.md`)
- [x] **35-5** Remover referências a `/install-skill` nos arquivos de agente do workspace e templates (GEMINI.md, CLAUDE.md, etc.)
- [x] **35-6** Executar testes unitários e de integração (`go test ./...` e `go vet ./...`)

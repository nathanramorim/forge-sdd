# feat/fix-branch-pr-flow

**Branch:** `feat/fix-branch-pr-flow`
**Fase:** 30
**Depende de:** `feat-29-remove-homebrew`
**Status:** `todo`

## Objetivo
Corrigir falhas no fluxo em que agentes não criam a branch de novas features antes de escrever os specs correspondentes, e automatizar a geração de Pull Requests com `gh` CLI no fechamento de features sem depender de perguntas bloqueantes.

## Critério de conclusão
```
- Templates do Specifier (.gemini, .claude, .github) instruem de forma mandatória a criação da branch como Passo 1.
- Templates do Orquestrador (.gemini, .claude, .github) automatizam a criação de PRs via `gh pr create` (ou git push + log com link caso sem gh) sem parar para perguntar ao usuário.
- Habilidades e prompts locais do projeto atualizados.
- Todos os testes Go passam com sucesso.
```

## Tarefas
- [x] **30-1** Criar especificação em `sdd/features/feat-30-fix-branch-pr-flow.md`
- [x] **30-2** Atualizar templates do Specifier em `internal/scaffold/templates/` (.gemini, .claude, .github) definindo a criação de branch como Passo 1 obrigatório e impeditivo
- [x] **30-3** Atualizar templates do Orquestrador em `internal/scaffold/templates/` (.gemini, .claude, .github) para automatizar `gh pr create` no fechamento de feature
- [x] **30-4** Replicar essas mudanças nas skills locais de `.gemini/skills/` deste projeto
- [x] **30-5** Executar `go test ./internal/scaffold -update` para atualizar os arquivos golden
- [x] **30-6** Atualizar `progress.md` e `index.md`

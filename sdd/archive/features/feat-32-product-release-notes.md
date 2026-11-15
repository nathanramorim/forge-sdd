# feat/product-release-notes

**Branch:** `feat/product-release-notes`
**Fase:** 32
**Depende de:** `feat-23-agent-specific-mcp-config`
**Status:** `todo`

## Objetivo
Permitir a geração automática de release notes concisas, escritas em linguagem de produto (focada em valor para o usuário/negócio), sempre ao finalizar a implementação de uma feature (no fluxo de handoff do Orquestrador).

## Critério de conclusão
```
- O CLI forge-sdd scaffolda a pasta sdd/releases/ e o arquivo sdd/releases/history.md.
- Os templates de Orquestrador (.gemini, .claude, .github/copilot) são atualizados para instruir a geração e o registro das release notes em sdd/releases/history.md ao finalizar a feature.
- As release notes geradas devem ser curtas, sem detalhes técnicos de código, usando tom de produto (ex: "Agora é possível...").
- Os testes unitários e de integração validam a criação da pasta sdd/releases/ e do arquivo history.md.
```

## Tarefas
- [x] **32-1** Criar especificação da feature em `sdd/features/feat-32-product-release-notes.md`
- [x] **32-2** Criar template `internal/scaffold/templates/sdd/releases/history.md.tmpl`
- [x] **32-3** Atualizar templates de Orquestrador em `internal/scaffold/templates/` para incluir o passo de release notes no Handoff
- [x] **32-4** Atualizar `internal/scaffold/scaffold.go` e testes para garantir a distribuição de `sdd/releases/history.md`
- [x] **32-5** Atualizar logs de progresso e index.md

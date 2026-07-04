# feat/structured-documentation

**Branch:** `feat/structured-documentation`
**Fase:** 36
**Depende de:** `feat-35-remove-install-skill`
**Status:** `done`

## Objetivo

Gerar uma documentação de arquitetura técnica e conceitual unificada e estruturada do framework `forge-sdd`, com mapeamento de pastas/arquivos, papéis e responsabilidades, pipelines operacionais com portões de qualidade, e 6 diagramas Mermaid detalhados, no mesmo nível de riqueza do Craft.ai.

## Critério de conclusão

```bash
# 1. O arquivo docs/arquitetura-e-fluxo-forge-sdd.md deve existir e conter os 6 diagramas Mermaid
grep -q "mermaid" docs/arquitetura-e-fluxo-forge-sdd.md
[ $(grep -c "mermaid" docs/arquitetura-e-fluxo-forge-sdd.md) -ge 6 ]

# 2. O README.md deve conter link apontando para a nova documentação
grep -q "docs/arquitetura-e-fluxo-forge-sdd.md" README.md

# 3. Os testes unitários devem continuar passando
go test ./...
go vet ./...
```

## Tarefas

- [x] **36-1** Criar especificação da feature em `sdd/features/feat-36-structured-documentation.md`
- [x] **36-2** Atualizar `sdd/features/index.md` e `sdd/memory/progress.md` com a nova feature
- [x] **36-3** Criar o documento `docs/arquitetura-e-fluxo-forge-sdd.md` com o mapeamento completo e os 6 diagramas Mermaid
- [x] **36-4** Adicionar referências e link à nova documentação no `README.md`
- [x] **36-5** Validar os testes da CLI rodando `go test ./...` e `go vet ./...`

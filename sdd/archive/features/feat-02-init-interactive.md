# feat/init-interactive

**Branch:** `feat/init-interactive`
**Fase:** 2
**Depende de:** `feat/templates-embed` (mergeada)
**Status:** `todo`

## Objetivo
`forge-sdd init` sem flags apresenta formulário huh, coleta dados e cria a árvore Forge-SDD no diretório corrente.

## Critério de conclusão
```bash
mkdir /tmp/test-interactive && cd /tmp/test-interactive
forge-sdd init
# (responder interativamente: nome=demo, stack=go, db=postgres, telemetria=sim, lang=pt-BR)
ls sdd/memory/constitution.md .github/copilot-instructions.md .vscode/mcp.json
# → todos os 3 devem existir, Exit 0
grep "demo" sdd/memory/constitution.md
# → deve conter o nome do projeto
```

## Tarefas
- [ ] **02-1** `go get github.com/charmbracelet/huh@v0.3.0` (consultar context7 antes)
- [ ] **02-2** Implementar `survey.Run() (Config, error)` com formulário de 5 campos (nome, stack, db, telemetria, idioma)
- [ ] **02-3** Implementar `scaffold.Run(cfg Config, targetDir string) ([]string, error)` que renderiza + escreve
- [ ] **02-4** Conectar `survey.Run()` → `scaffold.Run()` em `cmd/forge-sdd/main.go`
- [ ] **02-5** Imprimir próximos passos ao finalizar (`✓ Estrutura criada`, instruções VS Code)

## Arquivos gerados
```
internal/survey/survey.go      (implementado)
internal/scaffold/scaffold.go  (implementado: Run + render)
cmd/forge-sdd/main.go          (atualizado: conecta survey → scaffold)
go.sum                         (atualizado com huh)
```

## Skills relevantes
(consultar `skills/index.md`)

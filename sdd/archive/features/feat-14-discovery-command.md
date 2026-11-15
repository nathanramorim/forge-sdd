# feat/discovery-command

**Branch:** `feat/discovery-command`
**Fase:** 14
**Depende de:** `feat-11-unify-agent-structures`
**Status:** `done`

## Objetivo

Implementar o comando `/discovery` que realiza um processo de "Discovery" para demandas complexas. O comando deve simular as perspectivas de um Analista de Produto Sênior e de um Engenheiro de Software Sênior para gerar dois artefatos: uma especificação de produto e critérios de aceitação técnicos.

## Critério de conclusão

```bash
# 1. Executar o comando discovery via prompt (simulação)
# O agente deve ler uma descrição e gerar:
# sdd/discovery/discovery-XX-<nome>.md
# sdd/discovery/criteria-XX-<nome>.md

# 2. Verificar existência dos templates no scaffold
test -f internal/scaffold/templates/agents/gemini/.gemini/prompts/discovery.prompt.md.tmpl
test -f internal/scaffold/templates/.github/prompts/discovery.prompt.md.tmpl
```

## Tarefas

- [x] **14-1** Criar pasta `sdd/discovery/` no template do scaffold (opcional ou sob demanda)
- [x] **14-2** Criar prompt `/discovery` para Gemini (`.gemini/prompts/discovery.prompt.md.tmpl`)
- [x] **14-3** Criar prompt `/discovery` para Copilot (`.github/prompts/discovery.prompt.md.tmpl`)
- [x] **14-4** Criar comando `/discovery` para Claude (`.claude/commands/discovery.prompt.md.tmpl`)
- [x] **14-5** Adicionar lógica no `Specifier` para assumir as personas de Product Senior e Engineer Senior durante o discovery
- [x] **14-6** Atualizar `GEMINI.md`, `CLAUDE.md` e `README.md` com a documentação do novo comando

## Arquivos gerados/modificados

```
internal/scaffold/templates/agents/gemini/.gemini/prompts/discovery.prompt.md.tmpl
internal/scaffold/templates/agents/claude/.claude/commands/discovery.prompt.md.tmpl
internal/scaffold/templates/.github/prompts/discovery.prompt.md.tmpl
internal/scaffold/templates/agents/gemini/GEMINI.md.tmpl
internal/scaffold/templates/agents/claude/CLAUDE.md.tmpl
README.md
docs/metodologia-sdd.md
```

## Skills relevantes

- `specifier.chatmode.md` (Discovery & Specification)

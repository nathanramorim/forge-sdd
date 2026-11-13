# feat/constitution-command

**Branch:** `feat/constitution-command`
**Fase:** 16
**Depende de:** `feat-11-unify-agent-structures`, `feat-14-discovery-command`
**Status:** `done`

## Objetivo

Implementar o comando `/constitution` que automatiza a definição da arquitetura e das regras do projeto. O comando deve ser capaz de:
1. **Analisar Codebase:** Ler a estrutura de pastas e arquivos existentes para preencher a `constitution.md` e `spec/stack.md` com as tecnologias e padrões reais do projeto.
2. **Elaborar Arquitetura:** Para projetos do zero, sugerir uma estrutura base e regras de boas práticas baseadas na stack escolhida.
3. **Manter Consistência:** Garantir que novas definições não conflitem com as regras imutáveis já estabelecidas.

## Critério de conclusão

```bash
# 1. Executar o comando constitution em um projeto existente
# O agente deve ler o diretório atual e atualizar:
# sdd/memory/constitution.md (Stack, Regras detectadas)
# sdd/spec/stack.md (Dependências, Layout do projeto)

# 2. Verificar templates de prompt criados
test -f internal/scaffold/templates/agents/gemini/.gemini/prompts/constitution.prompt.md.tmpl
test -f internal/scaffold/templates/.github/prompts/constitution.prompt.md.tmpl
```

## Tarefas

- [x] **16-1** Criar prompt `/constitution` para Gemini (`.gemini/prompts/constitution.prompt.md.tmpl`)
- [x] **16-2** Criar prompt `/constitution` para Copilot (`.github/prompts/constitution.prompt.md.tmpl`)
- [x] **16-3** Criar comando `/constitution` para Claude (`.claude/commands/constitution.prompt.md.tmpl`)
- [x] **16-4** Definir a lógica de "Codebase Discovery" no papel de **Specifier** (usar `list_dir` e `read_file` em arquivos de config como `go.mod`, `package.json`, etc.)
- [x] **16-5** Atualizar `GEMINI.md`, `CLAUDE.md` e `README.md` com the novo comando universal
- [x] **16-6** Atualizar `docs/metodologia-sdd.md` para incluir a fase de "Architectural Alignment" via `/constitution`

## Arquivos gerados/modificados

```
internal/scaffold/templates/
docs/metodologia-sdd.md
README.md
.gemini/prompts/
.github/prompts/
```

## Skills relevantes

- `specifier.chatmode.md` (Architectural Discovery)

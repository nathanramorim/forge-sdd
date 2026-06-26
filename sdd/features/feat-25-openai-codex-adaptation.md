# feat/openai-codex-adaptation

**Branch:** `feat/openai-codex-adaptation`
**Fase:** 1
**Depende de:** feat-09-multi-agent
**Status:** `todo`

## Objetivo
Adaptar a estrutura SDD para suporte nativo ao OpenAI Codex/GPT, permitindo que usuários de ferramentas baseadas em OpenAI (ChatGPT, Cursor, Continue, etc.) tenham instruções otimizadas.

## Critério de conclusão
```bash
./forge-sdd init --agent openai --name test-openai
# → deve criar OPENAI.md na raiz
# → deve criar pasta .openai/ com prompts base
# → OPENAI.md deve seguir o padrão de qualidade do GEMINI.md/CLAUDE.md
```

## Tarefas
- [ ] **25-1** Adicionar `AgentOpenAI = "openai"` em `internal/config/config.go` e atualizar `validAgents`.
- [ ] **25-2** Criar diretório `internal/scaffold/templates/agents/openai/`.
- [ ] **25-3** Adicionar estrutura `.openai/` em `internal/scaffold/templates/agents/openai/`:
    - `system_instructions.md.tmpl`
    - `prompts/` (vazio ou com .gitkeep)
    - `skills/` (vazio ou com .gitkeep)
- [ ] **25-4** Criar `OPENAI.md.tmpl` em `internal/scaffold/templates/agents/openai/`.
- [ ] **25-5** Atualizar `agentTemplateRoot` em `internal/scaffold/scaffold.go` para incluir o caso `openai`.
- [ ] **25-6** Atualizar `internal/survey/survey.go` para incluir "OpenAI (GPT-4/Codex)" no MultiSelect de agentes.
- [ ] **25-7** Validar a geração de arquivos com `go run cmd/forge-sdd/main.go init`.

## Arquivos gerados/modificados
```
internal/config/config.go
internal/scaffold/scaffold.go
internal/survey/survey.go
internal/scaffold/templates/agents/openai/.openai/system_instructions.md.tmpl
internal/scaffold/templates/agents/openai/.openai/prompts/.gitkeep
internal/scaffold/templates/agents/openai/.openai/skills/.gitkeep
internal/scaffold/templates/agents/openai/OPENAI.md.tmpl
```

## Skills relevantes
- `Specifier`: para definição da estrutura e prompts.
- `Builder`: para implementação no código Go.

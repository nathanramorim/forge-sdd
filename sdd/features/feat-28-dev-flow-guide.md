# feat/dev-flow-guide

**Branch:** `feat/dev-flow-guide`
**Fase:** 28
**Depende de:** `feat-17-unified-onboarding-docs`
**Status:** `done`

## Objetivo
Documentar e embutir no SDD um **Guia de Fluxo de Desenvolvimento**, ensinando ao desenvolvedor e aos agentes as regras de uso dos comandos e quando acioná-los:
- **`/discovery`** — quando fazer discovery de uma ideia.
- **`/split-features`** — quando uma feature cresceu demais e deve ser decomposta.
- **`/nova-feature`** — quando registrar uma feature nova no backlog.
- **Critérios de quebra incremental** — como garantir que o split siga princípios de desacoplamento e arquitetura evolutiva.
- **Uso do `gh` CLI** — preferir `gh` em vez de `git` para operações de PR, review e issues; cair para git apenas quando gh não estiver disponível.
- **Correção de fluxo `nova-feature`** — criar a branch da feature **antes** de criar o arquivo `feat-XX.md`, invertendo a ordem atual.

## Critério de conclusão
```bash
# O guia de fluxo deve existir e conter as palavras-chave essenciais
grep -q "discovery" sdd/FLOW.md && grep -q "split" sdd/FLOW.md && grep -q "gh pr" sdd/FLOW.md
# O prompt nova-feature deve criar a branch ANTES do arquivo de spec
grep -q "git checkout -b" .gemini/prompts/nova-feature.prompt.md
```

## Tarefas
- [x] **28-1** Criar `sdd/FLOW.md` com o Guia Completo de Fluxo de Desenvolvimento.
- [x] **28-2** Documentar os critérios de quebra de features (decoupling, incrementalidade, arquitetura evolutiva).
- [x] **28-3** Atualizar `.gemini/prompts/nova-feature.prompt.md` para criar a branch ANTES do arquivo `feat-XX.md`.
- [x] **28-4** Atualizar `.gemini/skills/orquestrador.chatmode.md` para preferir `gh` CLI em operações de PR/push.
- [x] **28-5** Atualizar `.gemini/skills/specifier.chatmode.md` para incluir a instrução de criação de branch antecipada.
- [x] **28-6** Propagar as alterações para os templates dos demais agentes (Copilot/Claude) em `internal/scaffold/templates/`.
- [x] **28-7** Atualizar golden tests (`go test ./internal/scaffold -update`).

## Arquivos gerados/modificados
```
sdd/FLOW.md                                        [NEW]
.gemini/prompts/nova-feature.prompt.md
.gemini/skills/orquestrador.chatmode.md
.gemini/skills/specifier.chatmode.md
internal/scaffold/templates/agents/gemini/.gemini/prompts/nova-feature.prompt.md.tmpl
internal/scaffold/templates/agents/claude/.claude/commands/nova-feature.prompt.md.tmpl
internal/scaffold/templates/.github/prompts/nova-feature.prompt.md.tmpl
```

## Skills relevantes
- `Specifier`
- `Builder`
- `Revisor`

## Critérios de quebra de feature (referência)
Uma feature deve ser quebrada quando:
1. Abrange **mais de um bounded context** (ex: UI + API + DB na mesma feature).
2. Tem **mais de 7 tasks** — sinal de escopo excessivo.
3. Quebrar não cria **dependência circular** entre as sub-features.
4. Cada sub-feature entrega valor **independentemente testável** (critério executável próprio).
5. A ordem de implementação das sub-features deve respeitar a **camada arquitetural** (infra → domínio → aplicação → apresentação).

# feat-79-01 — Diretrizes Pipeline-First nos Prompts, Skills e Documentação

**Branch:** `feat/79-clareza-pipeline-e-identificacao-tmux`
**Fase:** 79-01
**Tamanho:** `P`
**Depende de:** —
**Status:** `completed` (FLX-01 aprovado)

## Contexto & Motivação
A regra legada "pergunte se vai para um subagente" nos prompts e skills do agente causava confusão ao usuário, que interpretava a pergunta como despacho para as sessões da esteira (Builder/Reviewer). O comportamento padrão deve ser seguir a esteira: inspecionar `forge-sdd status`, alertar sobre tarefas pendentes na fila (`queued` para Builder, `review_pending` para Reviewer) e sugerir o handoff antes de novas concepções.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Alteração puramente declarativa em templates de prompts e skills markdown, sem lógica em Go (`ARQ-11`).

## Decisões do projeto aplicadas
- `ADR-0004`: Pipeline Multi-Sessão (Discovery, Builder e Reviewer).
- `D8`: Idioma pt-BR padrão em mensagens e documentação.
- `COD-05`: Clareza e precisão na linguagem.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Ciclo de prompt nos agentes | `GEMINI.md`, `CLAUDE.md`, templates `.agents/skills/sdd-concepcao` | "2. PLAN: reporte a intenção, aguarde confirmação e pergunte se vai para um subagente" | refatorar | Substituir por inspeção de fila do pipeline e priorização da esteira. |
| Templates scaffold | `internal/scaffold/templates/agents/` | Templates de prompt distribuídos para novos projetos | atualizar | Manter paridade com as instruções do repositório core. |
| Documentação do líder / discovery | `.agents/skills/sdd-concepcao/references/discovery.md` | Instruções da Discovery | estender | Reforçar orientação de verificar status da esteira e despachar para Builder/Reviewer. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `COD-05`, `FLX-01`, `DUP-02`.
- Rules: `sdd-fluxo-e-portao.md`, `sdd-codigo-e-testes.md`.

## Escopo
1. Atualizar `GEMINI.md` no passo 2 do ciclo: remover menção ambígua a subagente; adicionar leitura de fila e priorização da esteira.
2. Atualizar `.agents/skills/sdd-concepcao/SKILL.md` e `references/discovery.md` com a heurística: se houver tasks em `queued` ou `review_pending`, priorizar handoff para Builder/Reviewer.
3. Atualizar os templates correspondentes em `internal/scaffold/templates/agents/` para que scaffolds novos herdem as diretrizes atualizadas.

## Critérios de Aceitação Executáveis
```bash
grep -rn "pergunte se vai para um subagente" GEMINI.md .agents/ internal/scaffold/templates/
# Deve retornar 0 ocorrências
```
- Testes de scaffold continuam verdes: `go test ./internal/scaffold/...`.

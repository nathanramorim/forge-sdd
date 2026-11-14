# feat-80-01 — Template Canônico de Status nos Prompts de Agentes

**Branch:** `feat/80-status-executivo-triplice-baixo-token`
**Fase:** 80-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `done` (implementado e validado)

## Contexto & Motivação
Padronizar a verificação rápida de status em todas as CLIs (`/status`) via Padrão Tríplice de baixo consumo de tokens (`tmux capture-pane`, `forge-sdd queue panel`, `git log --oneline -n 5`) e template canônico de saída (< 200 tokens).

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica múltiplos adaptadores de IA e templates de scaffold (`ARQ-11`).

## Decisões do projeto aplicadas
- `ADR-0003`: Skills e Contexto (economia de tokens).
- `ADR-0006`: Visibilidade tmux.
- `D8`: Idioma pt-BR padrão.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Prompt status Gemini | `.gemini/prompts/status.prompt.md` | Prompt do comando /status | refatorar | Padronizar com a tríplice de baixo token. |
| Prompt status Claude | `.claude/commands/status.md` | Comando /status | refatorar | Unificar formato canônico com as demais CLIs. |
| Prompt status Copilot | `.github/prompts/status.prompt.md` | Prompt /status | refatorar | Unificar formato canônico com as demais CLIs. |
| Referência Operação | `.agents/skills/sdd-operacao/references/status.md` | Instrução de status | atualizar | Incluir o padrão tríplice e template canônico. |
| Templates scaffold | `internal/scaffold/templates/` | Templates distribuídos | atualizar | Manter paridade com repositório core. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-01`, `DUP-10`: Formato canônico único (< 200 tokens) e limites de orçamento.
- `COD-01`, `COD-05`: Instruções precisas em pt-BR.

## Critérios de Aceitação Executáveis
1. [x] Todos os prompts de agentes instruem a execução dos 3 passos de inspeção (`tmux capture-pane` ou logs, `forge-sdd queue panel` e `git log --oneline -n 5`).
2. [x] O template de resposta gerado pelos agentes é idêntico em Claude, Gemini, Copilot e Codex.
3. [x] Orçamento de resposta estritamente menor que 250 tokens.

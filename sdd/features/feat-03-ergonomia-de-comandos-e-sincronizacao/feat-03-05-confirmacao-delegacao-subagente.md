# feat/03-ergonomia-de-comandos-e-sincronizacao (03-05)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-05
**Depende de:** — (independente das demais features deste pacote)
**Status:** `todo`

## Objetivo

O protocolo de lifecycle (`CLAUDE.md`/`GEMINI.md`/chatmode Copilot ou ponto único equivalente) ganha uma etapa explícita no passo `PLAN`: antes de iniciar a próxima atividade/comando, perguntar objetivamente se deve ser delegada a um subagente — decisão hoje implícita e não auditável pelo usuário. Critério objetivo evita fricção em tarefas triviais.

## Critérios de Aceitação Executáveis

1. Passo `PLAN` do lifecycle (`READ-MIN → PLAN → ACT → WRITE → CLOSE`) passa a incluir, além de "reportar intenção, aguardar confirmação", a pergunta explícita sobre delegação a subagente.
2. Critério documentado de quando a pergunta é relevante (ex: tarefas de varredura/pesquisa extensa tendem a "sim"; edição pontual e pequena tende a "não") — decisão final é sempre do usuário, nunca automática.
3. Em ferramentas/ambientes sem conceito de subagente, a etapa é tratada como não aplicável (omitida), sem quebrar o corpo/protocolo compartilhado entre os três agentes.
4. Validado em uso real (não só documentado) — pelo menos uma sessão de teste confirmando que a pergunta aparece no momento certo, sem virar fricção constante em tarefas triviais.

## Handoff

Independente de 03-01/03-02/03-03/03-04 — pode ser feita em paralelo. Maior superfície de comportamento do pacote (afeta todo comando, não um ponto isolado); validar com mais cuidado antes de fechar `done`.

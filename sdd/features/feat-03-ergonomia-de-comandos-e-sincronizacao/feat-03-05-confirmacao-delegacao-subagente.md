# feat/03-ergonomia-de-comandos-e-sincronizacao (03-05)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-05
**Depende de:** — (independente das demais features deste pacote)
**Status:** `done`

## Objetivo

O protocolo de lifecycle (`CLAUDE.md`/`GEMINI.md`/chatmode Copilot ou ponto único equivalente) ganha uma etapa explícita no passo `PLAN`: antes de iniciar a próxima atividade/comando, perguntar objetivamente se deve ser delegada a um subagente — decisão hoje implícita e não auditável pelo usuário. Critério objetivo evita fricção em tarefas triviais.

## Critérios de Aceitação Executáveis

1. Passo `PLAN` do lifecycle (`READ-MIN → PLAN → ACT → WRITE → CLOSE`) passa a incluir, além de "reportar intenção, aguardar confirmação", a pergunta explícita sobre delegação a subagente.
2. Critério documentado de quando a pergunta é relevante (ex: tarefas de varredura/pesquisa extensa tendem a "sim"; edição pontual e pequena tende a "não") — decisão final é sempre do usuário, nunca automática.
3. Em ferramentas/ambientes sem conceito de subagente, a etapa é tratada como não aplicável (omitida), sem quebrar o corpo/protocolo compartilhado entre os três agentes.
4. Validado em uso real (não só documentado) — pelo menos uma sessão de teste confirmando que a pergunta aparece no momento certo, sem virar fricção constante em tarefas triviais.

## Handoff

Implementado: passo `PLAN` do lifecycle ganhou a pergunta de delegação a subagente em `CLAUDE.md`, `GEMINI.md`, `OPENAI.md` e `copilot-instructions.md` (dogfood + templates-fonte correspondentes), com o critério documentado e a nota de "não aplicável" para ferramentas sem esse conceito. Golden fixture de `copilot-instructions.md` regenerada. `go build/vet/test` passam. Commit `af2e820`.

**Nota sobre o critério 4 (validação em uso real):** esta implementação rodou de forma autônoma (sem usuário interativo na sessão), então não foi possível validar em uma sessão real se a pergunta aparece "no momento certo" sem virar fricção — isso só é observável em uso subsequente por um humano. Registrado como acompanhamento: a primeira sessão real após esta mudança deve avaliar se o critério 4 se sustenta na prática (frequência da pergunta, se atrapalha tarefas triviais) e ajustar o critério documentado se necessário.

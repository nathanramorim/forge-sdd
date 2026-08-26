# feat/03-ergonomia-de-comandos-e-sincronizacao (03-04)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-04
**Depende de:** — (independente das demais features deste pacote)
**Status:** `todo`

## Objetivo

`/nova-feature`, `/novo-fix` e `/discovery` ganham um passo de clarify (esclarecer dúvidas/ambiguidades da descrição recebida antes de agir), hoje ausente nos três — só perguntam sobre branch de partida/retomada, nunca sobre o conteúdo da demanda em si. A regra vive em uma única fonte referenciada pelos três comandos (mesmo padrão de `sdd/memory/naming-convention.md`), condicional — só pergunta quando detecta lacuna relevante, sem virar checklist forçado a cada invocação.

## Critérios de Aceitação Executáveis

1. Nova regra de clarify em arquivo único (ex: `sdd/memory/clarify.md`), com heurística objetiva de "quando perguntar": ausência de critério de aceitação implícito, escopo com mais de uma interpretação plausível, dependência externa não mencionada.
2. `.agents/commands/nova-feature.md`, `.agents/commands/novo-fix.md` e `.agents/commands/discovery.md` referenciam essa regra — nenhum dos três copia a heurística no próprio corpo.
3. Se nenhum sinal de lacuna for detectado na descrição recebida, o comando segue direto (sem pergunta forçada) — comportamento atual preservado para pedidos já claros.
4. Se algum sinal for detectado, o agente faz uma rodada objetiva de perguntas ao usuário antes do PASSO 1 (branch, em `nova-feature`/`novo-fix`) ou antes de produzir os três artefatos (`discovery`).
5. Comportamento validado nos três comandos, nos três agentes (Claude/Gemini/Copilot via corpo canônico compartilhado).
6. Documentação (`sdd/FLOW.md` ou equivalente) menciona o novo passo de clarify na descrição de cada um dos três comandos.

## Handoff

Independente de 03-01/03-02/03-03/03-05 — pode ser feita em paralelo. Compartilha a mesma preocupação de "quando interromper o usuário com uma pergunta" da feature 03-05 (subagente); revisor decide se vale unificar ou manter separadas.

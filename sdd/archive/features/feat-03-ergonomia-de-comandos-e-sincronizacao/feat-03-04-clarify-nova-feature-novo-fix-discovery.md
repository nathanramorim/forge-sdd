# feat/03-ergonomia-de-comandos-e-sincronizacao (03-04)

**Branch:** `feat/03-ergonomia-de-comandos-e-sincronizacao`
**Fase:** 03-04
**Depende de:** — (independente das demais features deste pacote)
**Status:** `done`

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

Implementado: novo `sdd/memory/clarify.md` (fonte única, mesmo padrão de `naming-convention.md`) com a heurística de 3 sinais objetivos; `.agents/commands/nova-feature.md`, `novo-fix.md` e `discovery.md` referenciam o arquivo num passo 0, antes do PASSO 1 (branch) ou da produção dos artefatos; `sdd/FLOW.md` menciona o novo passo nas seções de Discovery e de Registro/Criação de Branch. Novo template `sdd/memory/clarify.md.tmpl` (escaffoldado uma vez, depois preservado como domínio — mesmo contrato de `naming-convention.md`). Golden fixtures regeneradas. `go build/vet/test` passam. Commit `bbc7457`. Não unificada com 03-05 — mantidas como features separadas (superfícies de mudança distintas: comandos de criação vs. protocolo de lifecycle).

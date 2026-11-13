# Discovery 70 — Pipeline Multi-Sessão Orquestrada (v4.1.0)

Status: aguardando Gate 1 (`FLX-01`). Nenhum código alterado.

## Problema
Hoje uma sessão só faz tudo: concebe, implementa e revisa. Isso tem três custos:
1. **Contexto misturado.** O mesmo contexto carrega a conversa de discovery, o código em construção e a revisão, o que encarece tokens e enviesa a revisão (quem escreveu revisa a si mesmo).
2. **Serialização.** O usuário espera o build terminar para voltar a conversar; não há como seguir refinando a próxima feature enquanto a anterior é construída.
3. **Sem portão de rastreabilidade.** Nada impede de construir uma feature que não tem card no gerenciador de projeto (ClickUp, GitHub Projects, Azure DevOps, Jira), nem há ciclo automático de correção quando o PR viola uma ADR ou rule.

A 4.0.0-beta já separou estado por iniciativa e tornou as instruções enxutas. Falta separar **quem faz o quê em sessões diferentes** e **como elas conversam**.

## Para quem
- **Pessoa que conduz o produto (usuário):** quer conversar e decidir numa sessão de Discovery enquanto outras duas sessões constroem e revisam sozinhas, e só voltar para aprovar o PR já revisado.
- **Time:** quer que todo trabalho tenha card, que o PR respeite ADRs e rules e que o ciclo de correção seja automático e limitado.
- **Agentes das estações:** precisam de um contrato de trabalho explícito (job) em vez de depender de uma conversa longa.

## Decisões já tomadas (Clarify, 2026-11-13)
| Tema | Decisão |
|---|---|
| Branch base | `feat/69`, em worktree próprio (`../forge-sdd-70`). A 4.0.0-beta só existe nela. |
| `/implementar` | Por modo: com `.forge/config/forge.json` faz gate e despacho para o Builder; sem ele mantém o comportamento da 4.0.0-beta. `/implement` é alias. Nenhum projeto existente quebra, então a versão é minor (`4.1.0`). |
| Execução | Discovery escrito nesta sessão, sem subagente. |

## Proposta (o "como" macro)
Três sessões de terminal, nomeadas, que trocam **jobs** por uma fila local em arquivos:

- **`forge-discovery`** (interativa): Discovery e Split carregam ADRs e rules desde o início. `/implementar <feature>` roda o **Gate pré-build** (spec e quebra prontas, ADRs e rules mapeadas, card existente no gerenciador via MCP) e só então enfileira para o Builder e move o card.
- **`forge-builder`** (worker): consome a fila, planeja as ondas de subtarefas independentes, executa em paralelo (subagentes e `git worktree`), roda build, lint e testes, abre ou atualiza o PR ligado ao card e enfileira para o Reviewer.
- **`forge-reviewer`** (worker): confere o PR contra spec, ADRs, rules e padrões. Não conforme: devolve ao Builder (`pr_rework`, prioridade alta) até o limite de tentativas, depois escala para a pessoa. Conforme: move o card e **notifica**.
- **Notificação:** o canal padrão grava no inbox da sessão Discovery (com toast do sistema, se ligado) e exibe o link do PR pronto para aprovação; e-mail, webhook e comentário no card são opcionais.
- **Launcher** `forge-up.sh` e `forge-up.ps1` abrem as três sessões no macOS, Linux e Windows, com `claude` ou `agy`.

A divisão de responsabilidade que torna isso testável: **o que é decisão de modelo fica no agente** (planejar, codificar, julgar conformidade, falar com o MCP); **o que é regra de estado fica no Go** (fila, prioridade, tentativas, anti-loop, gate determinístico, notificações). Assim os seis cenários de aceitação rodam em testes sem chamar nenhum modelo.

## Escopo
**Dentro:** `.forge/` (config e runtime), barramento de jobs, subcomandos `dispatch`, `worker`, `bus`, `inbox`, máquina de estados do Reviewer, notificações (inbox, toast, e-mail, webhook), launcher sh e ps1, skills e agentes (`reviewer`, ajustes em `builder` e `concepcao`), hook de inbox, scaffold, ADR 0004 e release notes da 4.1.0.

**Fora:** substituir o fluxo da 4.0.0-beta (continua valendo sem `forge.json`); sessões ACP físicas (feat 68); criar o card no gerenciador (o gate só exige que exista); aprovar ou mesclar o PR pelo agente; qualquer serviço remoto próprio do Forge.

## Jornada alvo
1. `forge-up` abre as três sessões; Builder e Reviewer mostram `Aguardando jobs na fila...`.
2. Na sessão Discovery a pessoa refina e fatia a feature; o agente já conhece as ADRs e rules.
3. `/implementar auth-oauth`: o gate confirma spec, tasks, ADRs e o card `CU-8421`; move o card para `In Progress`; cria o job.
4. O Builder pega o job, paraleliza o que é independente, valida e abre o PR; o Reviewer revisa.
5. Se houver violação, o PR volta com feedback estruturado, até 3 vezes; na terceira falha a pessoa é avisada no inbox.
6. Se conforme, o card vai a `Ready for Review` e a sessão Discovery mostra: "PR revisado, sem inconformidades, pronto para aprovação: <link>".

## Riscos
| Risco | Mitigação |
|---|---|
| Sessão interativa de CLI não consegue "esperar" jobs sozinha | O worker é um processo Go em laço; para cada job ele invoca o agente de forma não interativa (comando por CLI configurável em `forge.json`). Decisão D-1 em `criteria.md`. |
| Flags do `agy` (headless, `--agent`, `-i`) não confirmadas | Argv por CLI em configuração, com `claude` como referência testada; `agy` marcado a confirmar na entrega e coberto por teste de montagem de comando, não de execução. |
| Job preso em `processing` se o worker cair | Lease com prazo e `bus recover` devolve ao `pending`. |
| Loop infinito de rework | `max_rework_attempts` aplicado pelo Go, com escalonamento. |
| Texto do PR ou do feedback tentando manipular o Builder (injeção) | `rework_feedback` e diff são dados não confiáveis (`FLX-08`); o Builder só aplica o que mapeia a regra ou ADR citada. |
| Segredos em e-mail e webhook | URL e credenciais só por variável de ambiente (`SEG-04`); nunca em `forge.json`. |
| Custo de tokens por job headless | Contexto mínimo por `forge-sdd context`; leitura por job limitada à spec, aos IDs citados e às rules aplicáveis. |
| Terminais por SO não testáveis em CI | Testes cobrem a geração dos comandos de abertura por SO; a abertura real é verificada manualmente. |

## Decisões em aberto para o Gate 1
1. **`forge.json` vs `.sddrc`:** a spec nomeia `.forge/config/forge.json`; proposta: manter, com `internal/config` como único dono dos dois arquivos de configuração (`DUP-01`).
2. **Agente `reviewer`:** definição própria, somente leitura sobre o repositório, com `gh` para comentar no PR quando `comment_on_pr` estiver ligado; proposta: manter `governance` para a Camada B das tasks e criar `reviewer` para PRs.
3. **ADRs:** a spec cita `docs/adr/`; neste projeto elas ficam em `sdd/spec/adr/`. Proposta: diretórios de ADR configuráveis, padrão `sdd/spec/adr`.
4. **Versão:** `4.1.0` (minor, sem quebra por causa do modo). Confirmar.
5. **Escopo do PR de 4.1.0:** PR contra `feat/69` até ela ser mesclada.

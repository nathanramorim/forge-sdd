# Discovery 72 — Visibilidade, tmux e Entrega ao Builder (v4.2.0)

Status: `FLX-01` pendente. Nenhum código alterado.

## Problema
O primeiro uso real do pipeline (projeto `barbearia-app`, 2026-11-14) mostrou cinco falhas no dia a dia, todas de experiência e confiabilidade, não de arquitetura:
1. **O worker fica mudo.** `forge-sdd worker` mostra "job iniciado" e nada mais; o log do job tinha 0 bytes enquanto o agente (agy) criava dezenas de arquivos. Quem quer saber o que acontece precisa abrir logs ou transcrições em JSON bruto.
2. **A Discovery acaba executando.** Depois do `/split-features`, nada despacha as tasks ao Builder; o agente da Discovery tentou dar o próximo passo sozinho, editou `forge.json` e gravou um `card-evidence.json` à mão.
3. **`require_existing_card: false` não vale.** O `dispatch` exige o arquivo de evidência mesmo com o card desligado (`loadEvidence` roda sempre), o que empurrou a Discovery a fabricar o arquivo.
4. **Falha sem motivo e sem pausa.** O Claude bateu o limite de uso de 5 horas e os jobs falharam três vezes com "o agente saiu com código 1"; a mensagem real ("You've hit your 5-hour limit · resets 10:40pm") só existia dentro do log.
5. **Três terminais soltos.** Cada sessão em uma janela separada dificulta ver o conjunto; com `tmux` instalado dá para ver as três juntas, nomeadas.

## Para quem
- **Pessoa que conduz o produto:** abre o pipeline e vê, numa tela, o que a Discovery, o Builder e o Reviewer estão fazendo, sem abrir log nem decifrar JSON.
- **Quem trabalha com cota de agente:** quando o limite estoura, o worker pausa e retoma sozinho, e o job não queima tentativa.

## Decisões já tomadas (Clarify, 2026-11-14)
| Tema | Decisão |
|---|---|
| Layout no tmux | **Três painéis na mesma tela:** Discovery grande à esquerda; Builder e Reviewer empilhados à direita, cada painel com o nome da sessão. |
| Limite de uso | **Pausar e retomar:** o job volta à fila sem gastar tentativa, o motivo e o horário aparecem na tela e no inbox, e o worker espera até o horário informado. |
| Release | **Tudo junto na `4.2.0-beta`:** este discovery entra na mesma beta do 71; a tag só sai quando os dois estiverem prontos. Branch própria `feat/72-...`, empilhada na `feat/71` (`FLX-02`), PR contra ela. |
| Origem do ruído `ECONNREFUSED 47600` | É um hook do **Dockia** em `~/.claude/settings.json` (global, não é do Forge). O Forge não edita configuração global (`SEG-02`); o worker só deixa de despejar esse ruído na tela. |

## Proposta (o "como" macro)
**1. A Discovery nunca executa o build (entrega ao Builder).**
- `/split-features`, em modo pipeline, termina em `forge-sdd dispatch <iniciativa>` (gate e job para o Builder) logo depois da confirmação humana que enfileira as tasks, e para: a Discovery acompanha, não implementa.
- Guarda em Go: em modo pipeline, `forge-sdd queue next` e `queue complete` só funcionam sob um worker (`FORGE_WORKER`); fora dele explicam que o Builder constrói e mostram `forge-sdd dispatch`.
- `dispatch` com `require_existing_card: false` **não pede evidência de card** e cria o job sem card.
- As skills dizem, em uma frase, que a Discovery não edita `forge.json`, `card-evidence.json` nem o código da feature.

**2. O worker mostra o que está fazendo, sem sujeira.** A saída completa do agente continua indo para o log; o terminal recebe linhas curtas:
```
▶ job d15b-core-platform · tentativa 1 · agy · permissões project
  ✎ backend/src/domain/unit.entity.ts
  ✎ backend/package.json (+2 arquivos)
  $ npm test            ✓ 42 passaram
  … trabalhando há 2m10s · 14 arquivos alterados
■ concluído em 4m12s · 14 arquivos · PR #12
```
As linhas vêm de eventos estruturados quando a CLI os emite (`stream-json` do Claude e do Gemini, `--json` do Codex) e, para qualquer CLI (inclusive o agy, que não imprime nada), de um acompanhamento do `git status` a cada poucos segundos. Segredos em comandos são mascarados (`SEG-05`) e o ruído conhecido de hooks globais vira uma linha só. `--raw` devolve a saída bruta.

**3. Falha legível e pausa por limite de uso.** O motivo da falha inclui a última linha útil do agente. Mensagem de limite de uso (Claude, Gemini, Copilot, Codex) devolve o job à fila (sem consumir tentativa), pausa o worker até o horário informado (com teto) e grava um aviso no inbox.

**4. Abrir pelo tmux.** Com `tmux` no PATH, `forge-sdd up` cria uma sessão `forge-<projeto>` com os três painéis nomeados e se conecta a ela (ou troca o cliente, se já estiver dentro do tmux); se a sessão existir, só conecta. `--terminal tmux|native|auto` e `terminal` em `forge.json` escolhem; sem tmux vale o terminal do sistema como hoje.

## Escopo
**Dentro:** guarda e despacho no split, `dispatch` sem card, skills e templates do modo pipeline, visão limpa do worker (eventos por CLI, acompanhamento por git, máscara de segredos, filtro de ruído, `--raw`), motivo legível e pausa por limite de uso (`bus.Requeue`), launcher tmux, ADR 0006, release notes.

**Fora:** painel ou interface ao vivo separado (a visão é o próprio terminal de cada sessão); mexer em `~/.claude/settings.json` ou outras configurações globais; trocar de agente sozinho quando o limite estoura; mudar o Reviewer ou o barramento além do `Requeue`.

## Jornada alvo
1. `forge` → Pipeline → Iniciar: se houver tmux, abre uma tela com três painéis (`forge-discovery`, `forge-builder`, `forge-reviewer`).
2. Na Discovery: refina, aprova o Gate 1, `/split-features`; ao confirmar o enfileiramento, o agente roda `dispatch` e avisa "entregue ao Builder"; não implementa.
3. No painel do Builder: `▶ job ...`, linhas de arquivos e comandos, batimento a cada 30 s.
4. Se o Claude estoura o limite: `⏸ limite de uso do claude, retomo às 22:40`, o job volta à fila, o inbox avisa; às 22:40 o worker retoma.
5. Reviewer aprova: `■ concluído`, inbox na Discovery.

## Riscos
| Risco | Mitigação |
|---|---|
| Formato dos eventos `stream-json`/`--json` de cada CLI não confirmado | Parser tolerante: evento desconhecido é ignorado, nunca impresso como JSON cru; o acompanhamento por `git` cobre qualquer CLI; flags marcadas "a confirmar". |
| Segredo aparecer num comando mostrado na tela (`SEG-05`) | Máscara com os padrões de segredo do gate (P-13), linha truncada e nada de variável de ambiente. |
| tmux: argumentos e nome de sessão viram comando de shell | Nome de sessão e painéis de lista fechada e validados; argv explícito para o `tmux`; comando dos painéis montado pelo mesmo `quoteArg` do launcher (`SEG-03`). |
| Falso positivo de "limite de uso" (texto do código ou da conversa) | Só com saída diferente de zero e só nas últimas linhas da saída; padrões por CLI, testados com mensagens reais. |
| Pausa que nunca termina | Teto de 6 horas; horário não entendido vira espera de 30 minutos com nova tentativa; `Ctrl+C` encerra. |
| A guarda no `queue next` atrapalha quem quer construir na própria sessão | Sem `forge.json` nada muda; com ele, a mensagem explica e há `--local` para o uso consciente. |
| Acompanhar `git status` em repositório grande | Intervalo de 5 s, só no diretório do job, com teto de tempo por consulta. |

## Decisões em aberto para o Gate 1
1. **Guarda:** em modo pipeline, `queue next` e `queue complete` exigem worker (`FORGE_WORKER`) ou `--local`. Aceita?
2. **Flags de saída estruturada nos perfis:** `claude --output-format stream-json --verbose`, `gemini -o stream-json`, `codex --json` entram nos perfis padrão (o log bruto guarda os eventos). Aceita?
3. **Teto da pausa:** 6 horas; passou disso o job falha com o motivo. Aceita?
4. **Entrega automática:** a Discovery roda o `dispatch` depois da confirmação do enfileiramento, sem perguntar de novo. Aceita?

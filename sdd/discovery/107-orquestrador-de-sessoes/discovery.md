# Discovery 107 — Orquestrador de sessões: a entrada do chat vira roteamento para o subagente certo

## Pedido do dono (literal, 20/11/2026)
"gostei de trabalhar assim: essa entrada aqui do chat consegue passar para um orquestrador e ele abre a sessão do subagente que precisa para aquele momento; crie um novo discovery para isso se tornar um padrão no forge-sdd, vamos gerando backlog".

## Por que o ID é 107
`forge next-id` (binário da branch, `go run ./cmd/forge-sdd next-id`) devolveu **108** nesta máquina. Os IDs 105 e 106 estão em PRs abertos (#121, #120) e o 107 foi reservado de propósito pelo dono; a pasta é `sdd/discovery/107-orquestrador-de-sessoes/`. O número de ADR segue a mesma regra: sem número até o merge (ver rascunho no fim).

## Clarify (`sdd/memory/clarify.md`)
O pedido traz critério (padrão repetível), escopo (orquestrador + roteamento) e dependências (104, 105, 106). Não perguntei antes de escrever. As lacunas reais viraram `## Decisões para o Gate 1`.

## Base de evidência e limites
- Binário: `go run ./cmd/forge-sdd` da branch (`brief`, `status`, `next-id`, `gate --help`). Nada foi executado com agentes reais; o desenho **não foi testado de ponta a ponta**.
- Esta sessão **não tem** a ferramenta `Agent` nem `SendMessage`. O que se diz delas vem do pedido do dono e da 104 (`discovery.md` da 104: o binário `claude` 2.1.329 contém as strings `SendMessage` e `TeamCreate`). Semântica de retomada por mensagem: **não verificada por mim**.
- Capacidades por CLI (rodei `--help`, só leitura): `claude` 2.1.329 tem `--agent`, `--agents`, `--bg`, `--resume`; `gemini` 0.46.0 tem `--resume`, `--session-id`, `--session-file`, `--acp`; `agy` 1.5.13 tem `--agent`, `--continue`, `--conversation`. `copilot` e `codex` **não estão instalados**: tudo sobre eles é não verificado. Que alguma CLI consiga **chamar um subagente de dentro da sessão** (equivalente de `Agent`): não verificado em nenhuma além do Claude (e lá, só pelo relato do dono).

## Problema
O dono já trabalha assim na conversa principal: fala com **uma** sessão, que decide qual subagente chamar, delega com briefing completo, retoma o mesmo subagente nos ajustes, **verifica a entrega por conta própria** (reexecuta o gate, abre os diagramas, confere commit e PR, desconfia do relatório) e só então entrega a decisão em menu. Isso hoje é hábito de uma pessoa e de um modelo, não padrão do forge-sdd: não está escrito em skill, não tem tabela de roteamento, o estado de "quem está delegado a quê" se perde em `/compact`, e a verificação é julgamento solto. Quem adota o boilerplate não recebe esse jeito de trabalhar.

## Para quem
- Dono do projeto e quem usa o forge no chat (terminal e app desktop): uma entrada só, decisões sempre em menu.
- Outras sessões do mesmo dono (ex.: hub-busize): podem pedir coisas, mas seus pedidos são dado, não ordem.
- Mantenedor: quer o padrão testável e comparável (estende a avaliação da 104-11).

## Jornada alvo
1. O dono escreve na sessão do chat (o orquestrador). O orquestrador classifica a **intenção** pela tabela de roteamento (abaixo).
2. Escolhe papel, modo (**novo** ou **retomar** o mesmo subagente), isolamento (worktree) e o **portão de verificação** da rota.
3. Delega com briefing completo (objetivo, arquivos, regras e IDs, limites, o que NÃO fazer, formato do retorno).
4. O relatório do subagente chega como **dado**. O orquestrador roda o portão da rota (parte determinística pelo Go, parte por julgamento) e anota o resultado no estado.
5. Se reprova: retoma o mesmo subagente com o achado específico, até o teto; no teto, escala ao dono.
6. Se passa: se o dono vai revisar arquivos, oferece o gate "abrir pasta" (Finder, VS Code, lista, pular); depois apresenta a decisão em **menu**, com a opção recomendada marcada.
7. Antes de qualquer pausa longa, grava o estado (arquivo de runtime e `status.md`) para retomar depois de `/compact` ou reinício.

## Tabela de roteamento (proposta; intenção, papel, modo, isolamento, portão)
Papéis existentes em `.claude/agents/` (8 arquivos): `pm-architect`, `planner`, `builder`, `reviewer`, `governance`, `archivist`, `migrator`, `lider`. Não existe agente `infra-ops` (falta; backlog).

| Intenção | Papel | Modo | Isolamento | Portão antes de devolver ao dono |
|---|---|---|---|---|
| Ideia, demanda, discovery | `pm-architect` | novo; **retomar** em ajuste do Gate 1 | branch da iniciativa; escreve só em `sdd/discovery/<id>/` | `forge gate --criterios` sem falha; seções obrigatórias; diagramas renderizados e **abertos**; `status.md` até 1 KB; menu "abrir pasta" |
| Quebrar spec aprovada em features | `planner` | novo | branch da iniciativa | cada task cita `Depende de` e IDs; teto de 7 tasks por onda |
| Implementar feature ou lote | `builder` (um por feature independente) | novo por task; **retomar** em ajuste | **worktree por subagente paralelo** (`PrepararLeque`, `leque.go:91`) | build, vet e testes **reexecutados pelo orquestrador**; `forge gate --mode bloqueio`; commit existe e a branch não é `main`; diff dentro do escopo da task |
| Revisar entrega (Camada B) | `reviewer` ou `governance` (somente leitura) | **sempre novo** (independência) | sem escrita | parecer pelo contrato `harness.Parecer`; veredito aplicado pelo Go (`ApplyVerdict`, `review.go:60`) |
| Corrigir desvio | `builder` via `/novo-fix` | retomar o builder da task original, se vivo; senão novo | worktree próprio | igual a implementar, mais o teste que reproduz o desvio |
| Arquivar, migrar versão | `archivist`, `migrator` | novo | branch da iniciativa | `forge archive --dry-run` antes; idempotência |
| Operação de infraestrutura (deploy, DNS, release) | **inexistente hoje** | n/a | n/a | fora do escopo; backlog 1 |
| Pergunta de status ou conversa | o próprio orquestrador | n/a | n/a | `forge status`; sem delegar |

## Quem é o orquestrador (opções comparadas; a decisão é do dono)
| Opção | O que é | A favor (evidência) | Contra (evidência) |
|---|---|---|---|
| **A. Sessão principal do chat** | O papel Orquestrador vive em skill/instrução; o modelo da sessão decide e delega | É como o dono já trabalha; a 104 aceitou "chat como único coordenador, um nível de subagentes" (Gate 1 da 104, decisão 1) | Contexto do chat cresce; depende de a ferramenta de agente ter equivalente de `Agent`; a verificação fica por julgamento se nada a tornar determinística |
| **B. Comando `forge` (modo chat da 104, `RunChat`)** | O Go conduz o laço; o modelo só devolve conteúdo | Ordem, limites, lock e retorno determinísticos (`laco.go:124`); estado gravado em `.forge/runtime/chat/<id>.json` | `RunChat` só é referenciado em `laco.go` e `laco_test.go`, **sem chamador em `cmd/`**; `pipeline_mode` **não existe** em `internal/config` (grep sem resultado); o Go não abre sessão de modelo, a interface `Agentes` (`laco.go:48`) é implementada por quem chama; cobre só build e review, não discovery nem roteamento por intenção |
| **C. Agente `lider`** | O agente já existente "Líder (Orquestrador)" | Já descrito como orquestrador (`.claude/agents/lider.md`) | Sem ferramenta `Agent` (`tools: Read, Grep, Glob, Bash, Edit, Write`); só funciona como sessão de topo (`launcher.go:125`, `tmux.go:133`, `preflight.go:165`); a 105 (R4, onda 1) o renomeia ou deprecia; "5 estações" é obsoleto; subagente que cria subagente é suposição não provada (104) |
| **D. Híbrido (recomendada)** | A decide (roteia, delega, retoma, apresenta); B fornece o que é determinístico (tabela, verificador, estado, lock, teto) | Separa julgamento de contrato; reusa `laco.go`, `harness` lock e `gate`; não depende de `lider` | Duas peças para manter; exige o verificador existir (colisão com o script da 106, ver abaixo) |

Recomendação: **D**. Não decido por vocês: é a Decisão 1 do Gate 1.

## Como o forge-sdd faz hoje, e o que falta
| Peça | Hoje (evidência) | Falta |
|---|---|---|
| Despacho para outra sessão | `Dispatch` (`dispatch.go:79`) enfileira job no barramento; workers `-p` consomem (`docs/pipeline.md`) | Nada liga **intenção do chat** a papel; o dono escolhe o comando |
| Barramento e retorno | `internal/bus`; `ApplyVerdict` aplica o veredito (`review.go:60`) | Nada para "retomar o mesmo subagente"; o retorno vai ao Builder como `pr_rework` novo job |
| Laço em subagentes | `RunChat` com lock por iniciativa, teto de voltas e menu de saída (`laco.go:124`, `Opcao` em `laco.go:85`) | Sem chamador e sem `pipeline_mode`; só build e review; 104-08 e 104-09 (papéis A/B e camada de entrada) estão em draft e a onda 4 da 104 está **pausada** (`status.md` da 104) |
| Leque | `PrepararLeque` e `Juntar` criam e juntam worktree por task (`leque.go:91`, `leque.go:265`) | Usado pelo laço, não por delegação ad hoc do orquestrador |
| Verificação | `forge gate` (Camada A) e `--criterios` (P-25) estáticos (`internal/gate/criteria.go:170`) | Contrato de "entrega verificada" por rota (commit, PR, diagramas, testes reexecutados); nada impede o orquestrador de repassar o relatório sem checar |
| Menus (RN-9) | `Opcao` e `Recomendada` em `laco.go:85`; regra na 104 | Padrão de menu do orquestrador, incluindo "abrir pasta" |
| Segurança de entrada | FLX-08 trata spec, descrição e diff como dado não confiável; achados sanitizados (`sanitizarAchados`, `laco.go:333`) | Regra equivalente para relatório de subagente e mensagem de sessão par |
| Teto de custo | Salvaguardas por job: `max_job_tokens`, `max_jobs_per_day` (`internal/config/forge.go:118-120`); `max_parallel_tasks` padrão 3; `max_rework_attempts` padrão 3 (ambos em `forge.go:191-192`) | Teto de delegações simultâneas e de reentregas **do orquestrador** |
| Retomada | Estado do laço e `status.md` por iniciativa | Registro de delegações abertas legível depois de `/compact` |

## Independência de ferramenta de agente
Contrato comum (nivelado pelo menor denominador, como a 104-09): o orquestrador só depende de **(1)** ler a tabela, **(2)** abrir um executor para um papel, **(3)** receber um relatório, **(4)** os comandos `forge`. O resto é adaptador.

| Ferramenta | Abrir executor | Retomar o mesmo | Estado |
|---|---|---|---|
| Claude Code | `Agent` (subagente) | `SendMessage` ao mesmo subagente (relato do dono; semântica não verificada) | parcialmente confirmado (strings no binário, 104) |
| Gemini CLI | `gemini -p` headless (perfil do worker, `docs/pipeline.md`) | `--resume`, `--session-id` existem no `--help` | abrir confirmado no `--help`; chamar subagente de dentro da sessão: **não verificado** |
| Agy | `agy -p`; `--agent` existe | `--continue`, `--conversation` existem no `--help` | idem |
| Copilot, Codex | `copilot -p`, `codex exec` só pela documentação do repo | **não verificado** (não instalados) | não verificado |
| Fallback | sessões em tmux (ADR-0008) com o barramento (`Dispatch`) | job novo no barramento; contexto do job anterior não é retomado | verificado na 104/ADR-0004; é o modo `sessions` |

Regra: onde a capacidade for `suposto` ou `nao`, o orquestrador **cai para o fallback com aviso**, como a decisão 5 do Gate 1 da 104. A matriz `Capabilities` pertence à 104-09; a 107 a consome e não cria outra.

## Verificação antes do handoff: contrato versus julgamento
Duas camadas, como a Camada A e B do `FLX-06`:
- **Determinística (script local, sem LLM):** build, vet e testes reexecutados; `forge gate`; commit existe, branch não é `main`; estado do PR lido pelo VCS da Constituição (github: `gh`; azure-devops: `az repos`; nenhum: pulado com a linha "n/a, sem VCS"); diff só toca arquivos do escopo; `status.md` até 1 KB; imagens dos diagramas existem. Sai com código e lista de checagens com evidência. Cada checagem é conferível; nenhuma depende do relatório do subagente.
- **Julgamento do orquestrador (só o que não é checável):** o diagrama diz a coisa certa? O briefing foi cumprido na intenção? Abrir e olhar continua obrigatório (o orquestrador não pode declarar um diagrama visto sem tê-lo aberto).
- O orquestrador **nunca aprova o próprio gate**: Gate 1 e o fecho de PR são do humano.

## Relação com a 105 e com a 106
Li `origin/feat/105-experiencia-cli-e-harness` e `origin/feat/106-pr-draft-revisao-e-ci` por `git show`, sem editar nada.

| Tema | 105 | 106 | Independente ou colide |
|---|---|---|---|
| Menus em todo handoff | RN-7 (= RN-9 da 104) | RN-11 | **Independente**: a 107 reusa a mesma regra |
| Tema e `internal/ui` | paleta, `NO_COLOR` | n/a | **Independente**; o menu do orquestrador no chat não usa `huh`; se o CLI renderizar, usa `internal/ui` quando existir |
| Agente `lider` | R4: renomear no launcher (`launcher.go:125`, `tmux.go:133`, `preflight.go:165`) | n/a | **Colide** com a opção C (a 107 não depende de `lider`) |
| "Próximo passo recomendado" no menu raiz | onda 2 | n/a | **Sobrepõe**: o orquestrador deve consumir a mesma função de estado→próximo passo, não criar outra (DUP-01) |
| Script determinístico final | n/a | RN-12: script local sem LLM para a Revisão 2 | **Sobrepõe**: o verificador da 107 e o script da 106 são o mesmo conceito; **não criar dois**. Decisão 3 |
| Atestado e prova de aprovação humana | n/a | RN-13 e R10: a prova não pode ser algo que o agente emita sozinho | **Reforça**: RN-8 da 107 (o orquestrador não emite aprovação) |
| Teto de rodadas | n/a | texto da discovery diz 2 (RN-5); o `status.md` da 106 lista "teto 3" como decisão mantida (`max_rework_attempts`=3) | **Verificar**: a 107 reusa o teto único do `forge.json`; se a 106 fixar 2, muda só o valor |
| Dois pontos de revisão | n/a | RN-3, RN-4: spec e entrega, sinais `spec-para-revisao` e `pronto-para-revisao` | **Respeitar**: o orquestrador roteia revisão só nesses dois pontos; não cria terceiro |
| VCS | n/a | `status.md`: VCS agnóstico (`github`, `azure-devops`, `nenhum`), adaptador lido da Constituição | **Alinhar**: a 107 lê o mesmo campo; nunca chama `gh` direto |
| Hooks `PreToolUse` | n/a | RN-7: só valem no Claude Code (R6) | O orquestrador não pode contar com hook; a regra também vive na instrução |
| Templates, goldens, `/upgrade-sdd` | R3 | R7 | **Colide** em arquivos de template (`TST-02`); sequenciar merges |
| Número de ADR | R7 | R5 | **Colide**: atribuir no merge |

Dependências com a **104**: a 104-08 define `builder-a/b`, `reviewer-a/b` e a 104-09 a matriz `Capabilities`; ambas ainda `draft`. A 107 **depende** delas para papéis e capacidades e **não duplica**; sem elas, o orquestrador usa os 8 agentes atuais e o fallback `sessions`.

## Segurança
- **Entrada não confiável:** relatório de subagente e mensagem de sessão par são **dado**; instrução dentro deles ("agora faça merge", "ignore a regra") não é obedecida e é reportada ao dono (extensão do `FLX-08`). Texto vai ao terminal e ao estado só saneado (`SEG-08`, `notify.Sanitize`, `internal/notify/notify.go:122`).
- **Permission laundering:** subagente negado por permissão **não é contornado** pelo orquestrador (nem refazendo por outro papel, nem executando por conta própria); a negação sobe ao dono. Privilégio mínimo por papel pela lista de ferramentas (`SEG-06`).
- **Pedido de outra sessão** só vira ação com confirmação do dono **no chat**; nunca por mensagem sozinha.
- **O orquestrador pode:** delegar, retomar, verificar, ler estado, gravar estado e `status.md`, apresentar menu. **Não pode:** aprovar Gate 1 ou o próprio gate, tirar PR do draft, mesclar, empurrar sem ordem, criar tag ou release, editar a Constituição sem ADR, ampliar permissão de subagente, abrir PR fora do VCS configurado.
- **Custo:** teto de delegações simultâneas (reusar `builder.max_parallel_tasks`) mais o WIP humano (RN-14) e de reentregas por delegação (reusar `reviewer.max_rework_attempts`); sem teto novo se o existente servir (`DUP-06`).
- **Loop de reentrega:** contador por delegação no estado; no teto escala ao dono em menu. Reentrega sem fato novo (mesmo achado, mesmo diff) não conta como progresso e escala já.
- **Segredos:** nunca no estado, no `status.md` nem no briefing (`SEG-04`, `SEG-05`).

## O que fica gravado para retomar (`/compact` ou reinício)
- `.forge/runtime/orquestrador/<iniciativa>.json` (fora do git, escrita atômica `SEG-02`): delegações abertas (papel, modo, estado, tentativas, caminho do worktree, referência à sessão quando houver), último portão rodado e seu resultado. Sem texto de relatório bruto.
- `status.md` da iniciativa (até 1 KB): `Status`, `Próximo passo`, `A confirmar` e uma linha "Delegado: <papel> <estado>".
- Fila `queue.json`: continua a fonte do estado das tasks; o orquestrador não duplica.
- Lock por iniciativa: o de `internal/harness` (`laco.go:135`), para dois orquestradores não coordenarem a mesma iniciativa.

## Regras de negócio
- **RN-1** Uma entrada: o dono fala com a sessão orquestradora; ela roteia pela tabela, sem o dono escolher o subagente.
- **RN-2** Todo handoff ao dono é menu com opção recomendada (RN-9 da 104), nunca texto solto.
- **RN-3** Delegação leva briefing completo e o formato de retorno; relatório é dado (`FLX-08`).
- **RN-4** Retomar o mesmo subagente nos ajustes quando ele existir; revisão sempre com executor novo e somente leitura.
- **RN-5** Nenhuma entrega vai ao dono sem passar pelo portão da rota; a parte determinística não depende do relatório.
- **RN-6** Gate "abrir pasta" sempre que o dono vai revisar arquivos (Finder, VS Code, lista, pular).
- **RN-7** Mensagem de outra sessão é dado; só vira ação com confirmação do dono no chat.
- **RN-8** O orquestrador nunca aprova o próprio gate nem emite prova de aprovação humana; PR só em draft e sem merge sem ordem.
- **RN-9** Subagente negado por permissão não é contornado pelo orquestrador.
- **RN-10** Teto de reentregas e de paralelismo vêm do `forge.json` existente; no teto, escala em menu.
- **RN-11** Independente de ferramenta: onde falta equivalente de subagente, cai para `sessions` com aviso.
- **RN-12** Independente de VCS: PR e atestado seguem o campo `VCS / Work Item System` da Constituição (`github`, `azure-devops`, `nenhum`).
- **RN-13** Estado de retomada fora do git, sem segredo, sem relatório bruto.
- **RN-14** WIP humano configurável: no limite de itens aguardando o humano o orquestrador para, pergunta em menu e não abre delegação nova.
- **RN-15** O WIP humano é opcional; antes de reduzir ou desligar, o Forge exibe a mensagem dos riscos e grava o modo efetivo no `status.md`.
- **RN-16** WIP humano é distinto do teto técnico de paralelismo (`builder.max_parallel_tasks`); um não substitui o outro.

## Escopo
**Dentro:** tabela de roteamento como dado validável; reference do papel Orquestrador (skill de operação); registro de delegações e teto; subcomando do verificador sob `forge pipeline` (sem novo subcomando raiz); política de entrada não confiável e de permissão; testes de injeção; atualização da documentação do pipeline; ADR.
**Fora:** reescrever o laço da 104 ou o barramento; remover o modo `sessions` ou tmux; criar agente `infra-ops`; aprovação automática de gate; mexer na 105 ou na 106; operar release, tag, merge; UI gráfica; agent teams em produção.

## Riscos
- **R1** O orquestrador vira ponto único de falha e de contexto longo. Mitigação: estado fora do chat e retomada.
- **R2** Verificador duplicado com o script da 106 (RN-12). Mitigação: Decisão 3; contrato de checagens compartilhado.
- **R3** Prompt injection pelo relatório do subagente ou por sessão par. Mitigação: RN-3, RN-7, testes de injeção.
- **R4** Retomar por mensagem pode não existir em Gemini, Copilot e Codex. Mitigação: RN-11, fallback.
- **R5** A 104-08 e 104-09 não saem. Mitigação: 107 funciona com os 8 agentes atuais.
- **R6** Custo: orquestrador e verificação reexecutada aumentam tokens. Mitigação: portão determinístico barato primeiro; medir pela avaliação.
- **R7** Colisão de templates e goldens com a 105 e a 106 (`TST-02`).
- **R8** Verificação determinística dá falsa segurança se checar pouco. Mitigação: cada checagem tem teste negativo.

## Gate 1: decisão do dono (Nathan, 20/11/2026)
Aprovados: **(1)** orquestrador híbrido D; **(2)** entrada não confiável: sempre dado, ação só com confirmação do dono no chat, lista fechada do que o orquestrador pode e não pode aprovar, permissão negada não é contornada por outro agente; **(3)** um verificador só, compartilhado com o script local da 106, como subcomando de `forge pipeline`; onda 1 já, onda 2 só depois de a 106 fixar o script. Não escolhidos, em `A confirmar`: teto de reentregas 2 ou 3 (reusa `max_rework_attempts`) e depender ou não de 104-08/09 (recomendação: não depender). Pedido novo, em palavras do dono: "acho que deve ter um WIP configurado, mas WIP humano, porque pro trabalho com IA ir tomando segurança ainda tem que ter essa intervenção humana, por isso que na feature anterior nos colocamos opcional e uma mensagem sobre os riscos". Virou RN-14 a RN-16 e a seção abaixo.

## WIP humano configurável (requisito novo)
- **O que é:** o número máximo de itens que o orquestrador mantém **aguardando decisão ou atenção do humano** (entregas verificadas esperando menu, escaladas, gates abertos) antes de **parar de abrir trabalho novo e pedir atenção**. É distinto do teto técnico de paralelismo do Builder (`builder.max_parallel_tasks`, padrão 3, `internal/config/forge.go:191`), que limita execução simultânea e não depende de gente.
- **Opcional e habilitável por configuração**, como o modo sem humano e o modo do atestado da 106 (RN-14, RN-18 a RN-20). Reduzir ou desligar é escolha explícita.
- **Aviso de riscos obrigatório** antes de reduzir ou desligar: sem pessoa no meio, erro de julgamento do orquestrador ou do subagente passa adiante, entregas se acumulam sem leitura e o custo de tokens fica sem teto humano. O aviso vem em menu ([Manter], [Reduzir], [Desligar]) e o modo efetivo é gravado no `status.md` (uma linha) e no estado do orquestrador.
- **No limite:** o orquestrador não segue; mostra o menu com os itens pendentes e a recomendada (resolver o mais antigo) e só abre delegação nova depois que a fila humana baixar.
- **Padrões propostos (sem fixar; Decisão 9):** ligado, limite 3 itens, aviso sempre, modo gravado no `status.md`.
- **Onde mora (alinhado à decisão xv da 106, sem decidir):** a 106 compara `.forge/config/forge.json` (local, não versionado, esquema estrito rejeita campo novo em binário antigo, `forge.go:198-200`, só existe com pipeline) e campo na Constituição em `Ferramentas e Integrações` ao lado de `VCS / Work Item System` (versionado, aparece no diff, custa ADR e goldens). Proposta: **mesmo local e mesmo vocabulário de modos que a 106 escolher** (`desligado`, `local`, `vcs` no atestado; no WIP, `desligado` ou limite numérico), sem duas fontes de verdade (`DUP-06`). Nome provisório: `wip_humano` (a 106 não define nome de WIP). Risco comum (R14 da 106): rebaixar é editar arquivo; mitigação proposta igual: vale o da branch base e o status grava o modo efetivo.

## Decisões para o Gate 1 (recomendações originais; resultado acima)
1. **Quem é o orquestrador.** Recomendo **D (híbrido)**: sessão do chat decide, o Go verifica e guarda estado. Alternativas: A pura, B, C.
2. **Onde o padrão mora.** Recomendo uma **reference nova em `sdd-operacao`** (junto de `lider.md`), em vez de skill nova; evita uma sexta "estação".
3. **Verificador.** Recomendo **um só** contrato de checagens, compartilhado com o script local da 106 (RN-12), exposto como `forge pipeline verificar` (sem subcomando raiz novo, `ARQ-06` inalterado). A 107 só cria as checagens que a 106 não cobre. Exige alinhar com a 106 antes da onda 2.
4. **Relação com 104-08 e 104-09.** Recomendo **depender** delas e consumir `Capabilities`; sem duplicar.
5. **Tetos.** Recomendo reusar `max_rework_attempts` e `max_parallel_tasks`; sem chave nova. Confirmar o teto 2 ou 3 conforme a 106.
6. **Poder de aprovação.** Recomendo a lista "pode" e "não pode" da seção Segurança como regra fechada.
7. **Sessões pares.** Recomendo confirmação humana no chat para toda ação pedida por outra sessão (RN-7), sem canal de aprovação entre sessões.
8. **Ordem entre iniciativas.** Recomendo a onda 1 (tabela, reference, estado) já, e a onda 2 (verificador) só depois de a 106 fixar o script.
9. **WIP humano** (novo, do dono). Recomendo ligado por padrão com limite 3, aviso de riscos antes de reduzir ou desligar e o campo no local que a 106 fixar na decisão xv. Defaults e local seguem em `A confirmar`.

## Rascunho da ADR (não criar o arquivo; número só no merge)
**Título:** Orquestrador de sessões: roteamento por intenção, delegação, verificação e menu.
- **Status:** proposto. **Regras afetadas:** `SEG-06`, `FLX-01`, `FLX-06`, `FLX-08`, `DUP-01`, `GOV-01`. `ARQ-06` não muda se o verificador for subcomando de `forge pipeline`.
- **Relação:** estende ADR-0015 (modo chat) e ADR-0004; convive com ADR-0008; respeita ADR-0013 (Gate 1 humano salvo `advanced`). Não substitui nenhuma.
- **Contexto:** o padrão existe como hábito do dono; falta contrato.
- **Decisão:** RN-1 a RN-13; opção D; verificador compartilhado com a 106; estado fora do git.
- **Consequências:** (+) padrão repetível e testável; (-) mais um papel a manter, custo de verificação, dependência de 104-08/09 e 106.

## Backlog (fora do escopo; candidatas a futuras iniciativas)
1. Agente `infra-ops` (deploy, DNS, release) com privilégio mínimo e confirmação humana por ação.
2. Revisão em leque: `reviewer` e `governance` em paralelo, com síntese única dos achados.
3. Painel do orquestrador em `forge status --live`: delegações abertas, tentativas, custo.
4. Roteamento que aprende com `sdd/memory/lessons.md` (sugere papel e briefing por histórico de falhas).
5. Orçamento de tokens por orquestração, somado por `forge report`, com alerta antes do teto.
6. Canal tipado entre sessões pares (ex.: hub-busize): pedido estruturado, sempre confirmado pelo dono.
7. Suíte de avaliação do orquestrador: injeção em relatório, reentrega em loop, permissão negada (estende `forge eval` da 104-11).
8. Orquestração de várias iniciativas ao mesmo tempo, com fila de prioridade e lock por iniciativa.

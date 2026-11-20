# Discovery — 106 Política de PR em draft, revisão por conferência e CI só ao finalizar

## Pedido do dono
Definir uma política global (ADR no forge-sdd e mudança nos comandos) porque o GitHub bloqueou as Actions da conta por cobrança. **Isto não resolve o bloqueio atual** (é de pagamento); reduz o consumo depois que a cobrança voltar. Ao longo do Gate 1 o escopo cresceu: VCS agnóstico, atestado opcional, split pelo Builder e, por último, **a especificação sempre volta para a branch principal; o Builder abre uma branch e um PR em draft por feature** ("senão fica amarrado o backlog e não tem escala").

## Por que o ID é 106
`forge next-id` devolveu 105, mas a 105 já está reservada pela Discovery da experiência do CLI. A 107 (orquestrador) foi criada depois. Para não disputar pasta nem número, esta iniciativa usa 106 de propósito.

## Insumo de outra sessão (conferido)
`/tmp/wt-c9fb`, branch `feat/c9fb-politica-ci-draft` (somente leitura). Aproveitei o fato técnico do draft e as regras. Divergências após conferir o repositório: a ADR 0013 já existe (`adr-0013-reviewer-revisa-spec-no-modo-avancado.md`); o forge-sdd é dono do fluxo dos comandos; o auto-PR da 100 colide com a ordem nova.

## Clarify (`sdd/memory/clarify.md`)
Sem pergunta antes de escrever: o pedido traz critério, escopo e dependências. As lacunas reais viraram `## Decisões para o Gate 1`.

## Problema
O GitHub bloqueou as Actions por pagamento; os PRs ficam sem check e o fluxo trava. O consumo evitável vem de rodar o CI em trabalho ainda não conferido. Além disso a spec fica pendurada numa branch de iniciativa até a implementação acabar: o backlog fica amarrado a uma branch e não escala (um único tronco de trabalho por iniciativa, estado de fila editado por todos).

## Para quem
Nathan (dono; decide gates e merge) e os agentes que abrem PR nos repositórios dele (hub-busize, hub-enterprise, forge-sdd, gerador-videos e quem adotar o boilerplate via `/upgrade-sdd`).

## Marcação IA, HUMANO e AUTOMÁTICO (vale para o texto e para todos os diagramas)
- **HUMANO:** decisões e aprovações do Nathan.
- **IA:** agentes e LLM: `pm-architect`, `builder`, `reviewer`.
- **AUTOMÁTICO (sem LLM):** script local de validação, hooks, pipeline do VCS, adaptador de VCS.
Três marcas, não duas: o diferencial da política é tirar do LLM tudo que for checável, e só a terceira marca mostra isso. **Recomendo as três.**

## Fato técnico (apurado pelo dono; não reverificado aqui)
- Não existe opção de "não disparar workflow em draft". O padrão é `if: github.event.pull_request.draft == false` nos jobs e `ready_for_review` nos `types` do `pull_request`.
- Job pulado por `if` não usa runner; a execução aparece vazia. Job pulado conta como sucesso em check obrigatório: um draft pode ser mesclado sem CI. Mitigação: hook e veredito com SHA.

## Diagramas (diagram-studio, assinatura Forge SDD, legenda IA, HUMANO e AUTOMÁTICO)
- Fluxo: `diagramas/forge-106-fluxo.png`.
- Atestado e modos: `diagramas/forge-106-atestado.png`.
- BPMN da especificação: `diagramas/forge-106-bpmn-especificacao.png`; BPMN da construção: `diagramas/forge-106-bpmn-construcao.png`. O dono aprovou manter o BPMN simplificado e regerá-lo mais limpo; o desenho único tinha texto pequeno e setas que cruzavam até três raias, então o dividi em dois, com texto maior e no máximo uma seta que atravessa uma raia. Conferi cada PNG em imagem (legíveis sem ampliar).

## Jornada alvo (novo fluxo; marca entre colchetes)
**Especificação**
1. [HUMANO] No início da sessão escolhe o modo: **com humano** (padrão) ou **sem humano** (aviso de riscos, RN-14), e o modo do atestado (RN-18).
2. [IA pm-architect] `/discovery`, `/nova-feature` ou `/novo-fix`: cria a branch `feat/<id>-<nome>` a partir da base (campo de branch base do projeto, `main` por padrão), escreve **só a spec**, commita, empurra e abre o **PR da spec em draft** (adaptador de VCS). Grava `spec-para-revisao`.
3. [AUTOMÁTICO] O script local valida a spec (estrutura, rules por ID, ADR citada, gate, critérios P-25) sem LLM.
4. [IA reviewer] **Conferência 1**: só o que sobrou de julgamento (spec contra ADRs, rules, Constituição e padrão). Não conforme: volta ao `pm-architect` (teto de rodadas, RN-5).
5. [HUMANO] Conforme: pergunta [Seguir] ou [Ajustar] (Gate 1; no modo sem humano, automático e gravado).
6. [AUTOMÁTICO] [Seguir]: o PR da spec sai do draft e a spec é **mesclada na principal** (quem mescla: Decisão xviii). A jornada de especificação termina aqui; a branch da spec some.

**Construção (por feature)**
7. [IA builder] Parte da **principal** (nunca da branch da iniciativa), lê o `plan.md` e faz o split em features (`FLX-01`), sem nova aprovação.
8. [IA builder] Para cada feature, em paralelo quando sem dependência: **uma branch por feature** (`feat/<id>-<nome>-<sid>`, da principal) e **já abre o PR dela em draft**; vai subindo commits até terminar. Feature dependente espera a predecessora ser mesclada (Decisão xx).
9. [AUTOMÁTICO] Feature pronta (checagem local verde): `pronto-para-revisao` da feature. O script local valida código contra spec (gate, rules, ADR, lições, SHA).
10. [IA reviewer] **Conferência 2, por feature**: só o julgamento do que sobrou. Desvio: o reviewer abre um **novo-fix**, mapeia o erro e grava a lição; [IA builder] corrige com prioridade, em paralelo, e devolve; repete até conformidade, com teto por feature.
11. [HUMANO] Conforme: [Continuar] ou [Ajustar] (agrupado por onda; Decisão xix; automático no modo sem humano).
12. [AUTOMÁTICO] [Continuar]: o PR da feature sai do draft (`marcar pronto` do adaptador), o CI roda uma vez e a pipeline confere o **atestado** por PR (RN-13). Sem atestado válido, bloqueia. Com checks verdes, [HUMANO] decide o merge.
13. Sem VCS (`nenhum`): não há PR. A spec e cada feature ficam em branch pronta; o merge é local (`git merge`) feito pelo dono ou, após [Seguir]/[Continuar], pelo Forge (RN-25).

## Regras de negócio
- **RN-0** A política é **independente de VCS**: um adaptador lido da Constituição (`VCS / Work Item System`: `github`, `azure-devops`, `nenhum`; `sdd/memory/constitution.md:28`) escolhe os comandos. Onde o Azure DevOps não foi verificado, a spec diz "a verificar" e propõe a prova.
- **RN-1** (reescrita) Há **dois tipos de PR em draft**: o da spec e o de cada feature. O da spec nasce na concepção e sai do draft depois da Conferência 1 e do [Seguir]; o da feature nasce quando o Builder abre a branch e sai do draft depois da Conferência 2 e do [Continuar]. Nenhum PR vive em draft "até o fim" de toda a iniciativa.
- **RN-2** Em draft o CI não roda (jobs pulados). Roda uma vez no `ready_for_review` e na principal.
- **RN-3** (reescrita) Duas conferências do reviewer: (1) a spec, antes de ela voltar à principal; (2) o código, **por feature**, antes da pipeline. Nunca no meio do desenvolvimento, nem por commit.
- **RN-4** (reescrita) Sinais determinísticos no `status.md` da iniciativa e no estado da feature: `spec-para-revisao` e `pronto-para-revisao` (da feature: fila da feature vazia e checagem local verde). A conferência só se repete quando o sinal do ponto se repetir.
- **RN-5** (reescrita) O reviewer devolve o veredito ao humano e o ajuste volta a quem originou (pm-architect na spec; builder na feature). Teto de rodadas **por ponto e por feature**: o padrão atual é 3 passadas e 2 devoluções (`max_rework_attempts` = 3); estourou, pergunta ao humano.
- **RN-6** (reescrita) A pergunta de saída do draft ([Seguir] da spec, [Continuar] da feature) só aparece com veredito conforme em tudo, válido para o SHA do HEAD **daquela branch**; commit novo invalida.
- **RN-7** (reescrita) Criar PR só em draft; marcar pronto e mesclar só com veredito conforme do SHA atual daquele PR, em qualquer VCS (GitHub: `gh pr create/ready/merge`; Azure DevOps: `az repos pr ...` e o MCP; `nenhum`: nada a bloquear). Garantido pelos passos dos comandos **e** por hooks do Claude Code (`gh`, `az repos pr` e, se alcançar, MCP: a verificar).
- **RN-8** Todo achado do revisor vira uma lição: um arquivo por lição em `sdd/memory/licoes/<id>.md`, com a rule violada e como evitar; quem escreve spec ou código lê só as lições da rule que vai tocar (`licoes.py --rule COD-02`).
- **RN-9** CI: `concurrency` com cancelamento só em PR; matriz de 3 sistemas só em PR pronto e na principal, nunca em draft; alerta de gasto e checagem mensal por script, sem token no git (`SEG-04`).
- **RN-10** Linux, Windows e macOS; scripts Python com `uv` e `pathlib`; cada ferramenta com um prompt "cole isto na sua IA" curto no README.
- **RN-11** Todo handoff ao humano é pergunta com opções (AskUserQuestion no Claude Code, `Select` no CLI): [Seguir]/[Ajustar], [Continuar]/[Ajustar], merge.
- **RN-12** A validação é a mais determinística possível: um script local, sem LLM, confere o que é checável e entrega a lista do que sobrou para julgamento; o LLM só trata essa lista. Vale nas duas conferências (`--ponto spec|entrega`) e **não roda na pipeline**.
- **RN-13** A pipeline só publica. Conforme o modo do atestado, um job leve confere o atestado do **PR da feature**: veredito conforme para o SHA do HEAD mais a aprovação do dono (ou o modo sem humano declarado). Sem atestado válido, bloqueia.
- **RN-14** Modo sem humano: opcional, escolhido no início, padrão com humano; antes de ativar o Forge informa os riscos (nenhuma pessoa vê o código antes de o PR sair do draft; erro do revisor passa; custo sem teto humano); gravado no `status.md` e no atestado.
- **RN-15** (reescrita) As conferências são as únicas; o novo-fix e o retorno ao Builder fazem parte da Conferência 2 da feature e só reabrem a revisão quando `pronto-para-revisao` se repetir.
- **RN-16** Fix de código já desenvolvido tem prioridade sobre feature nova na fila do Builder; o Builder paraleliza o que não tem dependência (features e fixes).
- **RN-17** Caminho de menor custo em tokens: `gh` no GitHub; MCP do Azure DevOps (autenticação interativa) com `az` de alternativa; o custo real é medido na task 0.1 antes de fixar (número ainda não medido).
- **RN-18** Prova externa do atestado **opcional**, por repositório: `desligado`, `local` ou `vcs`; vale também para o GitHub.
- **RN-19** Fallback determinístico: `vcs` (e a alternativa assinada, onde houver) cai para `local` com aviso e pergunta ao humano quando não configurada ou sem resposta; nunca bloqueia para sempre nem aprova em silêncio; no modo sem humano avisa e grava.
- **RN-20** O Forge informa o que se perde em cada modo antes de ativar e a cada queda; o modo efetivo vai ao `status.md` e ao atestado.
- **RN-21** (nova) **A especificação sempre volta para a branch principal** antes de o Builder começar. A spec não fica em branch de iniciativa esperando a implementação.
- **RN-22** (nova) O Builder lê a fila **a partir da principal** (base configurada) e abre **uma branch e um PR em draft por feature**, da principal, subindo commits até finalizar.
- **RN-23** (nova) O estado de execução não pode causar conflito entre features paralelas: nenhuma feature escreve em arquivo compartilhado por todas (Decisão xvi).
- **RN-24** (nova) Todo diagrama identifica IA, HUMANO e AUTOMÁTICO com legenda.
- **RN-25** (nova) Sem VCS (`nenhum`): não há PR; a spec e cada feature ficam em branch pronta e o merge é local; o atestado é só local.

## Adaptador de VCS
O VCS vem da Constituição; o `card` do PM MCP em `internal/config/forge.go:185` é outro conceito (work item). Preferência: GitHub com `gh`; Azure DevOps com o MCP (autenticação interativa; flag exata: **a verificar**) e `az repos` como alternativa. **fato** = conferido no repositório; **a verificar** = não confirmei (prova na coluna final).

| Operação | `github` | `azure-devops` | `nenhum` | Prova do "a verificar" |
|---|---|---|---|---|
| Criar PR em draft | `gh pr create --draft` (fato: hoje sem `--draft`, `internal/pipeline/pr.go:144`) | MCP (suporte a draft: a verificar); `az repos pr create --draft true` (a verificar) | sem PR: branch pronta | PR de teste por VCS (task 3.5); documentação no context7 |
| Procurar PR aberto da branch | `gh pr list --head <branch> --json` (fato: `pr.go:231`) | `az repos pr list --source-branch` ou MCP (a verificar) | n/a | idem |
| Marcar pronto | `gh pr ready` | MCP ou `az repos pr update --id <n> --draft false` (a verificar) | n/a | idem |
| Mesclar | `gh pr merge` | `az repos pr update --status completed` ou MCP (a verificar) | merge local | idem |
| Aprovação (prova) | Environment com revisor (aprovada); fallback assinado | Environment do Azure Pipelines com aprovadores (aprovada); se o autor aprova o próprio PR: a verificar | só local | PR de teste |
| Pular CI em draft | Actions: `if` de draft e `ready_for_review` | build validation da política de branch; se roda em draft: a verificar | sem CI | PR de teste draft |
| Bloquear o agente | hook em `gh pr ...` | hook em `az repos pr ...` e MCP (a verificar) | nada | teste do hook |

### Modos do atestado (RN-18 a RN-20)
| Modo | O que confere | Fallback | O que se perde |
|---|---|---|---|
| `desligado` | nada além dos hooks e passos locais | n/a | ninguém confere veredito nem aprovação; PR pode sair sem conformidade |
| `local` | veredito conforme com `head_sha` mais aprovação no chat | n/a (fim da cadeia) | sem prova externa: agente com acesso ao repositório e às credenciais pode forjar |
| `vcs` | `local` mais a prova do VCS | prova ausente ou sem resposta: cai no `local`, avisa e pergunta (sem humano: avisa e grava) | exige configuração e uma pessoa para aprovar (gargalo possível) |
`nenhum` só tem `desligado` ou `local`. **Onde mora a configuração (Decisão xv):** `forge.json` tem esquema estrito (`internal/config/forge.go:198-200`), só ativa no modo pipeline (`forge.go:123`) e `.forge/` está no `.gitignore` deste repositório (local, fora do diff); a Constituição (`constitution.md:28`) é versionada, visível à pipeline e aparece no diff, ao custo de ADR (`GOV-01`) e goldens. Risco comum: rebaixar o modo é editar arquivo (R14).

**Hook:** `mergeClaudeHook` grava o grupo sem `matcher` (`internal/scaffold/hooks.go:34-111`); se o `PreToolUse` sem matcher cobre as ferramentas de MCP: a verificar. **Pontos do código que assumem `gh`:** `internal/pipeline/pr.go:31,138,144,231,244`; `internal/queue/pr.go:19,41,60,101,108`; `cmd/forge-sdd/review.go:90`; `internal/config/forge.go:181`; `.agents/skills/sdd-operacao/references/inbox.md:9`. Não há abstração de VCS no Go.

## Spec volta à principal: mecanismo (a)
Mecanismo proposto: o **PR da spec** (draft desde a concepção) é mesclado na principal após a Conferência 1 e o [Seguir]. Quem mescla é a Decisão xviii. Hoje o marcador do Gate 1 no modo normal é a existência de `queue.json` (`goalSpecApproved`, `internal/pipeline/goal.go:285`: "criada pelo Gate 1 humano"); no novo modelo o marcador passa a ser o `status.md` da iniciativa na principal (`Gate 1: aprovado`) e `queue.json` deixa de nascer no Gate 1 (Decisão xvi).

**Por comando:** `/discovery`, `/nova-feature` e `/novo-fix` passam a abrir o PR **só da spec** (branch `feat/<id>-<nome>` ou `fix/<id>-<nome>`), a esperar a Conferência 1 e o [Seguir], e a encerrar com a spec mesclada (ou branch pronta no `nenhum`). O Builder (`/implementar`, esteira) abre um PR por feature. `/split-features` deixa de ser passo humano e vira passo do Builder (FLX-01 aprovado).

## O estado da fila na principal (b)
**Evidência.**
- A fila é por iniciativa e versionada: `sdd/discovery/<id>/queue.json` (`internal/layout/layout.go:293`; ADR-0002 opção 3). O estado de execução (atividade, filas por estação) já vai para `sdd/.runtime/`, fora do git (ADR-0002, `layout.go:215`).
- Quem escreve o status das tasks no `queue.json`: o consumo da fila que incrementa `Attempts` (`internal/queue/queue.go:441`), `queue.Review` (`:488`), `SetTaskStatuses` e `BlockTasks` (`internal/queue/taskstatus.go:40,60`), todos na árvore de trabalho onde rodam.
- Todo escritor regrava o campo de topo `updated_at` (`taskstatus.go:102`, `queue.go:103`). **Reproduzi o conflito** numa cópia de `sdd/discovery/101-tema-missoes-v5/queue.json` (8 itens): duas branches mudando itens diferentes (e `updated_at`) deram `Automatic merge failed` no segundo merge. Não testei a variante sem `updated_at`; itens distantes provavelmente mesclam, mas o campo de topo garante conflito. Com um PR por feature, todo par de PRs paralelos conflitaria.
- A iniciativa ativa vem da **branch atual** (`ActiveInitiative`, `layout.go:222-236`; `FromBranch`, `:304`): numa branch `feat/<id>-<nome>-<sid>` o slug `<id>-<nome>-<sid>` não existe em `sdd/discovery/`, então `QueuePath` cai na fila legada `sdd/.queue/queue.json` (`layout.go:239-245`).
- `archive` exige estar na branch da iniciativa (`branchGuard`, `internal/pipeline/archive.go:187-197`) e lê o `queue.json` (`archive.go:78,241`); com a spec mesclada e a branch removida, isso quebra.

**Opções para onde mora o estado de execução** (Decisão xvi):
- **A. Fora do git** (`sdd/.runtime/`, `.forge/runtime`): sem conflito; mas é por máquina, invisível a outro clone e à pipeline, e perde histórico.
- **B. Um arquivo de estado por feature, na branch da feature** (ex.: `sdd/discovery/<id>/state/<sid>.json`): cada PR toca só o seu arquivo (sem conflito); a fila é **calculada na leitura** (como a visão geral na ADR-0002); o "claim" é a existência da branch (um segundo `git push` do mesmo nome falha); o estado da feature em andamento é visível por `forge status --all-branches`.
- **C. Derivado do VCS**: PR aberto/draft/mesclado define o estado; sem arquivo. Funciona no `github` e `azure-devops`; no `nenhum` cai nas branches locais; depende de rede.
- **Aprovada: B**, com A só para o transitório (jobs, leases, que já estão fora do git) e `queue.json` deixando de receber status por feature (vira plano, escrito uma vez). Custo: mexe em `Claim`, `Review`, `SetTaskStatuses`, `BlockTasks`, `goal.go:285`, `archive.go`, `forge status` e `layout.Current`.

## Branch e PR em draft por feature (c) e o leque
**Evidência.**
- Já existe branch por task no fluxo `goal`: `job.PR.Branch = layout.InitiativeBranch(initiative) + "-" + s.ID` (`internal/pipeline/goal.go:375`), com PR contra a branch da iniciativa (`prBase` → `queue.ResolveBase`, `internal/pipeline/pr.go:167`, `internal/queue/base.go:108`). Hoje a base vem da flag, depois do `queue.json` (`BaseBranch`), depois do reflog da branch da iniciativa, depois do `base_branch` do `forge.json` (padrão `main`, `forge.go:195`). No novo modelo a base é sempre a principal (campo `base_branch`).
- `pipeline.OpenPR` já trata o PR pré-existente (`findOpenPR`, `pr.go:131,230`) e corrige a base; só falta criar em draft e via adaptador.
- O leque (`internal/pipeline/leque.go`) cria um worktree e uma branch local `forge/<iniciativa>-<task>` por task (`leque.go:182-190`) e **junta tudo na branch da iniciativa** com `git merge --no-ff` (`Juntar`, `leque.go:262-300`); `Base` precisa ser a branch em checkout no worktree principal (`leque.go:36`).
- **Conflito:** o leque entrega por junção local numa branch de iniciativa; o novo modelo entrega por PR por feature contra a principal. O que muda: o worktree por task continua útil como isolamento do paralelismo, mas a branch da task passa a ser a **branch da feature publicável** (`feat/<id>-<nome>-<sid>`), `Base` passa a ser a principal e `Juntar` deixa de ser a entrega (a junção passa a ser o merge de cada PR; o conflito entre features vira conflito de PR, resolvido na ordem de merge). O bloqueio com menu (`resolver`, `pular`, `abortar`) migra para o conflito de PR. Decisão xxii.
- **Nome da feature:** `feat/<id>-<nome>-<sid>` já é o nome usado em `goal.go:375`; `FromBranch` o aceita (prefixo `feat`), mas `ActiveInitiative` não acha a pasta (evidência acima). Proposta: `layout.Current` resolver a iniciativa pelo **maior prefixo que existe em `sdd/discovery/`**; `next-id` (`idOfRef`) e o gate P-01 (só bloqueia `main`/`master`) continuam servindo.

## Conferências, teto e atestado por feature (d)
- O teto vale **por feature**: o job é por task (`goalJob(... []string{s.ID} ...)`, `goal.go:~372`) e tem `Attempt` e `MaxReworkAttempts` próprios (`internal/bus/job.go:116-117`); a devolução só ocorre se `Attempt < Max` (`internal/pipeline/review.go:132`). Para a spec o contador é por iniciativa (`nextSpecAttempt`, `internal/pipeline/spec.go:31`).
- O veredito já é por job (`verdicts/<job>.json`, `review.go:298-318`), logo por feature; o `head_sha` é o do HEAD da **branch da feature** (task 1.1).
- O atestado é **por PR de feature**. O PR da spec é só documentação: basta o modo `local` (veredito da Conferência 1 mais o [Seguir]); a prova do VCS é opcional e não vale o gargalo (Decisão xxi).

## O que muda na ordem (conflito com a 100)
Hoje: veredito conforme, depois PR criado sem draft: `queue.OpenPR` (`cmd/forge-sdd/queue.go:195`; `internal/queue/pr.go:61,105`) e `pipeline.OpenPR` (`internal/pipeline/pr.go:71`). Agora: o PR da spec nasce na concepção; o PR de cada feature nasce quando o Builder abre a branch; o veredito conforme só habilita a pergunta de saída do draft.

## Escopo
**Dentro:** ADR (rascunho abaixo); adaptador de VCS (`internal/vcs`); passos dos comandos e skills (spec só; Builder por feature); veredito com `head_sha`, `ponto`, `modo`; sinais; hooks; `licoes.py`; `validar_entrega.py` (`--ponto spec|entrega`); atestado por modo; modelo de workflow; `ci-politica.py`; READMEs; templates, goldens e `/upgrade-sdd`; estado por feature (Decisão xvi); `layout.Current` por prefixo; adaptação do leque; arquivamento fora da branch da iniciativa; primeira leva de repositórios.
**Fora:** pagar ou alterar billing; migrar de CI; repositório de cliente sem aprovação; reescrever o laço da 104 ou o CLI da 105; editar a 104, 105 ou 107; criar o arquivo da ADR agora; alterar a Constituição agora.

## Riscos
- **R1** Draft mesclado sem CI. Mitigação: RN-7; job "pronto?" (Decisão iii, aprovada).
- **R2** Loop de revisão. Mitigação: RN-3, RN-4 e teto por feature.
- **R3** Aprovação velha: o veredito de hoje não tem SHA (`internal/harness/parecer.go:16-21`). Task 1.1.
- **R4** Repositório pessoal privado sem draft: rótulo `wip` como alternativa.
- **R5** Colisão de ADR: número "próxima livre" só no merge.
- **R6** Hooks só valem no Claude Code; Gemini e Copilot dependem dos passos.
- **R7** Goldens e `/upgrade-sdd` (`TST-02`).
- **R8** `uv` nos repositórios-alvo e no Windows: não verificado.
- **R10** Atestado forjável: se o agente roda com a credencial do dono, tudo que a API do VCS emite com ela é emitível por ele; só resiste o que exige algo fora do alcance dele.
- **R11** Split sem aprovação humana: o Builder pode fatiar fora do escopo; o script confere que cada feature cita o `plan.md`.
- **R12** Modo sem humano: ninguém vê o código antes de o PR sair do draft.
- **R13** Loop fix-revisão: teto 3 (2 devoluções) por feature (`review.go:132`); o novo-fix reusa o mesmo `Attempt`.
- **R14** Rebaixar o modo do atestado editando a configuração: ler o modo da base e gravar o efetivo.
- **R15** Gargalo da prova do VCS: opcional, com tempo limite e queda para `local`.
- **R16 (novo, o maior)** Conflito de estado entre features paralelas: `queue.json` com `updated_at` conflita (reproduzido); `ActiveInitiative` não acha a pasta numa branch de feature; `goalSpecApproved` e `archive` dependem de `queue.json` e da branch da iniciativa. Mitigação: Decisão xvi e task 0.7.
- **R17 (novo)** Feature dependente de outra não pode partir da principal até a predecessora ser mesclada; serializa a onda. Mitigação: `PlanWaves` já ordena; Decisão xx.
- **R18 (novo)** Constituição: Regra 15 diz "nunca por task" e `FLX-02` fala em "branch única por iniciativa" (`constitution.md:49,99`); a ADR-0002 diz que a branch `<prefixo>/<id>-<nome>` aponta para a pasta. Branch por feature contradiz; exige ADR e `GOV-01` (Decisão xvii).
- **R19 (novo)** Merge da spec na principal sem pessoa (modo sem humano) ou sem VCS: ver Decisão xviii.

## Rascunho da ADR (não criar o arquivo agora)
**Número:** próxima livre depois das ADRs da 104 (0016 a 0020), da 105 e da 107; atribuir no merge. **Título:** Política de PR em draft, spec que volta à principal, branch por feature e CI só ao finalizar.
- **Contexto:** Actions bloqueadas por cobrança; consumo evitável; spec amarrada a branch de iniciativa; auto-PR da 100 sem draft; `gh` fixo em dois lugares, mas a Constituição já prevê `azure-devops` e `nenhum`.
- **Decisão:** RN-0 a RN-25. Adaptador de VCS (`internal/vcs`); duas conferências do reviewer (spec antes de voltar à principal; código por feature), com validação local determinística sem LLM; Gate 1 por pergunta e split pelo Builder sem nova aprovação; spec mesclada na principal; uma branch e um PR em draft por feature; novo-fix aberto pelo reviewer com prioridade; atestado opcional por modo com fallback; estado de execução sem conflito (Decisão xvi); CI com `if` de draft, `concurrency` e matriz só em PR pronto e principal.
- **Impacto na Constituição** (nada editado agora; entra na task 2.7 via `GOV-01`, aprovado pelo dono): (1) `FLX-01`, redação aprovada: "Spec antes do código: nada é implementado nem enfileirado sem Spec revisada e aprovada pelo humano no Gate 1 ([Seguir]/[Ajustar]); com o Gate 1 aprovado o Builder faz o split e enfileira sem nova aprovação, só dentro do escopo do `plan.md` aprovado; mudança de escopo revalida a Spec." (2) **Regra 15** (`sdd/memory/constitution.md:46`). Hoje: "Uma branch por iniciativa (`<prefixo>/<id>-<nome>`), nunca por task: pergunte o ponto de partida (padrão `main`) e verifique se há branch a retomar. (`FLX-02`)". **Proposta exata:** "15. A spec vive numa branch por iniciativa (`<prefixo>/<id>-<nome>`), com PR próprio, e volta para a principal antes da construção; a construção usa uma branch por feature do split (`<prefixo>/<id>-<nome>-<sid>`), aberta pelo Builder a partir da principal, cada uma com o seu PR em draft: pergunte o ponto de partida (padrão a base configurada, `main`) e verifique se há branch a retomar. (`FLX-02`)" (3) **`FLX-02`** (`constitution.md:97`). Hoje: "Isolamento de branch e iniciativa: sem commit direto em `main` (Regra 1); branch por feature (`feat/*`, `fix/*`, Regra 2); tasks agrupadas em `sdd/discovery/<id>-<nome>/` (Regra 13) e uma branch única por iniciativa (Regra 15)." **Proposta exata:** "Isolamento de branch e iniciativa: sem commit direto em `main` (Regra 1); branch por feature (`feat/*`, `fix/*`, Regra 2); tasks agrupadas em `sdd/discovery/<id>-<nome>/` (Regra 13); a spec em branch e PR próprios que voltam para a principal, e a construção em uma branch por feature a partir da principal (Regra 15)." A Regra 2 ("Branch por feature") já existe e não muda. (4) `GOV-03`: troca `lessons.md` por `sdd/memory/licoes/`. (5) a nota da linha 30 de `VCS / Work Item System` remete ao adaptador. (6) possível campo do modo de atestado (Decisão xv). A ADR-0002 é estendida (estado por feature), não substituída.
- **Consequências:** menos minutos e menos repetição de erro; backlog não amarrado a branch; custo em hooks, goldens, `layout`, fila e leque. Estende ADR-0002, ADR-0004, ADR-0012 e ADR-0013.
- **Alternativas descartadas:** `paths-ignore`; `workflow_dispatch` manual; revisão por task; branch de iniciativa com junção local (contradiz o pedido).

## Decisões para o Gate 1
**Aprovadas e mantidas** (dono, 2026-11-20): (i) `feat/`; (ii) PR de teste de draft em repositório pessoal; (iii) job "pronto?" só na principal; (iv) primeira leva hub-busize, hub-enterprise, forge-sdd, gerador-videos; (v) migrar o `lessons.md`; (vi) atestado GitHub: Environment com revisor, fallback assinado; (vii) teto 3 (2 devoluções); (viii) assinatura Forge SDD; (ix) BPMN simplificado (aprovado manter, regerar mais limpo); `FLX-01` com split pelo Builder; (xi) atestado Azure: Environment com aprovadores, opcional e com fallback; (xii) `nenhum` só local; (xiii) MCP versus CLI a medir; adaptador de VCS e as ondas de VCS agnóstico.

**Aprovadas na reabertura "spec volta à principal"** (dono, 2026-11-20, parcial):
- **(xvi) Estado de execução: opção B.** Um arquivo de estado por feature, na branch da feature (`sdd/discovery/<id>/state/<sid>.json`); a fila é **calculada na leitura** a partir da principal (plano em `plan.md` e `features/`) mais os estados por feature; o **claim é a existência da branch** (o segundo `git push` do mesmo nome falha). Vira a task 0.7, que cobre: `layout.Current` pelo maior prefixo existente (ele alimenta `ActiveInitiative`, `QueuePath` e `StatusPath`); `goalSpecApproved` (o marcador do Gate 1 passa a ser `Gate 1: aprovado` no `status.md` da principal, não `queue.json`); `archive` (`branchGuard` e leitura de `queue.json` passam a rodar a partir da principal e a ler os estados por feature); `queue.json` deixa de receber status por feature (vira plano, escrito uma vez); leitura da fila a partir da principal (`forge status`, `Load`); `Claim`, `Review`, `SetTaskStatuses` e `BlockTasks` passam a gravar o arquivo da feature.
- **(xvii) Constituição por ADR e `GOV-01`**, na task 2.7 (o plano numera a ADR e a Constituição na 2.7; a 2.6 é a dos passos dos comandos). Redação exata da Regra 15 e do `FLX-02` no rascunho da ADR acima.
- **(xviii) Quem mescla a spec:** o [Seguir] do Nathan autoriza e o adaptador de VCS executa o merge; automático no modo sem humano; no VCS `nenhum` o merge é local.
- **(xix) [Continuar] ou [Ajustar] agrupado por onda.**
- **(xx) Feature dependente espera a predecessora ser mesclada** (ondas do `PlanWaves`).
A pergunta "PRs de task do leque em draft?" está respondida: cada feature abre PR em draft.

**Abertas** (com recomendação; o dono ainda não respondeu):
- **(xiv) Padrão do modo de atestado.** Recomendo `local` (sem configuração, sem gargalo, com aviso); `vcs` por adesão depois da prova da task 3.5.
- **(xv) Onde mora a configuração do modo.** Recomendo o campo na Constituição (versionado, visível à pipeline, aparece no diff), ao custo de ADR e goldens.
- **(xxi) Atestado do PR da spec.** Recomendo modo `local` (é documentação; o veredito da Conferência 1 mais o [Seguir]); atestado por PR só nas features.
- **(xxii) Leque: conflito de PR no lugar do merge local.** **Evidência:** `PrepararLeque` e `Juntar` não têm chamador fora de `leque.go` e dos testes (`git grep`), então a mudança é de baixo raio. **Recomendação concreta:** manter `PrepararLeque` (worktree por task, lock do worktree principal) mas (a) `Base` passa a ser a principal (`base_branch`) e deixa de exigir que ela esteja em checkout no worktree principal (`leque.go:36,87-91`, "na branch errada"); (b) a branch da task passa de `forge/<iniciativa>-<task>` para a branch publicável `feat/<id>-<nome>-<sid>` (`leque.go:182-190`; mesmo nome de `goal.go:375`); (c) `Juntar` (`leque.go:262-300`, `merge --no-ff` na branch da iniciativa) **deixa de ser a entrega**: vira a operação "mesclar" do adaptador `nenhum` (merge local) e, nos VCS `github` e `azure-devops`, é substituída por `Publicar` (push e PR em draft pelo adaptador); (d) o menu de conflito (`resolver`, `pular`, `abortar`) é mantido e passa a valer para o conflito de PR e para o merge local do `nenhum`; (e) testes de `leque_test.go` migram com o código. Alternativa: manter `Juntar` também nos três VCS (junção local antes de publicar), que contradiz o pedido de a spec e a feature irem para a principal por PR.
- **Hooks só no Claude Code** (Gemini e Copilot dependem dos passos dos comandos): segue em aberto.

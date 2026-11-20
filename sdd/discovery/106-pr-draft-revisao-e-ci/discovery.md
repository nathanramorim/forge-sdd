# Discovery — 106 Política de PR em draft, revisão em dois pontos e CI só ao finalizar

## Pedido do dono (confirmado por ele no chat desta sessão)
Definir uma política global (ADR no forge-sdd e mudança nos comandos) porque o GitHub bloqueou as Actions da conta por cobrança. **Isto não resolve o bloqueio atual** (é de pagamento); reduz o consumo depois que a cobrança voltar.

## Por que o ID é 106
`forge next-id` devolveu 105, mas a 105 já está reservada pela Discovery da experiência do CLI (branch local `feat/105-experiencia-cli-e-harness`, ainda não mesclada; a pasta `sdd/discovery/105-experiencia-cli-e-harness/` existe só nela). Para não disputar pasta nem número, esta iniciativa usa 106 de propósito.

## Insumo de outra sessão (conferido)
`/tmp/wt-c9fb`, branch `feat/c9fb-politica-ci-draft` (somente leitura). Aproveitei o fato técnico do draft, as regras e o plano em 2 ondas. Divergências após conferir o repositório: (1) o insumo propõe "ADR 0013", mas `sdd/spec/adr/adr-0013-reviewer-revisa-spec-no-modo-avancado.md` já existe (a `main` vai até a 0015; a discovery da 105 diz que a 104 reserva 0016 a 0020 em `bbd26d6`); (2) ele trata o forge-sdd como "upstream"; aqui o forge-sdd é dono do fluxo dos comandos, então a ordem dos comandos e os goldens entram nesta iniciativa; (3) ele não vê o auto-PR da 100, que colide com a ordem nova.

## Clarify (`sdd/memory/clarify.md`)
Sem pergunta antes de escrever: o pedido traz critério, escopo e dependências. As cinco lacunas reais viraram `## Decisões para o Gate 1`.

## Problema
O GitHub bloqueou as Actions da conta por pagamento ou limite de gasto; os PRs ficam sem check e o fluxo trava. Parte do consumo é evitável: cada push em PR roda o CI inteiro, inclusive em trabalho ainda não revisado, e a revisão pode se repetir. O fluxo não distingue PR rascunho de PR final.

## Para quem
Nathan (dono; decide gates e merge) e os agentes que abrem PR nos repositórios dele (hub-busize, hub-enterprise, forge-sdd, gerador-videos e quem adotar o boilerplate via `/upgrade-sdd`).

## Fato técnico (apurado pelo dono; não reverificado aqui)
- Não existe opção de "não disparar workflow em draft". O padrão é `if: github.event.pull_request.draft == false` nos jobs e `ready_for_review` nos `types` do gatilho `pull_request`.
- Job pulado por `if` não usa runner (não gasta minuto); a execução aparece vazia.
- Job pulado conta como sucesso em check obrigatório: um draft pode ser mesclado sem CI. Mitigação: hook e veredito com SHA (RN-6, RN-7).

## Adaptador de VCS (reaberto pelo dono: "a ferramenta de VCS é definida pela Constituição")
O dono pediu que a política não sirva só a quem usa GitHub. O VCS vem da Constituição (`VCS / Work Item System`: `github`, `azure-devops`, `nenhum`; o `card` do PM MCP em `internal/config/forge.go:185` é outro conceito, o work item). Preferência por economia de tokens: GitHub com `gh`; Azure DevOps com o MCP do Azure DevOps, autenticação interativa (flag exata: **a verificar**), e `az repos` como alternativa.

Legenda: **fato** = conferido no repositório; **a verificar** = não confirmei; a prova proposta está na coluna final.

| Operação | `github` | `azure-devops` | `nenhum` | Prova do "a verificar" |
|---|---|---|---|---|
| Criar PR em draft | `gh pr create --draft` (fato: o código hoje usa `gh pr create` sem `--draft`, `internal/pipeline/pr.go:144`) | MCP do Azure DevOps (criar PR; suporte a draft: a verificar); alternativa `az repos pr create --draft true` (nome e valor da flag: a verificar) | não há PR: branch pronta e aviso (fato: nota da Constituição, linha 30) | PR de teste em cada VCS (task 3.5) e consulta à documentação no context7 |
| Procurar PR aberto da branch | `gh pr list --head <branch> --json ...` (fato: `pr.go:231`) | `az repos pr list --source-branch` ou MCP (a verificar) | n/a | idem |
| Marcar pronto | `gh pr ready` | MCP (atualizar PR) ou `az repos pr update --id <n> --draft false` (a verificar) | n/a: o [Continuar] só libera a branch | idem |
| Mesclar | `gh pr merge` (decisão do dono) | `az repos pr update --status completed` ou MCP (a verificar) | merge feito pelo dono fora do Forge | idem |
| Veredito e SHA | arquivo `verdicts/*.json` com `head_sha` (igual em todos) | igual | igual | testes Go |
| Aprovação do dono (prova) | Environment com revisor obrigatório (**aprovada**, Decisão vi); fallback commit/tag assinado | candidatos: política de branch com revisores obrigatórios; aprovação em Environment do Azure Pipelines; se o autor pode aprovar o próprio PR: a verificar | só o veredito local com SHA mais a aprovação no chat | PR de teste e documentação; escolha em Decisão xi |
| Pular CI em draft | Actions: `if: github.event.pull_request.draft == false` e `ready_for_review` (dado pelo dono) | build validation da política de branch; se roda em PR draft: a verificar; variável de draft no pipeline YAML: a verificar | sem CI do VCS | PR de teste draft no Azure |
| Proteger contra o agente | hook `PreToolUse` em `gh pr ...` | hook em `az repos pr ...` e nas ferramentas MCP (se o matcher alcançar: a verificar) | nada a bloquear | teste do hook (task 1.5) |
| O que se perde | nada | depende do que o Azure suportar | prova externa: nada impede o agente de forjar veredito e aprovação; CI do VCS inexistente; risco igual ao do modo sem humano | Decisão xii |

Fatos sobre o hook no repositório: `mergeClaudeHook` grava o grupo `{"hooks":[{"type":"command","command":...}]}` **sem a chave `matcher`** (`internal/scaffold/hooks.go:34-111`); portanto hoje não filtra por ferramenta. Se o `PreToolUse` sem matcher cobre todas as ferramentas, incluindo as de MCP, e como o matcher nomeia ferramentas de MCP: **a verificar** (documentação do Claude Code). Plano: o `pr_guard.py` lê o `tool_name` e o comando do JSON do hook e decide por VCS.

Pontos do código que assumem `gh` hoje: `internal/pipeline/pr.go:31` (lista de comandos permitidos só `git` e `gh`), `:138`, `:144`, `:231`, `:244`; `internal/queue/pr.go:19`, `:41`, `:60`, `:101`, `:108`; `cmd/forge-sdd/review.go:90` (texto "publique com gh pr comment"); `internal/config/forge.go:181` (`AllowCommands: git, gh, forge-sdd`); skills e templates `.agents/skills/sdd-operacao/references/inbox.md:9` (`gh pr merge`) e `sdd-governanca/references/pr-review.md`. Já tratam os três valores só texto: `sdd-build/references/implementar.md:26` e `sdd-concepcao/references/constitution.md` passo 3. Não há abstração de VCS no Go (nenhuma ocorrência de "VCS" em `*.go`) e nenhum código lê o campo da Constituição.

## Diagramas (feitos pela outra sessão com o diagram-studio; conferidos contra esta spec)
- Fluxo: `diagramas/forge-106-fluxo.png` (`.json`, `.svg`).
- Atestado: `diagramas/forge-106-atestado.png` (`.json`, `.svg`).
- BPMN, quem faz o quê: `diagramas/forge-106-bpmn.png` (`.json`, `.svg`).
Conferência em imagem: os três batem com a jornada abaixo (dois pontos de revisão, [Seguir]/[Ajustar], split e paralelo pelo Builder, novo-fix com prioridade, [Continuar]/[Ajustar], job que só confere o atestado, script local fora da Action, modo sem humano). Ressalvas: (a) os PNGs foram regerados com a assinatura do projeto Forge SDD (Decisão viii, resolvida); (b) o fluxo mostra o novo-fix só como retorno ao Builder, sem o teto de rodadas (está na legenda); (c) o diagrama de atestado não nomeia a prova; a aprovada é o Environment com revisor obrigatório (Decisão vi).

## Jornada alvo
1. No início da sessão o Forge pergunta o modo: **com humano** (padrão) ou **sem humano** (opcional, com aviso de riscos, RN-14). O modo fica no `status.md` e no atestado.
2. `/discovery`, `/nova-feature` ou `/novo-fix`: cria a branch, commita a spec, empurra e abre o PR em draft, sem pedir. Grava `spec-para-revisao` no `status.md`.
3. **Revisão 1 (spec):** o revisor valida a estrutura da spec contra ADRs, Constituição, rules e padrão do projeto.
4. **Gate humano (Gate 1):** conforme, o Forge pergunta "posso seguir ou há ajuste?" com [Seguir] e [Ajustar] (ajustar = a pessoa digita). Não conforme: volta a quem originou.
5. [Seguir]: o **próprio Builder** faz o split em features e começa a implementar, **sem nova aprovação humana do split** (muda o `FLX-01`, ver ADR). Tasks sem dependência rodam em paralelo.
6. Commits locais, sem revisor no meio. Fila de todas as features vazia e checagem local verde: `pronto-para-revisao`.
7. **Revisão 2 (entrega):** primeiro o script local determinístico (RN-12); depois o revisor (LLM) só no que exige julgamento, conferindo o código contra a spec e, principalmente, ADRs, rules, Constituição e pré-definições.
8. Desvio: o revisor abre um **novo-fix** para cada desvio (mapeia o erro, alimenta as lições) e devolve ao Builder. Fix de algo já desenvolvido tem prioridade sobre feature nova; o Builder paraleliza os fixes e devolve ao revisor. Repete até conformidade, com teto (RN-5).
9. Conforme: o Forge pergunta [Continuar] ou [Ajustar]. Só [Continuar] tira o PR do draft (operação "marcar pronto" do adaptador de VCS; no GitHub, `gh pr ready`; o CI roda uma vez).
10. Antes de publicar, a pipeline só confere o **atestado** (RN-13). Sem atestado válido, bloqueia. Com checks verdes, o menu do merge (decisão do humano).

## Regras de negócio
- **RN-0** A política é **independente de VCS**: um adaptador lido da Constituição (`VCS / Work Item System`: `github`, `azure-devops`, `nenhum`; `sdd/memory/constitution.md:28`) escolhe os comandos. Não há requisito que valha só para GitHub. Onde o Azure DevOps não foi verificado, a spec diz "a verificar" e propõe a prova (tabela abaixo).
- **RN-1** O PR nasce em draft (operação "criar PR em draft" do adaptador; no GitHub, `gh pr create --draft`), sem perguntar, nos três comandos de concepção. Com `nenhum` não há PR: o Forge deixa a branch pronta e informa.
- **RN-2** Em draft o CI não roda (jobs pulados). Roda uma vez no `ready_for_review` e na `main`.
- **RN-3** Revisão em exatamente dois pontos: antes do Builder (spec) e depois (entrega). Nunca no meio, nem por task entregue.
- **RN-4** Cada ponto tem um sinal determinístico no `status.md`: `spec-para-revisao` e `pronto-para-revisao` (este exige fila da onda vazia e checagem local verde). A revisão só se repete quando o sinal do ponto se repetir.
- **RN-5** O revisor devolve o veredito ao humano. Ajuste volta a quem originou (discovery, nova-feature, novo-fix ou Builder). Teto de 2 rodadas automáticas por ponto; depois pergunta ao humano.
- **RN-6** "Fechar o PR?" só aparece com veredito conforme em tudo. O veredito guarda o SHA do HEAD; commit novo o invalida.
- **RN-7** Criar PR só em draft; marcar pronto e mesclar só com veredito conforme do SHA atual, em qualquer VCS (GitHub: `gh pr create/ready/merge`; Azure DevOps: `az repos pr ...` e as chamadas do MCP do Azure DevOps; `nenhum`: nada a bloquear). Garantido pelos passos dos comandos **e** por hooks do Claude Code que cobrem `gh`, `az repos pr` e, se o hook alcançar, as ferramentas do MCP (a verificar).
- **RN-17** Caminho de menor custo em tokens: o adaptador prefere, por VCS, o caminho que o dono indicou (GitHub: `gh`; Azure DevOps: MCP do Azure DevOps com autenticação interativa, `az repos` como alternativa) e registra o critério: saída curta (`--json`/`--query`) e sem carregar definições de MCP desnecessárias. A medição real de tokens por caminho é **a verificar** (task 0.1).
- **RN-8** Todo achado do revisor vira uma lição: um arquivo por lição em `sdd/memory/licoes/<id>.md`, com a rule violada e como evitar. Discovery, nova-feature, novo-fix e Builder leem só as lições da rule que vão tocar (`licoes.py --rule COD-02`). Nunca um arquivo único que cresce.
- **RN-9** CI: `concurrency` com cancelamento só em PR; matriz de 3 sistemas só em PR pronto e na `main`, nunca em draft; alerta de gasto e checagem mensal por script, sem token no git (`SEG-04`).
- **RN-10** Linux, Windows e macOS; scripts Python com `uv` e `pathlib`; cada ferramenta com um prompt "cole isto na sua IA" curto no README.
- **RN-11** Todo handoff ao humano é pergunta com opções (AskUserQuestion no Claude Code, `Select` no CLI): [Seguir]/[Ajustar], [Continuar]/[Ajustar], merge. Nunca texto solto.
- **RN-12** A validação final é a mais determinística possível: um script local, sem LLM, confere o que é checável (rules por ID citadas e aplicáveis, ADR citada e vigente, estrutura da spec, lições lidas, fila vazia, SHA do veredito) e entrega a lista do que sobrou para julgamento; o LLM só trata essa lista. O script **não roda na Action**.
- **RN-13** A Action só publica. Antes, um job leve (ubuntu, segundos) confere o **atestado**: veredito conforme do revisor para o SHA do HEAD **mais** a aprovação do dono (ou o modo sem humano declarado). Sem atestado válido, bloqueia. A prova da aprovação deve ser algo que o agente não emite sozinho (Decisão vi).
- **RN-14** Modo sem humano: opcional, escolhido no **início** da sessão, padrão com humano. Antes de ativar o Forge informa os riscos: nenhuma pessoa vê o código antes de o PR sair do draft; erro de julgamento do revisor passa direto; o custo de tokens fica sem teto humano. O modo é gravado numa linha do `status.md` (teto de 1 KB, P-21) e no atestado.
- **RN-15** Dois pontos de revisão continuam os únicos: o novo-fix aberto pelo revisor e o retorno ao Builder fazem parte do ponto 2 e só reabrem a revisão quando o sinal `pronto-para-revisao` se repetir. O teto de rodadas vale por ponto (RN-5).
- **RN-16** Fix de código já desenvolvido tem prioridade sobre feature nova na fila do Builder. O Builder paraleliza o que não tem dependência (features e fixes).

## O que muda na ordem (conflito com a 100)
Hoje: veredito conforme, depois PR criado sem draft. `forge queue review` chama `queue.OpenPR` (`cmd/forge-sdd/queue.go:195`), que roda `gh pr create` (`internal/queue/pr.go:61,105`); o Builder do pipeline chama `pipeline.OpenPR` (`internal/pipeline/pr.go:71`). Agora: o PR nasce na concepção; o veredito conforme só habilita "fechar o PR?" e então "marcar pronto" (adaptador de VCS). O que quebra está em `criteria.md` (`O que já existe`) e na task 1.7 do plano.

## Escopo
**Dentro:** ADR (rascunho abaixo; arquivo só depois do Gate 1); passos de `/discovery`, `/nova-feature`, `/novo-fix`, `/revisar` e `/implementar` e das skills de concepção, build e governança; veredito com `head_sha` e `ponto`; sinais no `status.md`; hooks `PreToolUse` e instalador; `licoes.py` e `sdd/memory/licoes/`; modelo de workflow; `ci-politica.py`; READMEs com prompts "cole isto na sua IA"; templates, goldens e `/upgrade-sdd`; primeira leva de repositórios (Decisão iv).
**Dentro (reaberto de novo, VCS agnóstico):** adaptador de VCS (`internal/vcs`: `github`, `azure-devops`, `nenhum`) lido da Constituição; `pr_guard.py` para `gh`, `az repos pr` e MCP (a verificar); workflow modelo também para Azure Pipelines (a verificar); PR de teste em cada VCS suportado. As decisões 1 a 5 e vi a ix valem como aprovadas, com a leitura GitHub-específica reescrita (job "pronto?", Actions, Environment são o caso `github`).
**Dentro (reaberto pelo dono):** script local de validação determinística; job de verificação do atestado; modo sem humano com aviso de riscos; split e implementação pelo Builder sem nova aprovação do split; novo-fix aberto pelo revisor com prioridade; paralelismo do Builder (features e fixes).
**Fora:** pagar ou alterar billing; migrar de CI; mexer em repositório de cliente sem aprovação; reescrever o laço do pipeline (104) ou o CLI (105); editar a 104 ou a 105; criar o arquivo da ADR agora.

## Riscos
- **R1 Draft mesclado sem CI.** Mitigação: RN-7. Job "pronto?" é a Decisão iii.
- **R2 Loop de revisão.** Mitigação: RN-3, RN-4 e teto de 2 rodadas; o teto atual é 3 (`internal/config/forge.go:192`).
- **R3 Aprovação velha.** O veredito persistido hoje não tem SHA (`harness.Parecer` = `veredito`, `achados`, `candidate_skill`, `resumo`; `internal/harness/parecer.go:16-21`). Task 1.1.
- **R4 Repositório pessoal privado sem draft.** Fallback: rótulo `wip` e `if: !contains(github.event.pull_request.labels.*.name, 'wip')`. Decisão ii.
- **R5 Colisão de ADR.** Número "próxima livre" só no merge.
- **R6 Hooks só valem no Claude Code.** Gemini e Copilot dependem dos passos dos comandos; documentar.
- **R7 Goldens e `/upgrade-sdd`.** Muda arquivos que o usuário recebe (`TST-02`).
- **R8 `uv` nos repositórios-alvo e no Windows:** não verificado; o `doctor` precisa checar.
- **R10 Atestado forjável (segurança).** Um arquivo no repositório pode ser escrito pelo agente. Se o agente roda com a credencial do dono (`gh`, `az` ou MCP autenticado), qualquer prova emitida pela API do VCS com ela também é emitível por ele; só resiste o que exige algo fora do alcance do agente (clique no navegador, chave com senha ou hardware). Opções na Decisão vi.
- **R11 Split sem aprovação humana.** O Builder pode fatiar fora do escopo aprovado. Mitigação: o script local confere que cada task cita o `plan.md` aprovado e que nenhuma sai dele; a Revisão 2 revisa de novo.
- **R12 Modo sem humano.** Ninguém vê o código antes de o PR sair do draft; erro do revisor passa; custo sem teto humano (RN-14).
- **R13 Loop fix-revisão.** O teto atual: `ApplyVerdict` só devolve se `job.Attempt < job.MaxReworkAttempts` (`internal/pipeline/review.go:132`) e o padrão é 3 (`internal/config/forge.go:192`): 3 passadas e 2 devoluções (leitura do código; não rodei). O "teto de 2 rodadas" desta política **coincide** com o padrão; não é preciso baixar para 2 (corrijo minha versão anterior, que dizia "padrão 2"). Já a devolução hoje cria `rework` com `Priority = high` (`review.go:138`), mas é nova tentativa do mesmo job, não um novo-fix. O novo-fix passa a ser o artefato do desvio e reusa o mesmo `Attempt`. Ao estourar, `ActionEscalated` (`review.go:157`) notifica o dono. Para a spec há outro contador (`nextSpecAttempt`, `internal/pipeline/spec.go:31`).
- **R9 PRs de task do leque** (base na branch da iniciativa, `internal/pipeline/goal.go:375`): a regra de draft vale para eles? Está em `status.md` (A confirmar).

## Rascunho da ADR (não criar o arquivo agora)
**Número:** próxima livre depois das ADRs da 104 (0016 a 0020) e da 105; atribuir no merge. **Título:** Política de PR em draft, revisão em dois pontos e CI só ao finalizar.
- **Contexto:** Actions do GitHub bloqueadas por cobrança (o gatilho); consumo evitável; revisão repetida; o auto-PR da 100 cria PR sem draft; o fluxo chama `gh` direto em dois lugares, mas a Constituição já prevê `azure-devops` e `nenhum`.
- **Decisão do VCS (correção do dono):** a política é independente de VCS. Um **adaptador de VCS** (novo pacote `internal/vcs`, com `github`, `azure-devops` e `nenhum`) lê o valor de `VCS / Work Item System` e expõe as operações criar PR em draft, procurar PR aberto, marcar pronto e mesclar; `queue.OpenPR` e `pipeline.OpenPR` passam a delegar a ele (hoje duplicam a mesma lógica). Onde o agente usa MCP, a skill orienta a chamada e o hook confere. **Impacto na Constituição:** o campo `VCS / Work Item System` e seus valores **não mudam**; muda só a nota da linha 30 (hoje: "use `az repos pr create` no lugar de `gh pr create`"), que passa a remeter ao adaptador e a incluir draft, pronto e o MCP do Azure DevOps. É texto, mesmo assim vai pela ADR e por `GOV-01`, na task 2.6; não editei a Constituição.
- **Decisão:** RN-1 a RN-16. Além do que segue: Gate 1 vira pergunta [Seguir]/[Ajustar] e o split passa ao Builder sem nova aprovação (altera `FLX-01`, ver abaixo); revisor abre novo-fix com prioridade; validação final por script local sem LLM; Action só confere o atestado; modo sem humano opcional com aviso de riscos. O PR nasce em draft na concepção; o revisor atua em dois pontos com sinais no `status.md`; o veredito guarda o SHA; `ready` e `merge` só com veredito conforme do SHA atual; lições em arquivos pequenos; CI com `if` de draft, `concurrency` e matriz só em PR pronto e `main`.
- **Consequências:** menos minutos e menos repetição de erro; custo em hooks, goldens e migração. O Reviewer deixa de criar PR. Estende ADR-0004, ADR-0012 e ADR-0013 sem substituí-las. Exige ajustar `GOV-03` (cita `lessons.md`) e `FLX-01` via `GOV-01` (a Constituição não é editada nesta discovery).
- **Impacto exato no `FLX-01`** (`sdd/memory/constitution.md:98`; também citado em `.agents/skills/sdd-concepcao/SKILL.md` regra 6, `references/discovery.md` passo 11 e `.agents/skills/sdd-planejamento/references/split-features.md` passos 2, 8 e 9). Redação atual: "Spec antes do código: nada é implementado nem enfileirado sem Spec revisada; mudança de escopo revalida a Spec." **Redação APROVADA pelo dono em 2026-11-20** (a Constituição continua sem edição; entra na task 2.6 via `GOV-01`): "Spec antes do código: nada é implementado nem enfileirado sem Spec revisada e aprovada pelo humano no Gate 1 ([Seguir]/[Ajustar]); com o Gate 1 aprovado o Builder faz o split e enfileira sem nova aprovação, só dentro do escopo do `plan.md` aprovado; mudança de escopo revalida a Spec." Não muda o Gate 1; remove só a segunda confirmação humana do `/split-features`. No código, `forge dispatch` hoje recusa iniciativa sem tasks (`internal/pipeline/dispatch.go:130`: "rode /split-features"), então o despacho precisa aceitar job sem tasks e o split vira passo do Builder.
- **Convivência com `Job.Attempt` e o teto:** ver R13.
- **Alternativas descartadas:** `paths-ignore` (não distingue draft); `workflow_dispatch` manual (perde o gatilho); revisão por task (loop e custo).

## Decisões para o Gate 1
1. **(i) Prefixo da branch.** Evidência: `layout.InitiativeBranch` devolve `"feat/" + id` (`internal/layout/layout.go:340`) e é usada em `pipeline/dispatch.go:221`, `pipeline/archive.go:194`, `pipeline/goal.go:375,405` e `queue/base.go:118`. `FromBranch` só reconhece `feat` e `fix` (`branchPrefixes`, `layout.go:163`): uma branch `init/…` **não seria vista** por `forge next-id` (que lê refs via `idOfRef`, `scan.go:322`) nem por `Current`, e dois IDs iguais poderiam nascer. O gate P-01 só bloqueia `main`/`master`, mas a remediação cita `feat/XX-nome` (`internal/gate/gate.go:207-217`). A regra 2 da Constituição fixa `feat/*` e `fix/*`; a regra 15 já diz `<prefixo>/<id>-<nome>` (`sdd/memory/constitution.md:36,49`). Trocar mexe nesses 5 pontos do Go, no texto de P-01, na Constituição (`GOV-01`), em goldens e nas branches abertas (104, 105). **Recomendo manter `feat/`** nesta iniciativa: é nomenclatura, não política de CI, e um prefixo novo merece iniciativa própria com ADR. Se o dono preferir `init/`, vira uma task extra na onda 2.
2. **(ii) Repositório pessoal privado aceita draft?** hub-busize aceita (apurado pelo dono). **Recomendo** a task 3.5 (PR de teste `--draft` em cada repositório da primeira leva) antes de aplicar; fallback `wip`.
3. **(iii) Job leve "pronto?" que falha em draft.** **Recomendo sim, só no check obrigatório da `main`:** custa segundos de runner e fecha o R1 no lado do GitHub, onde o hook não alcança (merge pela web).
4. **(iv) Primeira leva.** **Recomendo** hub-busize, hub-enterprise, forge-sdd e gerador-videos: são os citados pelo dono; o insumo diz que hub e gerador-videos já têm `concurrency` (não reverificado).
5. **(v) Migrar o `lessons.md` atual?** Hoje são 2 entradas e `forge lessons add` com orçamento de 2 KB (`cmd/forge-sdd/lessons.go:14-16`). **Recomendo migrar** (2 arquivos em `licoes/`; `lessons add` grava lá; `GOV-03` ajustada): é o princípio do dono e o custo é mínimo agora.

Decisões 1 a 5 aprovadas e mantidas. Gate 1 reaprovado pelo dono em 2026-11-20 com: **FLX-01 nova redação aprovada; (vi) aprovada a opção (e) Environment com revisor obrigatório, com veredito e SHA conferidos pelo job, fallback (d) commit/tag assinado se o plano não permitir (a task 3.5 prova Environments em repositório pessoal privado); (vii) aprovado manter 3 (2 devoluções); (viii) assinatura do projeto Forge SDD (nome "Forge SDD", link forge-sdd.org, bot do Forge sem armas), não a pessoal; (ix) aprovado o BPMN simplificado.** O texto abaixo guarda as opções e os motivos:

6. **(vi) Prova do atestado** (o agente não pode emiti-la sozinho). Limite geral: se o agente usa a credencial `gh` do dono, tudo que passa pela API do GitHub com ela é forjável; só resiste o que fica fora do alcance dele. Nada abaixo foi testado.
   - **a) Arquivo de atestado no repositório.** Pró: trivial, funciona em qualquer plano. Contra: o agente escreve e commita; **descartar** como prova única.
   - **b) Review `APPROVED` do GitHub no SHA do HEAD** (`commit_id`), conferido pelo job via API. Pró: nativo, atrelado ao SHA. Contra: o GitHub não deixa o autor aprovar o próprio PR; se o agente abre o PR com a conta do dono, o dono não consegue aprovar; exige segunda conta.
   - **c) Label aplicada pelo dono** (ex.: `atestado`). Pró: simples, um clique. Contra: label não prende ao SHA (precisa de workflow que a remove no `synchronize`) e o agente com o mesmo token também rotula.
   - **d) Commit ou tag assinado** com chave do dono (conferido via `verification.verified` ou chave pública fixada). Pró: prende ao SHA, forte se a chave estiver fora do agente (senha ou hardware). Contra: fricção, mais difícil no Windows, e a lista de chaves aceitas no repositório precisa de proteção (CODEOWNERS/branch protection).
   - **e) GitHub Environment com revisor obrigatório** para o job de publicação. Pró: a aprovação é um clique no navegador que o CLI do agente não emite; o job só continua depois. Contra: Environments com revisores em repositório privado dependem do plano (não verificado); prende ao run, não ao SHA (o job confere o SHA do veredito).
   - **Recomendo (e)** como prova principal, mais o veredito com SHA conferido pelo job; onde o plano não oferecer, **(d)** com chave fora do agente. Sem humano (modo RN-14), só vale o veredito do revisor, que é forjável por natureza: o risco é aceito pelo dono ao ativar o modo.
7. **(vii) Teto de rodadas com novo-fix:** manter o padrão atual (`max_rework_attempts` = 3, ou seja 2 devoluções) e contar o novo-fix na mesma conta do job. **Recomendo manter**; ao estourar, pergunta ao dono. Alternativa: contador separado por fix (mais permissivo, mais custo).
8. **(viii) Assinatura dos diagramas da 106:** padrão do plugin ou nenhuma? A outra sessão diz que o dono aprovou a padrão; não consegui confirmar (na 118 ele escolheu "sem assinatura"). Resolvida: assinatura do projeto (Forge SDD, forge-sdd.org, avatar `diagramas/forge-sdd-avatar.png`, desenho de `MascotFallbackSvg.astro`, cor #8aa4ff, marca "bar"); sem nome nem LinkedIn pessoais.
9. **(ix) Formato do BPMN.** O BPMN simplificado (gateways como caixas "◇", eventos como pílulas, 4 raias) mostra quem faz o quê e os dois gates, que é o que o Gate 1 precisa. Não mostra o teto de rodadas como evento de escalonamento, o script local nem a escolha do modo. **Recomendo manter o simplificado** na spec e só migrar para BPMN completo (eventos, escalonamento) se a ADR final pedir um diagrama normativo. A legenda no rodapé cobre o teto e o modo.
11. **(xi) Prova do atestado por VCS** (novo; não troca a (vi) aprovada para o GitHub). GitHub: (e) Environment com revisor, fallback (d). Azure DevOps: candidatos (1) aprovação em Environment do Azure Pipelines com aprovadores, (2) política de branch com revisores obrigatórios e build validation, (3) commit assinado. **Recomendo (1)**, por ser o análogo direto do Environment do GitHub, sujeito às provas da task 3.5; só vale se o Azure permitir que o agente não emita a aprovação (a verificar). `nenhum`: ver (xii).
12. **(xii) `nenhum`: aceitar atestado só local** (veredito com SHA e aprovação no chat)? **Recomendo sim, com aviso explícito** dos riscos (sem prova externa, igual ao modo sem humano) gravado no `status.md` e no veredito.
13. **(xiii) Critério de custo em tokens** entre MCP e CLI: manter a preferência do dono (`gh`; MCP do Azure DevOps com `az` de alternativa) e medir na task 0.1. **Recomendo** registrar a medição antes de fixar o padrão por VCS.
(Itens 6 a 9 e o `FLX-01` foram aprovados; 11 a 13 são novos e reabrem o Gate 1 para o escopo VCS agnóstico.) PRs de task do leque em draft segue em `status.md` (A confirmar); o dono precisa reaprovar o escopo antes do `/split-features`.

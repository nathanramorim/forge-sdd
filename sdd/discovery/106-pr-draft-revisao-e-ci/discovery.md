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

## Jornada alvo
1. `/discovery`, `/nova-feature` ou `/novo-fix`: cria a branch, commita a spec, empurra e abre o PR em draft, sem pedir. Grava `spec-para-revisao` no `status.md`.
2. **Revisão 1 (spec):** o revisor valida a estrutura da spec contra ADRs, Constituição, rules e padrão do projeto e devolve o veredito ao humano.
3. Com a spec aprovada (Gate 1), o Builder trabalha com commits locais, sem revisor no meio.
4. Fila da onda vazia e checagem local verde: `pronto-para-revisao`.
5. **Revisão 2 (entrega):** o revisor confere se a entrega cumpre a spec e as mesmas definições.
6. Só com veredito conforme em tudo, para o SHA do HEAD, aparece o menu "fechar o PR?". Sim: `gh pr ready` (o CI roda uma vez). Com checks verdes, o menu do merge (decisão do humano).

## Regras de negócio
- **RN-1** O PR nasce em draft (`gh pr create --draft`), sem perguntar, nos três comandos de concepção.
- **RN-2** Em draft o CI não roda (jobs pulados). Roda uma vez no `ready_for_review` e na `main`.
- **RN-3** Revisão em exatamente dois pontos: antes do Builder (spec) e depois (entrega). Nunca no meio, nem por task entregue.
- **RN-4** Cada ponto tem um sinal determinístico no `status.md`: `spec-para-revisao` e `pronto-para-revisao` (este exige fila da onda vazia e checagem local verde). A revisão só se repete quando o sinal do ponto se repetir.
- **RN-5** O revisor devolve o veredito ao humano. Ajuste volta a quem originou (discovery, nova-feature, novo-fix ou Builder). Teto de 2 rodadas automáticas por ponto; depois pergunta ao humano.
- **RN-6** "Fechar o PR?" só aparece com veredito conforme em tudo. O veredito guarda o SHA do HEAD; commit novo o invalida.
- **RN-7** `gh pr create` só com `--draft`; `gh pr ready` e `gh pr merge` só com veredito conforme do SHA atual. Garantido pelos passos dos comandos **e** por hooks do Claude Code.
- **RN-8** Todo achado do revisor vira uma lição: um arquivo por lição em `sdd/memory/licoes/<id>.md`, com a rule violada e como evitar. Discovery, nova-feature, novo-fix e Builder leem só as lições da rule que vão tocar (`licoes.py --rule COD-02`). Nunca um arquivo único que cresce.
- **RN-9** CI: `concurrency` com cancelamento só em PR; matriz de 3 sistemas só em PR pronto e na `main`, nunca em draft; alerta de gasto e checagem mensal por script, sem token no git (`SEG-04`).
- **RN-10** Linux, Windows e macOS; scripts Python com `uv` e `pathlib`; cada ferramenta com um prompt "cole isto na sua IA" curto no README.
- **RN-11** Todo handoff ao humano é menu ("fechar o PR?", "mesclar?"), nunca texto solto.

## O que muda na ordem (conflito com a 100)
Hoje: veredito conforme, depois PR criado sem draft. `forge queue review` chama `queue.OpenPR` (`cmd/forge-sdd/queue.go:195`), que roda `gh pr create` (`internal/queue/pr.go:61,105`); o Builder do pipeline chama `pipeline.OpenPR` (`internal/pipeline/pr.go:71`). Agora: o PR nasce na concepção; o veredito conforme só habilita "fechar o PR?" e então `gh pr ready`. O que quebra está em `criteria.md` (`O que já existe`) e na task 1.7 do plano.

## Escopo
**Dentro:** ADR (rascunho abaixo; arquivo só depois do Gate 1); passos de `/discovery`, `/nova-feature`, `/novo-fix`, `/revisar` e `/implementar` e das skills de concepção, build e governança; veredito com `head_sha` e `ponto`; sinais no `status.md`; hooks `PreToolUse` e instalador; `licoes.py` e `sdd/memory/licoes/`; modelo de workflow; `ci-politica.py`; READMEs com prompts "cole isto na sua IA"; templates, goldens e `/upgrade-sdd`; primeira leva de repositórios (Decisão iv).
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
- **R9 PRs de task do leque** (base na branch da iniciativa, `internal/pipeline/goal.go:375`): a regra de draft vale para eles? Está em `status.md` (A confirmar).

## Rascunho da ADR (não criar o arquivo agora)
**Número:** próxima livre depois das ADRs da 104 (0016 a 0020) e da 105; atribuir no merge. **Título:** Política de PR em draft, revisão em dois pontos e CI só ao finalizar.
- **Contexto:** Actions bloqueadas por cobrança; consumo evitável; revisão repetida; o auto-PR da 100 cria PR sem draft.
- **Decisão:** RN-1 a RN-11. O PR nasce em draft na concepção; o revisor atua em dois pontos com sinais no `status.md`; o veredito guarda o SHA; `ready` e `merge` só com veredito conforme do SHA atual; lições em arquivos pequenos; CI com `if` de draft, `concurrency` e matriz só em PR pronto e `main`.
- **Consequências:** menos minutos e menos repetição de erro; custo em hooks, goldens e migração. O Reviewer deixa de criar PR. Estende ADR-0004, ADR-0012 e ADR-0013 sem substituí-las. Exige ajustar `GOV-03` (cita `lessons.md`) via `GOV-01`.
- **Alternativas descartadas:** `paths-ignore` (não distingue draft); `workflow_dispatch` manual (perde o gatilho); revisão por task (loop e custo).

## Decisões para o Gate 1
1. **(i) Prefixo da branch.** Evidência: `layout.InitiativeBranch` devolve `"feat/" + id` (`internal/layout/layout.go:340`) e é usada em `pipeline/dispatch.go:221`, `pipeline/archive.go:194`, `pipeline/goal.go:375,405` e `queue/base.go:118`. `FromBranch` só reconhece `feat` e `fix` (`branchPrefixes`, `layout.go:163`): uma branch `init/…` **não seria vista** por `forge next-id` (que lê refs via `idOfRef`, `scan.go:322`) nem por `Current`, e dois IDs iguais poderiam nascer. O gate P-01 só bloqueia `main`/`master`, mas a remediação cita `feat/XX-nome` (`internal/gate/gate.go:207-217`). A regra 2 da Constituição fixa `feat/*` e `fix/*`; a regra 15 já diz `<prefixo>/<id>-<nome>` (`sdd/memory/constitution.md:36,49`). Trocar mexe nesses 5 pontos do Go, no texto de P-01, na Constituição (`GOV-01`), em goldens e nas branches abertas (104, 105). **Recomendo manter `feat/`** nesta iniciativa: é nomenclatura, não política de CI, e um prefixo novo merece iniciativa própria com ADR. Se o dono preferir `init/`, vira uma task extra na onda 2.
2. **(ii) Repositório pessoal privado aceita draft?** hub-busize aceita (apurado pelo dono). **Recomendo** a task 2.6 (PR de teste `--draft` em cada repositório da primeira leva) antes de aplicar; fallback `wip`.
3. **(iii) Job leve "pronto?" que falha em draft.** **Recomendo sim, só no check obrigatório da `main`:** custa segundos de runner e fecha o R1 no lado do GitHub, onde o hook não alcança (merge pela web).
4. **(iv) Primeira leva.** **Recomendo** hub-busize, hub-enterprise, forge-sdd e gerador-videos: são os citados pelo dono; o insumo diz que hub e gerador-videos já têm `concurrency` (não reverificado).
5. **(v) Migrar o `lessons.md` atual?** Hoje são 2 entradas e `forge lessons add` com orçamento de 2 KB (`cmd/forge-sdd/lessons.go:14-16`). **Recomendo migrar** (2 arquivos em `licoes/`; `lessons add` grava lá; `GOV-03` ajustada): é o princípio do dono e o custo é mínimo agora.

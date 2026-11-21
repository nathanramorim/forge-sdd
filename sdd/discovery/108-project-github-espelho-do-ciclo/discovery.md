# Discovery 108 — GitHub Project como espelho do ciclo do forge-sdd

## Pedido do dono (resumido, 20/11/2026)
Criar um GitHub Project dentro do projeto forge-sdd com as colunas **Backlog, Spec, Spec-validation, Todo, Building, Validation, Done**, que representam o ciclo do forge-sdd. Serve para "gestão à vista": cada agente deve saber andar nas etapas ali e, quando o dono mexer no quadro, "cada um sabe o que fazer". Hoje o controle da fila é em arquivo; o Project será o **espelho**: "o que reflete lá tem que refletir aqui e aqui lá" (duas vias). Gerar esta Discovery e **preparar** as issues das coisas em aberto (iniciativas 105, 106, 107 e os fixes 97-01 e 104-14), vinculando o PR de cada uma. A descrição da issue é **comercial** (sem jargão); o detalhe técnico fica na spec e a issue traz o link para a pasta.

## Por que o ID é 108
O dono pediu 108 de propósito: 105, 106 e 107 já existem em PRs abertos (#121, #120, #122). Nesta branch `forge next-id` devolveu `109` (rodado em 20/11/2026), então o número 108 não vem do comando; fica registrado aqui para ninguém reusar. Esta sessão não criou nada no GitHub.

## Clarify (`sdd/memory/clarify.md`)
Sinais encontrados: (1) o que o arrastar do humano autoriza (escopo com mais de uma leitura); (2) dependência externa (API do GitHub Projects, permissões); (3) o critério de "espelho em duas vias" não diz quem vence no conflito. Não perguntei antes de escrever: o pedido traz critério e escopo; as lacunas viraram `## Decisões para o Gate 1`, cada uma com recomendação, sem decidir pelo dono.

## Evidência do que já existe (comandos e arquivos lidos)
- **Ligação com gerenciador de projeto já prevista, mas só como texto para o agente.** `internal/config/forge.go:91-98` define `PMConfig` (`provider`, `supported`, `require_existing_card`, `status_on_implement`, `status_on_pr_approved`, `evidence_max_age_minutes`); padrão em `forge.go:183-190`: provedores `clickup`, `github-projects`, `azure-devops`, `jira`, `require_existing_card: false`, "In Progress" e "Ready for Review". A validação está em `forge.go:337-346`.
- **O que o código faz com card:** `internal/pipeline/dispatch.go:200-262` só exige `card-evidence.json` quando `require_existing_card` é `true`; `loadEvidence` (`dispatch.go:431-482`) valida esquema, frescura, provider e URL; `dispatch.go:260-261` e `internal/pipeline/review.go:107` apenas **devolvem uma frase** (`CardInstruction`: "Atualize o card ... via MCP"). Nenhum código chama a API de um quadro, lê coluna ou move cartão. `grep` de `StatusOnPRApproved` fora de `config` acha só `review.go:107`.
- **Não há `.forge/config/forge.json` neste checkout** (`ls .forge/config` falhou), então o modo pipeline está desligado aqui.
- **Constituição:** `sdd/memory/constitution.md:25` `| VCS / Work Item System | github |` e `:27` (com `azure-devops` usa `az repos pr`; com `nenhum`, branch pronta sem comandos de VCS). `sdd/memory/mcps.md` lista só `context7` e `git`: **nenhum MCP de gerenciador de projeto** está registrado.
- **Skill:** `.agents/skills/sdd-governanca/references/pr-review.md:13`, passo 5: "mova o card pelo MCP para o status indicado".
- **Fila em arquivo:** `internal/queue/queue.go:16-23` define os estados `queued`, `building`, `review_pending`, `approved`, `changes_requested`, `blocked`. `Complete` (`queue.go:464-480`) leva `building` para `review_pending`; `Review` (`queue.go:488`) leva a `approved`, `changes_requested` ou de volta a `review_pending`. `SpecStatus` assume `reviewed` (`queue.go:88,101,226`). `in_progress` e `completed` **não existem** como constantes em `internal/queue` (o pedido os citou como exemplo); `completed` aparece apenas no filtro da 105/1.5 (PR #124, ainda não mesclado). Jobs do barramento: `internal/bus/job.go:33-37` (`feature_build`, `pr_rework`, `pr_review`, `spec_review`). `forge queue` tem `enqueue`, `next`, `complete`, `review`, `list`, `panel`, `activity`.
- **Fatos do `gh` (somente leitura, 20/11/2026):** `gh auth status` mostra o escopo `project`; `gh project list --owner nathanramorim` lista os projetos 1 a 3 (não relacionados, intocados); `gh project --help` lista `close`, `copy`, `create`, `delete`, `edit`, `field-create`, `field-delete`, `field-list`, `item-add`, `item-archive`, `item-create`, `item-delete`, `item-edit`, `item-list`, `link`, `list`, `mark-template`, `unlink`, `view`. **Não existe `field-edit`**: o CLI não altera as opções do campo "Status" padrão. `field-create --data-type SINGLE_SELECT --single-select-options a,b,c` cria um campo novo. `item-edit` exige `--project-id`, `--field-id`, `--single-select-option-id`, `--id` (IDs que vêm de `field-list --format json` e `item-list --format json`; `item-list` expõe a etapa em `.status`). `project create` não tem `--visibility`; `project edit --visibility {PUBLIC|PRIVATE}` tem. `gh issue list` neste repositório está vazio.

## Sinais das outras iniciativas (lidos em branches de PR; nada editado)
- **106** (`origin/feat/106-pr-draft-revisao-e-ci`, `discovery.md`, `plan.md`, `status.md`): sinais `spec-para-revisao` e `pronto-para-revisao` no `status.md` (RN-4, task 1.6); veredito com `head_sha` (task 1.1); atestado por PR (RN-13, modo `desligado|local|vcs`); estado por feature em arquivo na branch da feature (task 0.7); `internal/vcs` com interface `Adaptador` (task 0.2; **não existe em `main`**); split pelo Builder sem nova aprovação (RN-22, task 2.3). Gate 1 aprovado (`status.md`).
- **107** (`origin/feat/107-orquestrador-de-sessoes`): orquestrador com WIP humano (limite 3, campo `WIP humano` na Constituição) e entrada sempre tratada como dado. Gate 1 aprovado.
- **105** (`origin/feat/105-experiencia-cli-e-harness`): menus e `forge status` temáticos; task 2.7 (descrição de negócio em toda spec) e 3.6 (log de achados). Gate 1 aprovado; onda 1 em 7 PRs de código em draft.

## Problema
O estado do trabalho vive em arquivos espalhados (`status.md` por iniciativa, `queue.json`, PRs). Quem quer saber "o que está em cada etapa" precisa rodar `forge status` ou abrir pastas; o dono não tem uma visão única, e os agentes não têm um lugar comum para ver "onde estamos". O campo `pm_mcp` já prevê um card, mas nenhum código o consulta nem o move: é só uma frase para o agente.

## Para quem
O dono do projeto (decide gates e merge; quer gestão à vista) e os agentes (`pm-architect`, `builder`, `reviewer`, o orquestrador da 107) que precisam saber a etapa de cada coisa e o que fazer nela. Em segundo plano, quem acompanha o projeto aberto no GitHub.

## Jornada alvo
1. [IA pm-architect] `/discovery` abre a spec: a issue-mãe entra no Project na coluna **Spec**.
2. [AUTOMÁTICO] Spec pronta para revisão (`spec-para-revisao`): o forge move o cartão para **Spec-validation**.
3. [HUMANO] Gate 1 aprovado em menu: o cartão **continua em Spec-validation** até o PR da spec ser mesclado na principal; só então vai para **Todo** (fila `queued`).
4. [IA builder] **Começa sozinho** assim que a spec é mesclada (decisão D1 do dono): o cartão vai de Todo para **Building**, dentro do WIP humano e do teto de tentativas; cada feature tem o seu cartão e o seu PR. **Via rápida (D2):** item só de `adr`, `rule` e/ou `documentação` (sem `código`) vai direto de Todo para **Validation**, sem passar por Building.
5. [AUTOMÁTICO] `pronto-para-revisao` ou `review_pending`: o cartão vai para **Validation** (WIP humano da 107 conta os cartões aqui).
6. [HUMANO] [Continuar] e merge do PR: o cartão vai para **Done**.
7. [HUMANO] Se o dono arrasta um cartão no quadro, o forge percebe, valida e **pede confirmação em menu** antes de gravar no arquivo.

## Regras de negócio
- **RN-1** O Project é **espelho**, não segundo controle: o arquivo no repositório continua sendo a fonte da verdade do estado (recomendação; decisão do dono, item 1 do Gate 1).
- **RN-2** O campo de etapa tem exatamente 7 valores, na ordem do pedido. Qualquer outro valor é recusado (lista fechada).
- **RN-3** Texto de issue, título, comentário e corpo de PR é **entrada não confiável**: nenhum agente obedece a esse texto; só o valor do campo de etapa, validado contra a lista fechada, é lido.
- **RN-4** Mover cartão é idempotente: repetir o mesmo movimento não muda nada e não gera ruído.
- **RN-5** Arrastar o cartão é um **pedido**, não uma aprovação. Gate 1 e [Continuar] só valem com o menu do forge (e, no modo `vcs` da 106, com o atestado). O ator do movimento é conferido; ator desconhecido vira divergência, nunca aprovação.
- **RN-6** Divergência arquivo versus quadro é detectada por `forge doctor` e mostrada; não é corrigida em silêncio.
- **RN-7** Agnóstico de VCS (decisão da 106): `github` usa GitHub Projects; `azure-devops` usa o Azure Boards (**a verificar**); `nenhum` não tem quadro e o espelho é só o arquivo.
- **RN-8** Issue publicada traz descrição comercial; o detalhe técnico fica na spec e a issue só leva o link. Repositório público: nada de cliente, cobrança, caminho local ou nome pessoal.
- **RN-9** Nada é criado no GitHub sem `--executar` explícito; o padrão do script é `--dry-run`.
- **RN-10** O quadro nasce **privado**; abrir ao público é decisão posterior do dono.
- **RN-12** (nova, 20/11/2026) O título da issue leva um prefixo derivado das etiquetas (função pura); o título base é a fonte; achar a issue por título nunca cria duplicata (renomeia).
- **RN-13** (D1, dono, 20/11/2026) O Builder **começa sozinho** quando a spec entra na principal. O gatilho é o **estado em arquivo mais o merge verificado**, nunca o texto do cartão. Freios: WIP humano da 107 (limite 3) e teto de 3 tentativas (`max_rework_attempts`); o início automático só vale se o dono escolheu o modo com humano pelo menos até o Gate 1 (106).
- **RN-14** (D2) **Via rápida:** item cujos arquivos classificam (por `ClassificaEtiquetas`, nunca pelo texto do cartão) só como `adr`, `rule` e/ou `documentação`, sem `código`, vai de Todo direto para Validation. O PR continua sendo conferido pelo dono. Mudança de `rule` continua exigindo ADR, `GOV-01` e a Conferência 1 do reviewer. Se o PR de via rápida passar a conter `código`, o cartão volta a Building e o dono é avisado.
- **RN-15** (D3) Quem fecha a issue e move o cartão para Done são as **automações do próprio Project** (escolha do dono). O forge não fecha nada nesse caso, mas a divergência arquivo versus quadro aparece no `forge doctor` (RN-6).
- **RN-11** (nova, pedido do dono em 20/11/2026) Todo PR e toda issue carregam **etiquetas** que dizem se o item contém código, ADR, rule, documentação ou é só especificação. Etiquetas, não um campo do Project (decisão do dono no chat). O cálculo é por caminho dos arquivos (heurística, sem ler conteúdo).

## Escopo
**Dentro:** mapeamento estado-arquivo para etapa; leitura somente leitura do quadro; mover cartão pelo CLI `forge project`; espelho nas duas vias com confirmação; verificação no `forge doctor`; reference única de rotina por agente; interface mínima no adaptador de VCS; issues preparadas e script de criação (rascunho, não executado).
**Fora:** criar o Project e as issues (isto é feito pelo dono, depois do Gate 1, com `criar-project.sh --executar`); webhooks e Actions (o CI está bloqueado por cobrança); Jira, ClickUp e Azure Boards além da interface; automação interna do Project (workflows) que mova cartões sozinha sem o forge; editar 105, 106 ou 107; relatório de economia (o ganho da 106 ainda não foi medido).

## Mapeamento estado do arquivo para coluna (resumo; completo em `criteria.md`)
| Coluna | Estado no arquivo | Evento que move |
|---|---|---|
| Backlog | ideia sem pasta de spec | alguém registra a ideia |
| Spec | pasta da discovery sem `spec-para-revisao` | `/discovery`, `/nova-feature`, `/novo-fix` abrem a spec |
| Spec-validation | spec aprovada pelo Gate 1 (ou `spec-para-revisao`, 106) e **ainda não mesclada** na principal; PR da spec aberto | pm-architect entrega a spec; menu do Gate 1 |
| Todo | spec já na principal e pronta para o Builder; itens `queued` | PR da spec mesclado na principal e Gate 1 aprovado |
| Building | item `building` ou `changes_requested` | Builder inicia sozinho depois do merge da spec (spec na principal, dentro do WIP humano 107 e do teto de 3 tentativas); reviewer devolve |
| Validation | trabalho pronto num PR aberto aguardando o dono conferir e mesclar (código, ADR, rule ou documentação); item `review_pending` ou `approved`; `pronto-para-revisao` (106) | Builder termina; veredito conforme |
| Done | PR mesclado | merge |
`blocked` não vira coluna nova: o cartão fica na etapa de origem com o campo `Bloqueado` (decisão do dono, item 6).

## Riscos
- **Falsificar o sinal de aprovação** (Gate 1, [Continuar]) arrastando ou editando o cartão: mitigado por RN-5 (menu de confirmação, ator, atestado).
- **Prompt injection** por título, comentário ou corpo de issue em repositório público: RN-3.
- **Divergência silenciosa** entre quadro e arquivo: RN-6.
- **Limites da API** (taxa, escopo `project`, projeto de usuário versus organização): a verificar; polling com intervalo mínimo.
- **Inflar contexto dos agentes** (P-24: Constituição 9189 de 9216 bytes; `sdd-concepcao/SKILL.md` 2045 de 2048): a rotina vive numa reference única e os comandos só apontam para ela.
- **Colisão de texto** com 105/2.7, 106/2.6 e 107 nas mesmas references: declarado em `plan.md` (`Sequenciar com`).
- **Início automático do Builder (D1).** Exige algo sempre ligado que perceba o merge da spec e inicie o Builder; **hoje nada faz isso** e a premissa anterior ("polling só quando roda um comando; nenhum processo vigia o quadro") deixa de bastar. Custo de tokens sem o dono presente: os freios são o WIP humano da 107 (limite 3), o teto de 3 tentativas e o modo/atestado da 106 (`desligado`, `local` padrão, `vcs`). O "[Seguir]" da 106 já autoriza o merge da spec e o início do Builder sem nova aprovação (`FLX-01` aprovado), então é coerente, desde que o dono esteja no modo com humano até o Gate 1. Segurança: texto de issue e comentário segue entrada não confiável.
- **Via rápida (D2):** a classificação é por caminho (heurística): um `.md` que na prática é código ou um PR que mistura tipos pode escapar do Building; mitigação: reclassificar a cada novo commit e voltar a Building se aparecer `código`.
- **Fechamento pelas automações do Project (D3).** (i) As automações (Settings > Workflows: "Item closed", "Pull request merged", "Auto-close issue") são ligadas na **interface**; a API GraphQL pública não as configura (**a verificar no Project real**), então o script não as liga e o dono precisa ligar à mão. (ii) O quadro tem **dois cartões por trabalho** (issue e PR): mesclar o PR move o cartão do PR para Done, mas o cartão da **issue** só vai para Done se a issue for **fechada**; os PRs atuais não usam palavra de fechamento (de propósito). (iii) O GitHub só fecha issue por palavra de fechamento quando o PR é mesclado na branch **padrão**; os PRs da onda 1 da 105 estão **encadeados** (base em outras branches), então mesclá-los não fecha a issue. (iv) Lacuna: issues abertas e cartões de issue em Validation depois do merge. (v) Enfraquece o controle do forge sobre o atestado da 106 no cartão (vai a Done sem o forge), mas não o merge, que segue sendo decisão do dono.
- **Visibilidade:** repositório público; issues e títulos ficam visíveis. Revisar o texto antes de publicar.

## Etiquetas (o que o item contém)
Pedido do dono: "devo conseguir identificar o que é código e o que é apenas ADR ou rule ou todos dois juntos". **Aplicado à mão no GitHub em 20/11/2026 (verificado pelo coordenador)**: 5 etiquetas no repositório e o Board "Quadro do ciclo" com os campos Título, Etiquetas e PRs vinculados (`updateProjectV2View` com `configuration.visibleFieldIds`, lido de volta).

| Etiqueta | Cor | Significa |
|---|---|---|
| `código` | #4f8ef7 | Muda código (Go, scripts, testes, templates) |
| `adr` | #9b7bff | Registro de decisão de arquitetura |
| `rule` | #f5b04a | Mexe em regra: Constituição ou `.agents/rules` |
| `documentação` | #99a2b8 | Muda documentação ou texto entregue |
| `spec` | #4fd1b5 | Ainda é só especificação, sem construção |

Várias por item: "só ADR" = `adr`; "código e ADR" = `código` + `adr`. `código`, `adr`, `rule` e `documentação` dizem o que o item **contém ou entregará**; `spec` diz que **ainda é só especificação** e some quando há construção; a coluna `Status` diz o andamento.

**Regra de classificação por caminho** (ignora `sdd/.metrics/`): `sdd/spec/adr/**` = adr; `.agents/rules/**` ou `sdd/memory/constitution.md` = rule; `*.go`, `*.py`, `*.sh`, `*.js`, `.github/workflows/**`, `.gitattributes`, `**/templates/**`, `**/testdata/**` = código; `sdd/discovery/**`, `sdd/spec/**` e `sdd/memory/**` (que não sejam ADR) = spec; `README.md`, `docs/**` e outros `.md` = documentação. PR só de especificação (#120, #121, #122) recebe `spec` (o #121 também `adr`); nos demais a etiqueta `spec` não é usada.

**Tabela aplicada em 20/11/2026** (issue / PR: etiquetas): 134/#123 adr; 135/#127 código+documentação+rule; 136/#126 código+documentação; 137/#125 código; 138/#124 código; 139/#128 código+documentação; 140/#129 código; 145/#131 código+documentação; 146/#130 documentação; 133 (iniciativa 105) código+adr+rule+documentação; 141 spec+código+documentação; 142 spec+código; 143 (106) spec+código+adr+rule; 144 (107) spec+código+adr+rule; 147 (108) spec+código+adr; backlog 148 a 151 sem etiqueta. O PR #132 (bots) recebeu código+documentação e ainda não tem issue.

**Rotina dos agentes:** ao abrir um PR ou uma issue, o forge calcula as etiquetas dos arquivos alterados (função pura caminho para etiquetas, testável) e as aplica no PR e na issue vinculada. Ao mesclar, remove `spec` se houver código. O humano pode ajustar à mão; o forge não remove etiqueta que o humano pôs.

**Limite:** a classificação é por caminho (heurística), sem ler o conteúdo: um `.md` que na prática é regra, ou um arquivo de outro tipo dentro de `docs/`, pode ser classificado errado. **Caminho fora das regras:** recomendo **ficar sem etiqueta e aparecer como aviso** (`forge project status` e `doctor` listam), em vez de cair em `documentação`: cair em `documentação` esconderia código de um tipo novo (por exemplo `.ts`) justamente na etiqueta que o dono usa para separar código. Exceção já na regra: `.md` desconhecido é documentação. Caminho com `..`, caractere de controle ou lista vazia é recusado e não gera etiqueta.

## Tipo no título da issue (prefixo derivado das etiquetas)
Para o tipo aparecer no celular sem abrir a etiqueta, o título de cada **issue** ganha um prefixo entre colchetes **calculado das etiquetas** (aplicado à mão em 20/11/2026): conteúdo = as de `código`, `adr` e `rule` presentes, nessa ordem, unidas por `+`; se não houver nenhuma, `doc` quando houver `documentação`; se `spec` estiver presente, o prefixo começa com `spec · ` (só `spec` quando não há mais nada). Sem etiqueta, sem prefixo. PRs não mudam de título. Exemplos: `[adr] Registrar a decisão ...`; `[código+rule] Trocar as menções ...`; `[doc] O guia inicial ...`; `[spec · código+adr+rule] Gastar menos ...`.
O `indice.csv` e os `issues/*.md` guardam o título **base** (sem prefixo); o prefixo é sempre derivado, por função pura `PrefixoDoTitulo(etiquetas)` em `internal/pipeline`, junto de `ClassificaEtiquetas`. Etiqueta fora da lista fechada é recusada. Para ser idempotente com títulos já prefixados, o `criar-project.sh` procura a issue pelo título base **ou** por qualquer título que termine em `] <base>`, **renomeia** em vez de criar outra e recalcula o prefixo quando as etiquetas mudam.

## Decisões para o Gate 1 (recomendação de cada uma; a decisão é do dono)
1. **Fonte da verdade.** Recomendo o **arquivo** para o estado (spec, fila, PR) e o quadro como espelho; o humano só muda estado pelo menu do forge. Alternativa: o quadro como fonte para a etapa (mais simples de ver, mas qualquer pessoa com escrita no repositório público passa a decidir o fluxo).
2. **O que arrastar o cartão autoriza.** Recomendo que arrastar seja **um pedido**: Spec-validation para Todo abre o menu de confirmação do merge da spec (exige Gate 1 aprovado); Validation para Done só vale com PR mesclado e atestado conforme o modo da 106. Se o `gh` não expuser o ator do evento (a verificar), arrastar nunca aprova sozinho.
3. **Polling versus webhook.** Recomendo **polling** pelo CLI (`gh project item-list`) a cada comando e no `doctor`: não depende de Action (o CI está bloqueado por cobrança) nem de servidor. Webhook e Action: a verificar para Projects v2 de usuário.
4. **Nome do campo.** O campo da etapa é o **`Status`** do Project, com exatamente as 7 opções. Registro histórico: a versão anterior desta discovery recomendava um campo novo `Etapa`, porque supunha que as opções do `Status` não podiam ser editadas; **foi uma premissa errada do pm-architect e do coordenador, corrigida em 20/11/2026**. Verificado na API real em 20/11/2026: a mutation GraphQL `updateProjectV2Field` (com `singleSelectOptions`) substitui as opções do `Status` (cores GRAY, BLUE, GREEN, YELLOW, ORANGE, RED, PINK, PURPLE; `description` obrigatória); `createProjectV2View` com `BOARD_LAYOUT` cria o quadro, que agrupa pelo `Status` por padrão (o agrupamento não é configurável pela API). Cores: Backlog GRAY, Spec BLUE, Spec-validation PURPLE, Todo YELLOW, Building ORANGE, Validation PINK, Done GREEN.
5. **Privado ou público.** Recomendo **privado** no início (`gh project edit <n> --owner @me --visibility PRIVATE`); abrir ao público depois com `--visibility PUBLIC` se o dono quiser mostrar o ciclo. O repositório é público: issue e título ficam visíveis de qualquer jeito.
6. **`blocked`.** Recomendo **não criar a oitava coluna**: o cartão fica na etapa de origem com o campo `Bloqueado`. Alternativa: coluna extra.
7. **Definição de Spec-validation e Todo (corrigida).** Spec-validation = especificação aprovada pelo Gate 1 mas **ainda não mesclada** na principal (aguardando o merge do PR da spec). Todo = a spec já está na principal e pronta para o Builder (o split é do Builder, 106). Validation = trabalho pronto num PR aberto aguardando o dono conferir e mesclar (qualquer tipo). Done = PR mesclado. **Correção (20/11/2026, indicada pelo dono):** a definição anterior ("Todo = Gate 1 aprovado e pronto para o Builder") era uma inconsistência da própria 108 com a 106, que só devolve a spec à principal depois de aprovada e só então deixa o Builder começar. Aplicada à mão no GitHub: #141, #142, #143, #144 e #147 foram para Spec-validation; Todo ficou vazio até um merge de spec.
8. **Escopo do repositório e dos Projects.** Recomendo um Project de usuário "Forge SDD — ciclo" ligado (`gh project link --owner nathanramorim`; com `@me` o link falhou, verificado em 20/11/2026) ao repositório; sem mexer nos projetos 1 a 3 existentes. Alternativa: Project de organização (permite mais controles de ator, a verificar).
9. **Etiquetas em vez de campo (aprovado pelo dono no chat).** Etiquetas permitem várias por item e aparecem no PR e na issue. Recomendação para caminho desconhecido: sem etiqueta e com aviso (ver seção Etiquetas).
10. **D1: Builder começa sozinho (dono).** Registrada com as ressalvas acima. **Pergunta aberta: mecanismo de disparo** (não decidido): (a) processo local do forge (comando de observação): simples e sem custo de CI, mas só funciona com a máquina do dono ligada e a sessão aberta; (b) agente agendado na nuvem (rotina) que consulta o GitHub: independe da máquina, mas é um serviço novo, com custo e credencial (**a verificar**); (c) GitHub Action no merge: nativa, mas o CI está bloqueado por cobrança e a 106 quer reduzir minutos (**a verificar** o custo). Recomendo começar por **(a)**, com o freio de WIP e de tentativas, e reavaliar (b) depois.
11. **D2: via rápida (dono aceitou a recomendação).** O critério é a etiqueta (só `adr`, `rule` e/ou `documentação`, sem `código`), calculada por `ClassificaEtiquetas` sobre os arquivos do PR. Recomendo que o **pm-architect** produza o PR de ADR e de documentação (já escreve só especificação e texto) e que o **Builder** continue necessário para o trabalho em si; só o **cartão** pula a coluna Building. Se o PR passar a conter código, o cartão volta a Building e o dono é avisado.
12. **D3: automações do Project fecham a issue e movem o cartão (dono).** **Perguntas abertas:** (a) ligar as automações na interface: recomendo que o dono ligue "Item closed" e "Pull request merged" (e, se quiser, "Auto-close issue") à mão e que o `forge doctor` confira pelo estado, já que a API não as configura (**a verificar no Project real**); (b) política de palavra de fechamento, sem decidir: usar palavra de fechamento no corpo dos PRs que miram a main (recomendo para PRs diretos); para PRs encadeados, o forge fecha a issue e move o cartão ao perceber o merge, como fallback; ou manter só o cartão do PR. Recomendo a combinação das duas primeiras.
13. **Quando publicar as issues.** Recomendo executar `criar-project.sh --executar` só depois do Gate 1 e de revisar o texto das issues (repositório público); o script é idempotente e começa em `--dry-run`.

## Descrição de negócio (release notes)
- Entrega (em construção, ainda nada pronto): um quadro no GitHub com as sete etapas do ciclo, mantido em sincronia com os arquivos do projeto, para o dono ver de relance onde está cada coisa.

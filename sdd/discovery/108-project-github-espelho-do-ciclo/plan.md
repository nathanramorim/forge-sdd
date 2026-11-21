# Plano — 108 (Project do GitHub como espelho do ciclo)

Ordem de baixo para cima: primeiro o mapa e a leitura (sem escrita), depois mover o cartão, por último o espelho nas duas vias e o `doctor`. Até 7 tasks por onda. Cada task traz teste negativo (`TST-03`) e passa em `forge gate --criterios`. Nada roda sem a reaprovação do escopo (`FLX-01`).

## Sequenciar com
- **105/2.7:** `references/descricao-de-negocio.md` e a seção em toda spec; a 108 só consome o texto. Sem editar o mesmo arquivo na mesma onda.
- **105/2.6:** menu raiz e `forge status`; a coluna só entra em `status` depois.
- **106/2.6 e 0.2:** passos dos comandos e `internal/vcs` (`Adaptador`). A 108/1.2 declara `Quadro` e implementa o adaptador `github` **depois** da 106/0.2 (DURA; decidido no fecho: a 106/0.2 entra primeiro, na onda 1).
- **106/1.6:** sinais `spec-para-revisao` e `pronto-para-revisao`; a 108/1.1 só lê.
- **106/1.6 e 2.6:** a mesma convenção de PR/issue; a 108/2.6 só acrescenta etiquetas ao que a 106 abrir (aplicar depois da 106/2.4, que abre o PR por feature).
- **106/2.4 (PR por feature) e 106/2.1 (`OpenPR`):** a 108/4.1 gera o corpo do PR que a 106 abre; sem editar o mesmo arquivo na mesma onda. **107:** o `observar` respeita o WIP humano e não duplica o laço do orquestrador (a verificar quem hospeda o laço).
- **107 (WIP humano) e 106 (modo, atestado e [Seguir]):** quem consome o limite de WIP e o modo é a 108/3.7 (o `observar`); a 2.7 é só a via rápida e não os usa. A 108 não os redefine.
- **Colisão de ARQUIVO, não dependência** (análise de fecho, 20/11/2026): a 108 NÃO depende de 105/1.4, 105/1.5, 105/1.7, do verificador 107-05 nem do fix-97-01; só mexem nos mesmos arquivos (`doctor.go`, `status.go`, `session.go`, references e goldens). Regra: sequenciar com, um PR por vez em cada arquivo.
- **107:** orquestrador e WIP humano; as references de `sdd-operacao` são compartilhadas. A 108/3.5 só acrescenta `project.md`.
Colisões de arquivo: `.agents/skills/sdd-operacao/references/`, `internal/vcs`, `forge status`. Regra: um PR por vez em cada arquivo; resolver na ordem de merge.

## Onda 0 — Configuração opcional e custo zero (RN-16)
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 0.1 | **Leitura da configuração e curto-circuito de custo zero, antes de qualquer chamada ao `gh`**: lê `Quadro de acompanhamento` e `Início automático do Builder` (padrão `desligado`; valor fora da lista fechada ou incoerente com o VCS é recusado); com o espelho desligado nada chama `gh` ou rede, não há polling, `doctor` só informa "desligado"; aviso de custo ao ligar e modo efetivo gravado no `status.md`. Compartilha o leitor da tabela da Constituição (branch base) com a 106/0.5 | **DURA:** 106/0.5 (leitor da tabela, `DUP-01`); PR único da Constituição (campos) | — | `ARQ-01`, `ARQ-05`, `SEG-01`, `TST-03` |
| 0.2 | ADR **da 108** (fonte da verdade, o que o arrastar autoriza) e linha em `sdd/spec/concepts.md`. **Não edita a Constituição**: os campos, as notas e o ponteiro vão no PR único da Constituição (task nova da 106, reserva a ADR 0022 e cria `.agents/rules/sdd-modos-de-projeto.md`) | **DURA:** PR único da Constituição | com 0.1 (depois do PR único) | `GOV-01`, `DUP-01` |

Sequenciar com a 106: o leitor da tabela de Ferramentas e Integrações é um só (decisão xv da 106); a 108/0.1 não cria um segundo.

## Onda 1 — Mapa e leitura (somente leitura)
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 1.1 | `internal/pipeline`: tipo `Etapa` com a lista fechada das 7; função pura `EtapaDoEstado` que lê o **estado por feature** (`sdd/discovery/<id>/state/<sid>.json`, 106/0.7; nomes de `internal/queue` mantidos: `queued`, `building`, `review_pending`, `approved`, `changes_requested`, `blocked`; o esquema do arquivo não é definido na 106 em 38f12e5), o marcador `Gate 1: aprovado` do `status.md` (106/2.2) e os sinais `spec-para-revisao` e `pronto-para-revisao` (106/1.6); `queue.json` é só o plano, não a fonte do estado; recusa de valor fora da lista; `Bloqueado` como marca | **DURA:** 106/0.7, 106/1.6, 106/2.2 (ver Conferência das ondas) | com 1.4 | `ARQ-01`, `DUP-01`, `TST-03` |
| 1.2 | Interface `Quadro` (ler etapa, mover, vincular PR) no consumidor e adaptador `github` só de leitura (`gh project item-list --format json`, `field-list`) em `internal/vcs`; `nenhum` devolve "sem quadro"; `azure-devops` = a verificar | 1.1; **DURA:** 106/0.2 (`internal/vcs`, interface `Adaptador`; a 108 NÃO cria pacote paralelo) | — | `ARQ-02`, `DUP-03`, `SEG-03`, `SEG-08`, `TST-03` |
| 1.3 | `forge project status` (somente leitura): lista cartões, etapa e divergência com o arquivo; saída saneada, sem token | 1.1, 1.2 | com 1.4 | `ARQ-01`, `ARQ-05`, `SEG-04`, `TST-03` |
| 1.4 | `internal/config`: nome do campo (`Status`), número do Project e `provider` `github-projects`; mapa entre `status_on_implement`/`status_on_pr_approved` e as 7 etapas (sem quebrar o padrão atual) | — | com 1.1 | `DUP-06`, `GOV-01`, `TST-02` |
| 1.5 | Teste de injeção: título, corpo e comentário com ordem embutida são dado; só o campo de etapa válido é lido | 1.2 | com 1.3 | `SEG-08`, `TST-03` |
| 1.7 | Funções puras `ClassificaEtiquetas(caminhos)` e `PrefixoDoTitulo(etiquetas)` (prefixo `[código+adr+rule]`, `doc`, `spec · `; recusa etiqueta desconhecida) em `internal/pipeline` (código, adr, rule, documentação, spec; ignora `sdd/.metrics/`; recusa `..`, controle e lista vazia; desconhecido fica sem etiqueta com aviso). Entra na onda 1 porque é pura, sem rede e sem escrita, como o mapa 1.1 | — | com 1.1 | `ARQ-10`, `DUP-01`, `TST-03` |

## Onda 2 — Mover o cartão pelo CLI
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 2.1 | Adaptador `github`: `Mover` (`gh project item-edit` com IDs de `field-list` e `item-list`) e `Vincular` (`item-add --url`); idempotente | 1.2 | com 2.2 | `ARQ-02`, `SEG-03`, `TST-03` |
| 2.2 | `forge project move <iniciativa|feature> --para <etapa> [--dry-run]`; `--dry-run` sem efeito; etapa validada na lista fechada | 1.1 | com 2.1 | `ARQ-01`, `ARQ-05`, `TST-03` |
| 2.3 | Gatilhos: depois de `queue next`, `queue complete`, `queue review` e do Gate 1 o forge lê o estado **por feature** (106/0.7) e o marcador `Gate 1: aprovado` (106/2.2) e chama o movimento (falha do quadro não bloqueia a fila; vira aviso) | 2.1, 2.2; **DURA:** 106/0.7, 106/2.2 | — | `FLX-06`, `COD-03`, `TST-03` |
| 2.4 | Sem VCS (`nenhum`) e sem Project configurado: nada é chamado; teste sem `gh` no PATH | 2.1 | com 2.5 | `ARQ-05`, `TST-03` |
| 2.7 | Via rápida: função pura `ViaRapida(arquivos)` sobre `ClassificaEtiquetas` (sem `código` e com ao menos uma de adr, rule, documentação); Todo vai direto a Validation; volta a Building se aparecer `código` e o dono é avisado | 1.7, 2.3 | com 2.6 | `ARQ-10`, `DUP-01`, `TST-03` |
| 2.6 | Aplicar etiquetas e o prefixo do título da issue (renomeia a existente em vez de duplicar): ao abrir PR ou issue, calcular pela 1.7 e aplicar no PR e na issue vinculada (`gh pr edit` e `gh issue edit --add-label`, pelo adaptador); ao mesclar, remover `spec` se há código; nunca remover a etiqueta posta à mão | 1.7, 2.1; **DURA:** 106/0.2 | com 2.5 | `SEG-03`, `SEG-08`, `TST-03` |
| 2.5 | `criar-project.sh` revisado e testado (rascunho desta iniciativa) com o Project real do dono; relatório sem token. **Tarefa do dono, não de um Builder isolado**: exige o Project real e o `gh` autenticado dele; a Constituição proíbe rede nos testes (`TST-04`), então o Builder só entrega o script e o teste com `gh` falso | 2.1 | com 2.4 | `SEG-04`, `TST-02` |

## Onda 3 — Espelho em duas vias e `doctor`
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 3.1 | (RN-18: ordena os cartões por prioridade, posição relativa da fila; reordenar via `updateProjectV2ItemPosition` é a verificar; desempate PROPOSTA a confirmar) Polling: o forge lê a etapa de cada cartão a cada comando e compara com o arquivo; divergência vira pedido, nunca gravação direta; ator conferido (a verificar se o `gh` o expõe) | 1.3, 2.3 | — | `SEG-01`, `TST-03` |
| 3.2 | (RN-18: arrastar na mesma coluna é pedido de reordenar a fila, com menu) Arrastar como pedido (RN-17: ao ver Validation, o cartão MOSTRA o comentário do Reviewer e o resultado; quem o escreve é o core): Spec-validation para Todo abre o menu de confirmação e exige Gate 1 aprovado e PR da spec mesclado; Validation para Done exige PR mesclado e atestado conforme o modo da 106 (só `local` entregue; `vcs` adiado) | 3.1; **DURA:** 106/3.2 (atestado, só `local`), que puxa 106/0.5, 0.6 e 0.3; a 0.3 é Azure e fica fora: a cadeia entregue é só `github` e `nenhum` | — | `FLX-01`, `GOV-01`, `TST-03` |
| 3.7 | Início automático (D1; só se `Início automático do Builder` = `local`, senão recusa com código diferente de zero) por **`forge project observar`** (mecanismo decidido pelo dono: processo local enquanto a sessão estiver aberta): polling com intervalo mínimo de 30 s (a verificar contra o limite de taxa), só lê estado e merge verificado, aplica a regra de Todo para Building (spec na principal, WIP humano da 107, teto de 3 tentativas, modo/atestado da 106), não grava aprovação sem menu, encerra limpo e tem `--dry-run`; lê o estado por feature (106/0.7) e o marcador do `status.md`, não o `queue.json` como fonte única; nuvem e Action ficam como alternativas futuras | 2.3, 2.7; **DURA:** 107-12 (WIP humano), 106/2.4 (um PR por feature), 106/0.5, 106/1.6 | — | `FLX-01`, `SEG-08`, `TST-03` |
| 3.3 | `forge doctor` lista issue aberta com PR mesclado e as automações do Project não conferidas (D3, a verificar no Project real); verifica Project configurado, escopo `project` do `gh`, campo `Status` com as 7 opções e divergências arquivo versus quadro | 3.1 | com 3.4 | `ARQ-01`, `TST-03` |
| 3.4 | Segunda fonte de pendências do orquestrador (107): implementa a interface `FonteDePendencias` (`Pendentes(ctx) (int, error)`, **proposta do coordenador, a confirmar pelo dono**; 107-12 em 203d98b; padrão da 107 = arquivo) contando cartões em Validation, **só quando o espelho estiver ligado**; lê só o campo validado; a 107 não depende da 108 | 3.1; **DURA:** 107-12 (WIP humano) | com 3.3 | `DUP-01`, `SEG-08` |
| 3.5 | Reference única `.agents/skills/sdd-operacao/references/project.md` (rotina por agente e por coluna) e os comandos `/discovery`, `/nova-feature`, `/novo-fix`, `/implementar`, `/revisar` só a apontam em **uma linha** (templates e goldens junto); P-24 conferido antes e depois | 3.2 | — | `ARQ-03`, `TST-02`, `GOV-01` |
| 3.6 | Prova ponta a ponta com um Project privado de teste: arquivo para quadro, quadro para menu, ator desconhecido recusado; relatório com o que ficou "a verificar". **Tarefa do dono** (Project real e `gh` autenticado dele; `TST-04` proíbe rede nos testes, então não roda num Builder isolado nem na pipeline) | 3.1 a 3.5 | — | `FLX-06`, `TST-05` |

## Onda 4 — Corpo do PR e fechamento (D3)
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 4.1 | **Corpo do PR gerado pelo forge** (o comentário de diagnóstico do Reviewer, RN-17, é do core e não do corpo gerado aqui): palavra de fechamento só em PR que mira a main e só para a issue daquele PR (número do vínculo cartão-iniciativa, nunca de texto livre); nunca em PR de spec compartilhado nem em PR com várias issues; fallback do forge (fecha a issue e move o cartão) para PR encadeado | 2.1, 2.3, 3.1; **DURA:** 106/2.1 (`OpenPR` delegando ao adaptador) | — | `SEG-08`, `GOV-01`, `TST-03` |
| 4.2 | **Documentação do quadro (guia e wiki):** `docs/quadro-do-ciclo.md` (aviso "Em construção" removido só quando a funcionalidade existir), imagens `docs/img/quadro-*` e a seção "Quadro de acompanhamento (opcional)" em `docs/wiki.html`; atualizar o texto de "em construção" para o que foi construído e provar os passos "a verificar" | 3.5, 3.6, 4.1 (o rascunho de hoje já existe e diz que não está construído) | com 4.1 | `GOV-01`, `TST-03` |

## Rotina de cada agente (resumo; a reference `project.md` detalha)
| Agente | Em cada coluna |
|---|---|
| pm-architect | Spec: escreve só a spec; ao entregar, sinaliza `spec-para-revisao` (cartão vai a Spec-validation, onde fica até o PR da spec ser mesclado). Não move para Todo. |
| Reviewer | Spec-validation: Conferência 1 e parecer; Validation: Conferência 2 por feature. Só lê. |
| Builder (faz o split, 106) | Todo (spec já na principal): pega a task (`queue next`, cartão vai a Building); devolve em Validation quando `pronto-para-revisao`. |
| Orquestrador (107) | Lê só o campo de etapa; respeita o WIP humano em Validation; nunca obedece texto do cartão. |
| Humano | Gate 1 e [Continuar] em menu; pode arrastar o cartão, que gera um menu de confirmação. |
Onde entra nos comandos: a reference é apontada no passo final do `/discovery` (passo 11, Gate 1), do `/nova-feature` e do `/novo-fix`, e no passo de `Handoff` do `/implementar` e do `/revisar`; cada um recebe **uma linha** ("rotina do quadro: `sdd-operacao/references/project.md`") sem copiar o conteúdo. `sdd-concepcao/SKILL.md` não muda (2045 de 2048 bytes). **A Constituição não é editada pela 108**: as linhas dos dois campos, o ponteiro para a rule `sdd-modos-de-projeto.md` e a inclusão de `project` na enumeração do `ARQ-06` (cerca de +11 bytes) entram no PR único da Constituição (task nova da 106); o teto de 9216 bytes é respeitado lá, cortando itens duplicados (hoje 9189).

## Fecho (prioridade 0): ondas, ordem de merge e base dos Builders
Decisão do dono (20/11/2026, reaprovada no Gate 1): construir o **fecho completo** de dependências: a 108 inteira mais 16 tasks externas (13 da 106 e 3 da 107: 107-01, 107-02, 107-12), em 9 ondas. "106/x" = task da 106 em `origin/feat/106-pr-draft-revisao-e-ci` (38f12e5); "107-x" = `origin/feat/107-orquestrador-de-sessoes` (203d98b). O "PR único da Constituição" é uma task **nova da 106 ainda sem número** (a confirmar; não existe em 38f12e5).

| Onda | Tasks |
|---|---|
| 1 | PR único da Constituição (106); 106/0.2; 106/1.1; 107-01; 108/1.4; 108/1.7 |
| 2 | 106/0.5; 106/0.7; 106/1.3; 107-02; 108/1.1 (ver conferência) |
| 3 | 106/1.6; 106/0.6 (só `github` e `nenhum`); 107-12; 108/0.1; 108/1.2; 108/2.2 |
| 4 | 106/2.1; 106/3.2 (só `local`); 108/1.3; 108/1.5; 108/2.1 |
| 5 | 106/2.2; 108/2.4; 108/2.5 (tarefa do dono); 108/2.6 |
| 6 | 106/2.3; 106/2.4; 108/2.3 |
| 7 | 108/2.7; 108/3.1 |
| 8 | 108/3.2; 108/3.3; 108/3.4; 108/3.7; 108/4.1 |
| 9 | 108/3.5; 108/3.6 (tarefa do dono); 108/4.2 |

Decisões do dono que valem para o fecho: (1) fecho completo; (2) **um único PR da Constituição** (reserva a ADR 0022 e cria a rule `.agents/rules/sdd-modos-de-projeto.md` para as NOTAS; a Constituição só recebe as linhas dos campos e um ponteiro, com o teto de 9216 bytes respeitado cortando itens duplicados); (3) o fallback do atestado só para `github` e `nenhum` (Azure DevOps fora; recusa com "não verificado"); (4) o modo `vcs` do atestado adiado (vale o `local`).

**Conferência das ondas** (cada ordem contra as dependências; **a confirmar pelo coordenador**):
- **108/1.1 na onda 2** depende de 106/0.7 (onda 2, a mesma) e de 106/1.6 (onda 3) e 106/2.2 (onda 5). Proposta: 1.1 consome só os nomes literais já fixados na spec da 106 (função pura sobre dados já lidos), então 1.6 e 2.2 contam como MOLE para o código da 1.1 e valem como DURA só no teste de integração da 2.3/3.7; 0.7 continua DURA e, estando na mesma onda, exige 0.7 mesclada antes: ou 1.1 desce para a onda 3 (e 1.2, 2.2 descem junto) ou o esquema do arquivo de estado é fixado na spec da 106.
- **108/0.2** não estava na tabela do coordenador. Proposta: onda 2 (depende só do PR único, onda 1).
- **Onda 9:** 3.6 depende de 3.5, e 4.2 depende de 3.5, 3.6 e 4.1. Proposta: 3.5, depois 3.6, depois 4.2 em sequência (a onda tem 3 tasks, mas sem paralelismo).
- Demais ordens conferidas e sem violação: 106/2.1 (0.2, 1.1, 1.6), 106/2.2 (0.7, 1.3, 2.1), 106/2.4 (0.7, 2.1, 2.2), 106/2.3 (1.3, 2.2), 106/3.2 (0.5, 0.6, 1.1), 107-12 (107-02, 106/0.5, PR único), 108/2.3, 2.6, 3.7 e 4.1.

**Logística.** Pré-requisito do dono: mesclar na main os PRs de spec #120, #121, #122 e #153; depois a cadeia da 105. Os Builders partem da `main` **depois de cada onda mesclada**. Branch **por task** enquanto a Regra 15 (uma branch por iniciativa) não for reescrita pelo PR único da Constituição, com bloco `constitution-exception` do dono (`GOV-02`).

**Caminho crítico:** 106/0.2, 106/0.7, 106/1.6, 106/2.1, 106/2.2, 106/2.4, 108/3.7, 108/3.5, 108/3.6.

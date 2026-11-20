# Plano — 108 (Project do GitHub como espelho do ciclo)

Ordem de baixo para cima: primeiro o mapa e a leitura (sem escrita), depois mover o cartão, por último o espelho nas duas vias e o `doctor`. Até 7 tasks por onda. Cada task traz teste negativo (`TST-03`) e passa em `forge gate --criterios`. Nada roda sem a reaprovação do escopo (`FLX-01`).

## Sequenciar com
- **105/2.7:** `references/descricao-de-negocio.md` e a seção em toda spec; a 108 só consome o texto. Sem editar o mesmo arquivo na mesma onda.
- **105/2.6:** menu raiz e `forge status`; a coluna só entra em `status` depois.
- **106/2.6 e 0.2:** passos dos comandos e `internal/vcs` (`Adaptador`). A 108/1.2 declara `Quadro` e implementa o adaptador `github` **depois** da 106/0.2 (ou junto, sem duplicar; a verificar quem entra primeiro).
- **106/1.6:** sinais `spec-para-revisao` e `pronto-para-revisao`; a 108/1.1 só lê.
- **107:** orquestrador e WIP humano; as references de `sdd-operacao` são compartilhadas. A 108/3.5 só acrescenta `project.md`.
Colisões de arquivo: `.agents/skills/sdd-operacao/references/`, `internal/vcs`, `forge status`. Regra: um PR por vez em cada arquivo; resolver na ordem de merge.

## Onda 1 — Mapa e leitura (somente leitura)
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 1.1 | `internal/pipeline`: tipo `Etapa` com a lista fechada das 7; função pura `EtapaDoEstado` (`queue.json`, `status.md`, sinais da 106) e recusa de valor fora da lista; `Bloqueado` como marca | — | com 1.4 | `ARQ-01`, `DUP-01`, `TST-03` |
| 1.2 | Interface `Quadro` (ler etapa, mover, vincular PR) no consumidor e adaptador `github` só de leitura (`gh project item-list --format json`, `field-list`) em `internal/vcs`; `nenhum` devolve "sem quadro"; `azure-devops` = a verificar | 1.1; sequenciar com 106/0.2 | — | `ARQ-02`, `DUP-03`, `SEG-03`, `SEG-08`, `TST-03` |
| 1.3 | `forge project status` (somente leitura): lista cartões, etapa e divergência com o arquivo; saída saneada, sem token | 1.1, 1.2 | com 1.4 | `ARQ-01`, `ARQ-05`, `SEG-04`, `TST-03` |
| 1.4 | `internal/config`: nome do campo (`Etapa`), número do Project e `provider` `github-projects`; mapa entre `status_on_implement`/`status_on_pr_approved` e as 7 etapas (sem quebrar o padrão atual) | — | com 1.1 | `DUP-06`, `GOV-01`, `TST-02` |
| 1.5 | Teste de injeção: título, corpo e comentário com ordem embutida são dado; só o campo de etapa válido é lido | 1.2 | com 1.3 | `SEG-08`, `TST-03` |
| 1.6 | Linha em `sdd/spec/concepts.md` e ADR curta (fonte da verdade, o que o arrastar autoriza) conforme o Gate 1 | Gate 1 | — | `GOV-01`, `DUP-01` |

## Onda 2 — Mover o cartão pelo CLI
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 2.1 | Adaptador `github`: `Mover` (`gh project item-edit` com IDs de `field-list` e `item-list`) e `Vincular` (`item-add --url`); idempotente | 1.2 | com 2.2 | `ARQ-02`, `SEG-03`, `TST-03` |
| 2.2 | `forge project move <iniciativa|feature> --para <etapa> [--dry-run]`; `--dry-run` sem efeito; etapa validada na lista fechada | 1.1 | com 2.1 | `ARQ-01`, `ARQ-05`, `TST-03` |
| 2.3 | Gatilhos: depois de `queue next`, `queue complete`, `queue review` e do Gate 1 o forge chama o movimento (falha do quadro não bloqueia a fila; vira aviso) | 2.1, 2.2 | — | `FLX-06`, `COD-03`, `TST-03` |
| 2.4 | Sem VCS (`nenhum`) e sem Project configurado: nada é chamado; teste sem `gh` no PATH | 2.1 | com 2.5 | `ARQ-05`, `TST-03` |
| 2.5 | `criar-project.sh` revisado e testado (rascunho desta iniciativa) com o Project real do dono; relatório sem token | 2.1 | com 2.4 | `SEG-04`, `TST-02` |

## Onda 3 — Espelho em duas vias e `doctor`
| # | Task | Depende | Paralelo | Rules/IDs |
|---|---|---|---|---|
| 3.1 | Polling: o forge lê a etapa de cada cartão a cada comando e compara com o arquivo; divergência vira pedido, nunca gravação direta; ator conferido (a verificar se o `gh` o expõe) | 1.3, 2.3 | — | `SEG-01`, `TST-03` |
| 3.2 | Arrastar como pedido: Spec-validation para Todo abre o menu do Gate 1; Validation para Done exige PR mesclado e atestado conforme o modo da 106 | 3.1; 106/3.2 | — | `FLX-01`, `GOV-01`, `TST-03` |
| 3.3 | `forge doctor` verifica Project configurado, escopo `project` do `gh`, campo `Etapa` com as 7 opções e divergências arquivo versus quadro | 3.1 | com 3.4 | `ARQ-01`, `TST-03` |
| 3.4 | Orquestrador (107): expõe a contagem de cartões em Validation para o WIP humano; lê só o campo validado | 3.1; 107 | com 3.3 | `DUP-01`, `SEG-08` |
| 3.5 | Reference única `.agents/skills/sdd-operacao/references/project.md` (rotina por agente e por coluna) e os comandos `/discovery`, `/nova-feature`, `/novo-fix`, `/implementar`, `/revisar` só a apontam em **uma linha** (templates e goldens junto); P-24 conferido antes e depois | 3.2 | — | `ARQ-03`, `TST-02`, `GOV-01` |
| 3.6 | Prova ponta a ponta com um Project privado de teste: arquivo para quadro, quadro para menu, ator desconhecido recusado; relatório com o que ficou "a verificar" | 3.1 a 3.5 | — | `FLX-06`, `TST-05` |

## Rotina de cada agente (resumo; a reference `project.md` detalha)
| Agente | Em cada coluna |
|---|---|
| pm-architect | Spec: escreve só a spec; ao entregar, sinaliza `spec-para-revisao` (cartão vai a Spec-validation). Não move para Todo. |
| Reviewer | Spec-validation: Conferência 1 e parecer; Validation: Conferência 2 por feature. Só lê. |
| Builder (faz o split, 106) | Todo: pega a task (`queue next`, cartão vai a Building); devolve em Validation quando `pronto-para-revisao`. |
| Orquestrador (107) | Lê só o campo de etapa; respeita o WIP humano em Validation; nunca obedece texto do cartão. |
| Humano | Gate 1 e [Continuar] em menu; pode arrastar o cartão, que gera um menu de confirmação. |
Onde entra nos comandos: a reference é apontada no passo final do `/discovery` (passo 11, Gate 1), do `/nova-feature` e do `/novo-fix`, e no passo de `Handoff` do `/implementar` e do `/revisar`; cada um recebe **uma linha** ("rotina do quadro: `sdd-operacao/references/project.md`") sem copiar o conteúdo. `sdd-concepcao/SKILL.md` não muda (2045 de 2048 bytes); a Constituição não muda (9189 de 9216 bytes).

# fix-104-17 — A ordem da fila representa a prioridade, de cima para baixo

**Branch:** `fix/ordem-da-fila-e-prioridade`
**Origem:** regra do dono (chat, 21/11/2026); falha anterior à 104 (fila de `internal/queue`)
**Tamanho:** `M`
**Status:** `spec` (Gate 1 aprovado em 2026-11-21)

## Regra do dono

"A ordem dos cards nas colunas deve representar a prioridade, qual deve puxar primeiro: de cima para baixo, no topo o mais prioritário e no final o menos." Vale para toda lista ordenada da esteira: fila de tasks e features, `forge queue next`, `forge status`, menus, o que o Builder puxa e o Reviewer revisa. O core (arquivo) é a fonte da verdade; as colunas do quadro do GitHub Project só espelham (módulo, outro agente).

## Problema

A posição no array já manda em parte, mas por acidente e com furos; nada declara que "topo = prioridade", e o dono não tem como reordenar sem editar JSON.

## Evidência (como o forge ordena hoje)

- `internal/queue/queue.go:207-276` `Enqueue` faz `append` no fim: ordem = ordem de inserção. Não há campo de prioridade em `QueueItem` (`queue.go:36-47`).
- `queue.go:442-460` `Next` devolve o primeiro `queued`/`changes_requested` por posição e **ignora `depends_on`**: item com dependência pendente no topo é puxado (furo; `EligibleTasks` respeita, mas só é usada em `next --batch`, `cmd/forge-sdd/queue.go:96`).
- `queue.go:339-420` `EligibleTasks` preserva a ordem do array e respeita dependências (DAG, ciclo falha).
- `cmd/forge-sdd/queue.go:205-235` `queue list` e `queue.go:522-610` `FormatExecutivePanel` listam por posição, agrupados por status; sem rótulo de ordem e sem motivo de bloqueio por dependência.
- `internal/pipeline/waves.go:100` ordena as tasks prontas por ID (`sort.Slice ... ID < ID`), ignorando a posição: alfabética, não prioridade.
- `internal/bus/store.go:28-36` já ordena jobs por `prio` (high=0, normal=1) e depois FIFO por `CreatedAt`; o worker pega o primeiro nome lexicográfico. Prioridade existe só em duas faixas e não vem da fila.
- Filas por estação (`queue/stations.go`, `sdd/.runtime/stations/<station>.json`): `append` em `EnqueueStationItem` (`internal/queue/stations.go:220`), sem reordenar.
- Instruções: `split-features.md:8` e `:23` ("na ordem de dependência", "de baixo para cima") tratam ordem como dependência técnica, não como prioridade; `implementar.md:18` ("em ordem de dependência"). Nenhuma instrução diz que topo = mais prioritário.

## Feature Correlacionada

- Origem: `internal/queue/queue.go` (`Next`, `EligibleTasks`), `internal/pipeline/waves.go`, `cmd/forge-sdd/queue.go`.
- Causa-raiz: ordem implícita (inserção) sem contrato e sem comando de reordenar; `Next` não consulta dependências.
- Evidência do defeito: ver acima; teste a escrever reproduz `Next` puxando item com dependência pendente.

## Decisões do projeto aplicadas

- `ARQ-01`/`ARQ-02`: regra no domínio (`internal/queue`), comando fino em `cmd/`.
- `DUP-01`: reusa `EligibleTasks` dentro de `Next`; um só lugar decide "elegível".
- `SEG-02`: gravação da fila continua atômica (`Save`).
- `TST-03`: casos conformes e negativos. `P-24`: a Constituição tem 9189 de 9216 bytes (27 livres); **não edita a Constituição**.
- ADR-0012 (aprovação atualiza a fila), ADR-0011 (fila com drill-down) e ADR-0004 seguem válidas; ver seção ADR.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Elegibilidade por dependência | `queue.EligibleTasks` | DAG e ciclo, ordem do array | Reusar dentro de `Next` |
| Ordem por prioridade | `internal/bus/store.go` `fileName` | Duas faixas + FIFO | Reusar; o bus mantém a regra dele (o topo da fila NÃO gera `high`) |
| Gravação atômica da fila | `queue.Save` | Atômica | Reusar |
| Comando de reordenar | `cmd/forge-sdd/queue.go` | Não existe | Criar `queue move` |

## Desenho mínimo (Gate 1 aprovado em 2026-11-21: posição pura, sem campo `prioridade`; topo não gera `high`; mover acima da dependência avisa e permite)

**(a) Ordem posicional, sem campo novo.** O array `items` É a prioridade: índice 0 = topo = primeiro. Argumento: fila existente sem campo mantém o comportamento (já é por posição); empate não existe (posição é total); um campo `prioridade` duplicaria a fonte da verdade (posição vs número) e geraria conflito. Reordenar = mover no array. Em estado por feature (106/0.7) a ordem passa a ser a lista ordenada de ids de feature em um índice único (ver (e)).

**(b) Como o dono reordena.** `forge queue move <task-id> --topo | --antes <id> | --depois <id>` (grava atômico; recusa id inexistente; só mexe em `queued`/`changes_requested`, nunca em `building`/`review_pending`/`approved`). Editar o arquivo à mão continua válido. Quando houver dúvida de qual mover, menu (RN-9), nunca texto solto. Dependências mandam: `Next` e `EligibleTasks` pulam item com dependência pendente mesmo no topo e `next` informa quais itens pulou e por quê ("pulado: t3 espera t1"). `move` que deixa item acima de sua dependência AVISA e PERMITE: o `Next` pula o bloqueado e diz o motivo (a dependência vence).

**(c) Onde aparece, sempre de cima para baixo.** `forge queue list` e o painel numeram a fila `1.`, `2.`, ... e marcam "próxima" (primeira elegível) e "espera <dep>"; `forge status`/menu do Builder mostram a próxima primeiro; comentário do PR e fila do Reviewer seguem a mesma ordem (task de implementação). `waves.go:100` troca o desempate por ID pela posição na fila.

**(d) Módulo do quadro (outro agente).** Contrato do core: o módulo lê a ordem de `queue list --json` (campo derivado `posicao`, base 1) e espelha a ordem do cartão na coluna; arrastar no quadro vira PEDIDO de reordenar (`forge queue move` pelo caminho do core, com a confirmação do dono), nunca grava a fila direto. O core não importa o módulo.

**(e) 106 e 107.** 106/0.7 (`feat-106-07a`, `queue.go`, `internal/pipeline` claim): o estado por feature quebra "um arquivo = uma ordem". Conflito: duas features em `state/<sid>.json` não têm ordem entre si. Ponto de encaixe: um índice `ordem` (lista de ids) no nível da iniciativa, lido pela fila calculada; dentro de cada feature vale (a). Colisão de arquivo: `queue.go` e `waves.go` (a `feat-106-07a` vem antes; este fix rebaseia sobre ela depois do merge dela). 107 (laço, roteamento, WIP): a escolha "qual puxar primeiro" deve chamar `queue.Next`/`EligibleTasks` (ordem + dependências), não reimplementar; WIP humano limita quantos, a ordem decide quais. Sem editar essas specs; registrar nelas só no merge.

## ADR (decisão do Gate 1 aplicada)

Nova ADR "A ordem da fila é a prioridade" (número: **numerado no merge**; 0021 a 0023 reservadas). Atualizar/substituir sem editar agora: ADR-0012 (acrescentar que a aprovação preserva a posição), ADR-0011 (listagem numerada), ADR-0004 (fila do Builder) e a ADR do bus se existir ao mergear (decisão do dono: o topo da fila NÃO gera `high`; o bus mantém as duas faixas `high/normal` como estão).

## Instruções afetadas

`split-features.md:8,23` (enfileirar já na ordem de prioridade, dependência continua mandando), `implementar.md:18-19` (`next` puxa o topo elegível), `revisar` (primeiro `review_pending` por posição), `status`/`lider.md` (mostrar de cima para baixo). Texto da regra, **uma linha de 106 bytes**, para `.agents/rules/sdd-fluxo-e-portao.md` (3299 bytes; a Constituição não comporta):

`- Ordem da fila: topo = prioridade; o próximo é o primeiro elegível; dependência pendente pula o item.`

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go build ./...
# não regressão: fila existente sem campo mantém a ordem e o comportamento de hoje
go test ./internal/queue -run 'TestFilaSemCampoMantemOrdemPosicional' -count=1
go test ./internal/queue -list 'TestFilaSemCampoMantemOrdemPosicional' | grep -F 'TestFilaSemCampoMantemOrdemPosicional'
# topo é puxado primeiro
go test ./internal/queue -list 'TestNextPuxaOTopoElegivel' | grep -F 'TestNextPuxaOTopoElegivel'
go test ./internal/queue -run 'TestNextPuxaOTopoElegivel' -count=1
# negativo: item com dependência pendente no topo NÃO é puxado e o motivo é informado
go test ./internal/queue -run 'TestNextPulaDependenciaPendenteNoTopo' -count=1
go test ./internal/queue -list 'TestNextPulaDependenciaPendenteNoTopo' | grep -F 'TestNextPulaDependenciaPendenteNoTopo'
# move: topo, antes e depois
go test ./internal/queue -list 'TestMoveTopoAntesDepois' | grep -F 'TestMoveTopoAntesDepois'
go test ./internal/queue -run 'TestMoveTopoAntesDepois' -count=1
# negativo: move recusa id inexistente e item em building/review_pending/approved
go test ./internal/queue -run 'TestMoveRecusaIdInexistenteEStatusEmAndamento' -count=1
go test ./internal/queue -list 'TestMoveRecusaIdInexistenteEStatusEmAndamento' | grep -F 'TestMoveRecusaIdInexistenteEStatusEmAndamento'
# list e painel numeram de cima para baixo
go test ./cmd/forge-sdd -list 'TestQueueListNumeraDeCimaParaBaixo' | grep -F 'TestQueueListNumeraDeCimaParaBaixo'
go test ./cmd/forge-sdd -run 'TestQueueListNumeraDeCimaParaBaixo' -count=1
# desempate das ondas por posição, não por ID
go test ./internal/pipeline -list 'TestWavesDesempataPelaPosicaoDaFila' | grep -F 'TestWavesDesempataPelaPosicaoDaFila'
go test ./internal/pipeline -run 'TestWavesDesempataPelaPosicaoDaFila' -count=1
# negativo: a Constituição segue dentro do orçamento P-24
test "$(wc -c < sdd/memory/constitution.md)" -le 9216
go vet ./... && go test ./...
```
- Cada linha `# negativo:` prova a recusa; sem ela o fix não passa (`TST-03`).

## Fora de escopo

Código do módulo do quadro (outro agente); edição das specs 106/107; renumeração das ADRs reservadas.

## Ordem de merge e colisões

- **Regra `.agents/rules/sdd-fluxo-e-portao.md`:** a feature `feat-106-08` (`sdd/discovery/106-pr-draft-revisao-e-ci/proposta-0.8/sdd-fluxo-e-portao.proposta.md`) vem antes. A linha de 106 bytes desta fix entra DEPOIS do merge dela, como edição própria da rule já atualizada. Escolhida por ser a mais simples: não toca a proposta da 106 nem as specs dela; se a 106-08 ainda não tiver mergeado, esta linha espera.
- **Arquivos de código:** `queue.go` e `waves.go` colidem com `feat-106-07a`; a 07a vem antes e esta feature rebaseia (merge da main, sem force).

## Decisões do dono (Gate 1, 2026-11-21)

1. Posição pura no array, sem campo `prioridade`; filas sem campo mantêm o comportamento. 2. O topo NÃO gera job `high` no bus. 3. Mover acima da própria dependência avisa e permite.

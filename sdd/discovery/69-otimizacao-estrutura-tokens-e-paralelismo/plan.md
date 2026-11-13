# Plano de Discovery 69 — Estrutura, Tokens e Paralelismo

Branch única `feat/69-otimizacao-estrutura-tokens-e-paralelismo` (`FLX-02`), em worktree próprio criado a partir do tip da `feat/68` (e9a00b6); o PR mira `feat/68` até ela ser mesclada. Pasta de execução: `sdd/discovery/69-otimizacao-estrutura-tokens-e-paralelismo/features/`. Status: `FLX-01` aprovado; fatiada em 7 tasks e enfileiradas.

## Quebra preliminar (Rule of 7: 7 features)
| Feature | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 69-01 Remoção do Studio | M | — | Apaga `internal/studio`, `studio.go`, manual; limpa `FLOW.md` §8, README, templates, golden, `concepts.md`, `sdd-acp.md`; ADR 0001; revisa spec da feat 68 | `sdd-arquitetura`, `sdd-codigo-e-testes`; `COD-04`, `TST-02`, `FLX-01`, `FLX-05` | — |
| 69-02 `internal/layout` | M | — | Dono dos caminhos da iniciativa; resolução por branch; troca dos literais `sdd/...` em `cmd/` e `internal/` | `sdd-arquitetura`, `sdd-seguranca`; `ARQ-01`, `ARQ-02`, `DUP-04`, `SEG-01`, `TST-03` | — |
| 69-03 Fila e status por iniciativa | M | 69-02 | `queue.json` e `status.md` na pasta; atividade e estações em `sdd/.runtime/` ignorado; `P-21` por iniciativa | `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`; `COD-01`, `TST-03`, `TST-04` | — |
| 69-04 `status` e `next-id` | M | 69-02 | `forge-sdd status [--all-branches]` e `forge-sdd next-id` lendo refs via `git` com argumentos explícitos | `sdd-arquitetura`, `sdd-seguranca`; `ARQ-06`, `ARQ-10`, `SEG-03`, `TST-04` | `novo-subcomando-cli-sdd` |
| 69-05 Migração do legado | G | 69-03, 69-04 | `update` move legado para `sdd/archive/` com `--dry-run`, backup e idempotência; templates e golden no novo layout; ADR 0002; teste de merge sem conflito | `sdd-seguranca`, `sdd-fluxo-e-portao`; `ARQ-03`, `ARQ-05`, `SEG-02`, `TST-02`, `GOV-01` | — |
| 69-06 Skills por estação | G | 69-02 | 5 skills com `references/`; remove `.agents/commands/` e `.agents/roles/`; adaptadores finos; instruções raiz enxutas; `FLOW.md` reduzido | `sdd-reuso-e-duplicacao`; `DUP-05`, `DUP-06`, `DUP-10`, `GOV-04`, `TST-02` | — |
| 69-07 Orçamento de contexto | M | 69-06 | Constitution com regra única por ID; `forge-sdd context <task>`; checks de orçamento no gate; `TestContextBudget`; ADR 0003 | `sdd-fluxo-e-portao`, `sdd-codigo-e-testes`; `GOV-01`, `FLX-06`, `TST-01`, `TST-05` | `novo-subcomando-cli-sdd` |

## Ordem e paralelismo
`69-01` e `69-02` em paralelo → `69-03`, `69-04` e `69-06` em paralelo → `69-05` e `69-07`.

Ordem de fila sugerida: 69-01, 69-02, 69-03, 69-04, 69-06, 69-05, 69-07.

## Agrupamento das skills (entrada da 69-06)
| Skill | Absorve comandos | Absorve papéis |
|---|---|---|
| `sdd-concepcao` | `discovery`, `nova-feature`, `novo-fix`, `c4-architecture`, `constitution` | `pm-architect` |
| `sdd-planejamento` | `split-features` | `planner` |
| `sdd-build` | `proxima-feature` | `builder` |
| `sdd-governanca` | `revisar` | `governance` |
| `sdd-operacao` | `status`, `doctor`, `archive`, `upgrade-sdd`, `tutorial` | `lider`, `archivist`, `migrator` |

## Decisões FLX-01 incorporadas (2026-11-13)
| # | Decisão | Resultado |
|---|---|---|
| 1 | Legado: mover discoveries soltos e `sdd/features/` para `sdd/archive/` sem reescrever | Aceita (69-05) |
| 2 | Features e fixes sem discovery usam pasta de iniciativa enxuta | Aceita (69-02, 69-03) |
| 3 | Base da branch | **Alterada:** `release/v3.0.1` não tem `internal/acp` nem as pastas 66 a 68. Base é o tip da `feat/68` em worktree separado, para não levar os 3 arquivos ACP não commitados da 68 |
| 4 | Versão `4.0.0-beta` | Aceita; bump fica para o PR de release, fora das 7 tasks |
| 5 | Nome da pasta `<id>-<nome>` | Aceita |

Efeitos sobre a 68 (commit e9a00b6 na `feat/68`): `68-04` e `68-05` canceladas e removidas da fila; heartbeat e aviso de preflight migram para `forge-sdd acp`/`attach` (tratado em 69-01).

## Fila e conflitos esperados
- `queue.json` é por branch e o `Enqueue` sobrescreve `feature_id` e `branch`. Nesta branch o arquivo foi reiniciado só com 69-01 a 69-07 (o histórico segue no git). Até a 69-03 entrar, mesclar com a `feat/68` conflita neste arquivo e em `progress.md`; a versão da 69 vence.
- `sdd/features/index.md` não é atualizado por esta iniciativa: é justamente o arquivo que a 69-03 retira do fluxo.

## Medição
Antes de iniciar a 69-02 e depois da 69-07, rodar `forge-sdd report` e comparar tokens de entrada por sessão de `/discovery` e de Builder. O resultado entra no PR.

## Critério final
`go build ./... && go vet ./... && go test -race ./...`

## Pós-Portão
Após `FLX-01`: `/split-features` grava as specs em `features/` e enfileira para o Builder. Sem PR até o humano retornar.

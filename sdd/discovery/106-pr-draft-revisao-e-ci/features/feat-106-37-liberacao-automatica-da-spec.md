# feat-106-37 — Liberação automática da spec: o [Seguir] tira do draft, move o card e despacha ao Builder em paralelo

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** `feat/106-37-liberacao-automatica-da-spec`
**Tamanho:** `L`
**Depende de:** `fix-106-03` (WIP humano), `fix-106-06` (modo da sessão e volta da spec), `fix-106-09` (atualização sem impacto), `feat-107-09` (teto de delegações)
**Status:** `spec` (Gate 1 aprovado em 22/11/2026, `[Seguir]`, com opt-in)

## Problema

Hoje, depois do `[Seguir]` do Gate 1, **cada passo é manual**: tirar o PR da spec do draft, mesclar, mover o card, rodar `forge dispatch`, subir workers. Na sessão de 22/11/2026 isso foi feito à mão para 8 specs, uma a uma, e o despacho nem chegou a rodar. O dono pediu: "na medida que for liberando o spec, um robô ou script dispara: PR ready em vez de draft, card na coluna Todo se houver GitHub Project, passa para a fila do Builder, e paraleliza o que der".

## Decisões do dono (22/11/2026)

- **Gatilho:** o `[Seguir]` do Gate 1. Aprovar a spec já é a decisão humana; o robô faz o resto sem nova pergunta.
- **Onde mora:** estende o `/toca-ai` e o laço da 107 (o robô é o "Início automático do Builder: `local`"). Roda local. **Sem GitHub Actions** (a ADR-0022 registra que foram bloqueadas por cobrança).
- **Card:** pelo `pm_mcp`, **só se configurado**; sem Project ou MCP, o passo é pulado em silêncio.

## O que já existe cobrindo parte

| Peça | Estado hoje |
|---|---|
| `forge dispatch` | roda o gate pré-build e enfileira em `builder/pending`; manual, uma iniciativa ou um fix por vez |
| `/toca-ai` | planeja, faz pré-voo único, enfileira etapas em ordem e executa sozinho; parte de **um objetivo novo**, não de spec já aprovada |
| `queue.json` | tem `depends_on` por task; é a fonte de ordem e de paralelismo |
| `builder.max_parallel_tasks` | `3` no `forge.json` |
| `forge worker --role builder` | reivindica um job por vez; **vários workers rodam juntos** e o claim é atômico (`TestClaim_UmSoJobEOitoWorkers`) |
| `pm_mcp` | `status_on_implement`, `status_on_pr_approved`; o card é movido **pelo agente via MCP**, o Go não move |
| `AutorizaVoltaDaSpec` e `VoltaSpecParaPrincipal` | decidem e executam a volta da spec à principal; **sem chamador** hoje (`fix-106-06` liga) |

## O que falta (a cola)

Nada liga o `[Seguir]` aos passos seguintes. O quadro como módulo (108 e 109) está parado, e o laço da 107 não foi implementado.

## Modelo do módulo

**Completo** (`ARQ-11`): uma função de domínio nova em `internal/pipeline` (liberação e seleção de tasks prontas), uma flag num subcomando existente e um passo no fluxo do agente. A dependência aponta para dentro.

## Decisões do projeto aplicadas

- `FLX-01` — o `[Seguir]` é a aprovação; o robô **não aprova** nada por conta própria e só age depois dele.
- `FLX-02` — uma branch e um PR em draft **por feature**: os PRs de **feature** do Builder **continuam em draft** até a Conferência 2 conforme e o `[Continuar]` do dono. O robô só tira do draft o **PR da spec**.
- ADR-0022 — a spec volta à principal após o `[Seguir]`; CI cara só fora do draft.
- `DUP-01` — **um** caminho de despacho: o robô chama o `Dispatch` existente, não cria outro.
- `ARQ-06` — **sem subcomando novo**: a liberação é a flag `--liberar` do `forge dispatch`.
- `ARQ-05` — `--dry-run` imprime o plano e não grava, não mescla e não enfileira.
- `SEG-01` — id da iniciativa e caminhos validados; nada fora de `sdd/discovery/`.
- **Compatibilidade (fix-106-09):** o comportamento novo é **opt-in**. O campo `Início automático do Builder` fica `desligado` por padrão; com `desligado`, o `[Seguir]` faz o que faz hoje. Projeto antigo não muda ao atualizar.
- Regra do dono: pergunta e decisão por componente de ask (`fix-106-07`).

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Gate e enfileiramento | `internal/pipeline/dispatch.go` (`Dispatch`) | gate pré-build + `builder/pending` | **reusar**; acrescentar `Liberar` por cima |
| Motor de etapas | `internal/pipeline/goal.go` (`EnqueueGoalSteps`) | enfileira um job por etapa, na ordem | **reusar** a seleção de etapas prontas |
| Volta da spec | `internal/pipeline/volta_da_spec.go` | autoriza e mescla via adaptador de VCS | **chamar** (ligado pelo `fix-106-06`) |
| Tirar do draft e mesclar | `internal/vcs` (`MarcarPronto`, `Mesclar`) | existe, com guarda de base (`fix-106-24d`) | **reusar** |
| Ordem e paralelismo | `queue.json` (`depends_on`), `builder.max_parallel_tasks` | existe | **reusar** |
| Mover o card | agente via MCP (`/implementar`) e `card-evidence.json` | existe, manual | **manter no agente**; acrescentar o passo no fluxo do `[Seguir]` |
| Status do card | `internal/config/forge.go` (`pm_mcp.status_on_*`) | só `implement` e `pr_approved` | **acrescentar** `status_on_spec_approved` (opcional; vazio = pular) |
| Subir vários workers | `forge up`, `forge worker` | um worker por processo | **reusar** para subir até N workers headless (a confirmar) |
| CI em PR só de spec | `.github/workflows/ci.yml`, `scripts/ci-politica.py` | matriz de 3 sistemas em qualquer PR fora do draft | **estender**: PR que só muda `sdd/**` e `*.md` não roda a matriz |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-02`, `ARQ-05`, `ARQ-06`, `ARQ-11`, `COD-03`, `DUP-01`, `DUP-02`, `FLX-01`, `FLX-02`, `SEG-01`, `SEG-02`, `TST-01`, `TST-03`, `TST-05`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`, `sdd-modos-de-projeto`.

## Escopo

**Dentro — a sequência do `[Seguir]`, só quando `Início automático do Builder` = `local`:**

1. **Conferir (gate):** `Dispatch` em modo gate (discovery, criteria, plan, cinco seções, IDs citados existem). Se reprova, **para** e mostra as pendências em menu; nada é mesclado nem enfileirado.
2. **PR da spec:** `MarcarPronto` e, conforme o modo, `AutorizaVoltaDaSpec` e `VoltaSpecParaPrincipal`. Respeita a guarda de base (`fix-106-24d`).
3. **Card (agente, opcional):** se há `pm_mcp` e Project configurados, o agente move o card para `pm_mcp.status_on_spec_approved` (ex.: `Todo`) e grava a evidência. Sem configuração, **pula em silêncio**.
4. **Despacho:** `forge dispatch <iniciativa> --liberar` enfileira as tasks **prontas** (status `queued`, todas as `depends_on` aprovadas), **até** `min(builder.max_parallel_tasks, prontas)`.
5. **Paralelização:** sobe até N workers de Builder headless, cada um em **worktree próprio** e **branch e PR em draft por feature**. Quando uma task é aprovada, as que dependiam dela viram prontas e são enfileiradas, até o teto, sem nova ação do dono.
6. **Parada e visibilidade:** respeita `.forge/runtime/STOP`; `forge status` mostra "N em execução, M na fila, K bloqueadas". No teto de tentativas (3 por delegação), escala em menu.
7. **CI em PR só de spec:** o job `test` decide por `git diff --name-only` e sai verde **sem rodar a matriz** quando só mudaram `sdd/**` e `*.md`. Remove o vermelho de macOS e Windows de PRs que não tocam código.

**Fora:**
- Tirar do draft o PR de **feature**, ou mesclar código: continua do dono (`[Continuar]`).
- GitHub Actions, GitHub Project como módulo (108/109) e o laço completo de roteamento da 107 (só o disparo de despacho entra aqui).
- Aprovar spec sem o `[Seguir]` (isso é o modo `sem-humano` do `fix-106-06`, que tem aviso de riscos).
- Resolver conflito entre tasks paralelas que mexem no mesmo arquivo (ver R2).

**A confirmar na implementação:** como `forge up` sobe workers headless e se aceita uma contagem; se o `pm_mcp` do projeto expõe a coluna `Todo` sob esse nome.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. Opt-in: com 'desligado', o [Seguir] não despacha nem mescla nada novo (compatibilidade).
go test ./internal/pipeline/... -run 'TestLiberarComInicioDesligadoNaoFazNada'
# 2. Com 'local': gate, volta da spec e enfileiramento, nessa ordem.
go test ./internal/pipeline/... -run 'TestLiberarGateDepoisVoltaDaSpecDepoisDespacho'
# 3. Gate reprova: nada mescla e nada enfileira.
go test ./internal/pipeline/... -run 'TestLiberarGateReprovadoNaoMesclaNemEnfileira'
# 4. Paralelismo respeita depends_on e o teto.
go test ./internal/pipeline/... -run 'TestLiberarSoTasksProntasAteOTeto|TestDependenteEntraQuandoPredecessoraAprova'
# 5. Card: sem MCP pula em silêncio.
go test ./internal/pipeline/... -run 'TestLiberarSemPmMcpPulaOCard'
# 6. PR de feature NUNCA sai do draft por aqui.
go test ./internal/pipeline/... -run 'TestLiberarNaoTiraPRDeFeatureDoDraft'
# 7. --dry-run não grava, não mescla, não enfileira.
go test ./cmd/forge-sdd/... -run 'TestDispatchLiberarDryRunNaoGrava'
# 8. CI de PR só de spec não roda a matriz.
go test ./scripts/... -run 'TestCIPoliticaPulaMatrizQuandoSoMudaSddEMarkdown'
# 9. Sem subcomando novo (ARQ-06 intacto).
! grep -rn 'Use: *"liberar' cmd/forge-sdd || exit 1
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** duas tasks A e B independentes e C dependente de A, teto 2: o robô enfileira **A e B**, **não** C; só enfileira C depois que A é aprovada. Com a spec **sem** a seção "O que já existe" (`DUP-02`), nada é mesclado.
- **Prova do opt-in:** num projeto com `Início automático do Builder | desligado`, o fluxo do `[Seguir]` produz exatamente o mesmo resultado de antes (comparação byte a byte da árvore).

## Riscos

- **R1 — Disparo indesejado.** Mitigação: opt-in por campo da Constituição, o `[Seguir]` explícito, o gate antes de mesclar, `--dry-run` e a parada por `STOP`.
- **R2 — Tasks paralelas mexem nos mesmos arquivos** e geram conflito de merge. Mitigação: só `depends_on` define o que roda junto; cada task tem worktree e PR próprios; o conflito aparece na revisão, não em silêncio. Evolução possível: declarar os arquivos de cada task no plan.
- **R3 — Custo.** N workers de Builder rodam ao mesmo tempo. Mitigação: o teto `max_parallel_tasks` (padrão 3) e `max_jobs_per_day`.
- **R4 — O filtro de CI esconde falha de PR misto** (docs e código juntos). Mitigação: qualquer arquivo fora de `sdd/**` e `*.md` liga a matriz inteira.
- **R5 — Dependência de specs ainda não implementadas** (`106-06`, `107-09`). Mitigação: a ordem de implementação é parte da spec; sem elas, o disparo fica desligado.

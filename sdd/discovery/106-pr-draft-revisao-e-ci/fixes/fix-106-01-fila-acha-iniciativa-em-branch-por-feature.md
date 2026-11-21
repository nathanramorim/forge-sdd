# fix-106-01 — A fila acha a iniciativa em branch por feature

**Iniciativa:** 106-pr-draft-revisao-e-ci
**Branch:** `fix/fila-acha-iniciativa-em-branch-por-feature`
**Origem:** relato do Builder da `feat-106-11`, verificado pela coordenação (21/11/2026)
**Tamanho:** `S`
**Status:** `spec` (Gate 1 aprovado em 2026-11-21)

## Problema

Com a exceção da Regra 15 (dono, até 2026-12-31) o Builder trabalha numa branch por feature, como `feat/106-11-parecer-com-head-sha`. Nela `forge queue list` diz "Fila vazia" e `forge queue next` não vê a iniciativa, embora `forge status` mostre a 106 com `queued=30`. O Builder não consegue marcar a feature como em andamento ou concluída pelo forge. Bloqueia as features 02 e 07a da 106 e toda feature futura com branch por feature.

## Feature Correlacionada

- Origem: `internal/layout/layout.go` (`FromBranch`, `Current`, `ActiveInitiative`, `QueuePath`).
- Causa-raiz: `FromBranch` (`layout.go:304-310`) aceita `feat|fix/<slug>` e devolve o slug inteiro (`106-11-parecer-com-head-sha`) como se fosse o nome da pasta. `ActiveInitiative` (`layout.go:222-236`) só confere se `sdd/discovery/<slug>` existe; como não existe, devolve `ok=false` e `QueuePath` (`layout.go:~244`) cai na fila legada `sdd/.queue/queue.json`, vazia.
- Evidência: `branchPrefixes = {feat, fix}` (`layout.go:163`); `initiativeSlug` aceita qualquer `<id>-<nome>` (`layout.go:160`); a única pasta é `sdd/discovery/106-pr-draft-revisao-e-ci`. `forge status` usa `ListInitiatives` (`scan.go:36`, varre o diretório), por isso enxerga a 106; a fila usa a branch.

## Quem chama (mapa do que muda)

- `layout.ActiveInitiative` -> `QueuePath`, `StatusPath` (fila e `status.md`), `cmd/forge-sdd/archive.go:34`, `internal/queue/queue.go:219` (Enqueue).
- `layout.Current` -> `ActiveInitiative` e `scan.go:36` (marca `Current` em `ListInitiatives`).
- `layout.FromBranch` direto: `internal/pipeline/archive.go:193` (exige `id == name`), `queue.go:218`, `scan.go:323` (`idOfRef`).
- `forge next-id` não usa branch (conta pastas); fora do defeito.
- A fila fica em `sdd/discovery/<id>/queue.json` (`layout.QueueFileName`), nunca por feature hoje.

## Decisões do projeto aplicadas

`DUP-02`/`DUP-06`: um só lugar deduz a iniciativa (`layout`); estende-se, sem segundo parser. `ARQ-02`: regra no domínio, `cmd/` fino. `SEG-01`: o id vira caminho, então vale validação ancorada. `TST-03`: casos conformes e negativos. `GOV-02`: nada aqui altera a Constituição.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Dedução da iniciativa pela branch | `layout.FromBranch`/`Current`/`ActiveInitiative` | Casa só nome exato | Estender `ActiveInitiative` |
| Lista de pastas | `layout.ListInitiatives`, `Discovery(dir)` | Existe | Reusar para achar `<N>-*` |
| Escape por flag/env | `cmd/forge-sdd` | Só `--dir` | Criar `--iniciativa` e `FORGE_INICIATIVA` |
| Resolver na 106/07a | `feat-106-07a` | Estado por feature e claim pela branch; NÃO trata a dedução da iniciativa (assume que ela já funciona) | Este fix vem antes; a 07a entra DEPOIS dele |

## Desenho mínimo (Gate 1 aprovado em 2026-11-21)

**(a) Prefixo numérico.** Nova função em `layout`, `ResolveInitiative(dir) (id, path string, err error)` (assinatura única; a branch é lida DENTRO dela, via `CurrentBranch(dir)`, e a regra de casamento mora numa função auxiliar `initiativeOfBranch(dir, branch)` reusada pelo `Enqueue`), usada por `ActiveInitiative` (e portanto por `QueuePath`, `StatusPath`, archive, Enqueue). Ordem: 1) pasta `sdd/discovery/<slug>` exata (comportamento de hoje, sem mudança); 2) se não existir, extrair `N` com regex ancorada `^(?:feat|fix)/([0-9]{1,6})-[a-z0-9-]+$` e casar com as pastas `sdd/discovery/N-*` (comparar o primeiro segmento, não prefixo de texto: `10` não casa `106`); 3) exatamente uma pasta -> é a iniciativa; zero -> `ok=false` (legado, como hoje); duas ou mais -> erro claro listando as pastas ("branch X é ambígua: A, B; use --iniciativa"). `FromBranch` permanece puro e inalterado (não regride `archive.go:193` nem `scan.go:323`).

**(a2) Como a ambiguidade chega ao usuário.** `ActiveInitiative(dir) (id, path, ok)`, `QueuePath` e `StatusPath` não devolvem erro; sozinhos transformariam a ambiguidade em `ok=false` e a fila legada, o que a DECISÃO 3 proíbe. Por isso `ResolveInitiative(dir) (id, path string, err error)` distingue três saídas: iniciativa achada; "sem iniciativa" (`id==""`, `err==nil`, vai à fila legada, comportamento atual); `ErrAmbiguousInitiative` com a lista das pastas que casaram. As assinaturas antigas ficam como adaptadores finos sobre ela (ambiguidade => `ok=false`), para não quebrar chamadores. Os comandos que usam a fila (`queue enqueue/next/complete/review`, e `queue move` quando existir, `status`, `archive`, `doctor`, `gate`) chamam UMA verificação única (`layout.RequireInitiative(dir) error`) antes de ler a fila, via o mesmo ponto de entrada em `cmd/forge-sdd`, e saem com código de erro diferente de zero e a lista das pastas. Em ambiguidade nada lê nem grava `sdd/.queue/queue.json`.

**(b) Escape.** UM mecanismo, em `layout`, sem segundo parser: a flag persistente `--iniciativa` no comando raiz grava o valor validado em `layout.SetOverride(id)`, chamada em um só `PersistentPreRunE`. `ResolveInitiative` lê, nessa ordem, o override da flag, `os.Getenv("FORGE_INICIATIVA")` e a branch; flag e variável são validadas por `ValidInitiative` e exigem pasta existente (senão erro, sem cair na branch). DECISÃO: não há campo `Iniciativa` em `.forge`; bastam flag e variável.

**(c) Compatibilidade.** Branch por iniciativa (`feat/106-pr-draft-revisao-e-ci`) casa no passo 1, idêntica a hoje. Branches sem prefixo numérico (`main`, `fix/ordem-da-fila-e-prioridade`, `chore/x`) não casam no passo 2 e seguem para a fila legada. 107 e 106/0.7 não mudam: continuam lendo `QueuePath`.

**(d) `queue enqueue/next/complete/review`.** Nenhum muda de código: todos chegam à fila por `QueuePath`/`ActiveInitiative`. Só `Enqueue` (`queue.go:218-223`) compara a iniciativa da branch informada com a ativa; passa a usar `ResolveInitiative(dir)` para a iniciativa ativa e `initiativeOfBranch(dir, branch)` para a branch informada para não rejeitar `feat/106-11-...` como "outra iniciativa". Fora isso, o comportamento onde a branch já casa é idêntico (teste de não regressão).

**(e) Segurança.** Branch vem do git (não confiável): regex ancorada com dígitos no prefixo e `len <= 200` (o código atual NÃO limita o tamanho; o limite é acrescentado aqui e coberto por `TestResolveRecusaBranchHostil`, que inclui branch com 201 caracteres), sem `/` além do prefixo, sem `..`, sem caractere de controle; o id devolvido sai SEMPRE da listagem real de pastas (nunca da branch) e passa por `Initiative()` (SEG-01). Flag/env: mesma validação.

**(g) Chamadores indiretos e marcador "atual".** `internal/gate/gate.go:545` e `cmd/forge-sdd/doctor.go:165` usam `layout.StatusPath` e herdam a correção; teste do gate numa branch por feature. O marcador `Current` de `ListInitiatives` (`internal/layout/scan.go:36`, via `Current`) seguiria falso em `feat/106-11-...`; passa a usar o resolvedor (mudança pequena; em ambiguidade o marcador fica falso, sem erro, pois é só listagem).

**(f) Colisões.** `feat-106-07a` e `fix-104-17` (PR #158, `queue move`): editam `queue.go`/`taskstatus.go`/`waves.go`; este fix toca `internal/layout/layout.go` e uma linha de `queue.go:218-223`. Conflito textual pequeno em `queue.go` apenas; sem dependência de ordem, DECISÃO: este fix entra ANTES da `feat-106-07a`, para ela ser testável em branch por feature. O `fix-104-17` (#158, só spec) já está na `main`, mas `queue move` ainda NÃO existe como código: a colisão é só de arquivo (`queue.go`). Quando `queue move` existir, usa a mesma verificação única.

## Rules e IDs da Constituição Aplicáveis

`DUP-02`, `DUP-06`, `ARQ-02`, `SEG-01`, `TST-03`.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go build ./...
# negativo: o teste da branch por feature existe de fato e passa
go test ./internal/layout -list 'TestResolveIniciativaPorPrefixoNumerico' | grep -F 'TestResolveIniciativaPorPrefixoNumerico'
go test ./internal/layout -run 'TestResolveIniciativaPorPrefixoNumerico' -count=1
# negativo: duas pastas com o mesmo prefixo dão erro de ambiguidade, sem escolher uma
go test ./internal/layout -list 'TestResolvePrefixoAmbiguoFalha' | grep -F 'TestResolvePrefixoAmbiguoFalha'
go test ./internal/layout -run 'TestResolvePrefixoAmbiguoFalha' -count=1
# negativo: branch sem prefixo numérico mantém o comportamento atual (fila legada)
go test ./internal/layout -list 'TestBranchSemPrefixoNumericoMantemComportamento' | grep -F 'TestBranchSemPrefixoNumericoMantemComportamento'
go test ./internal/layout -run 'TestBranchSemPrefixoNumericoMantemComportamento' -count=1
# negativo: branch hostil (controle, "..", prefixo não numérico, 10 vs 106) é recusada
go test ./internal/layout -list 'TestResolveRecusaBranchHostil' | grep -F 'TestResolveRecusaBranchHostil'
go test ./internal/layout -run 'TestResolveRecusaBranchHostil' -count=1
# não regressão: branch por iniciativa (nome exato) resolve como antes
go test ./internal/layout -list 'TestBranchPorIniciativaNaoRegride' | grep -F 'TestBranchPorIniciativaNaoRegride'
go test ./internal/layout -run 'TestBranchPorIniciativaNaoRegride' -count=1
# negativo: em ambiguidade o comando falha e NÃO lê nem grava a fila legada (sdd/.queue/queue.json inalterado)
go test ./cmd/forge-sdd -list 'TestQueueAmbiguidadeFalhaSemTocarFilaLegada' | grep -F 'TestQueueAmbiguidadeFalhaSemTocarFilaLegada'
go test ./cmd/forge-sdd -run 'TestQueueAmbiguidadeFalhaSemTocarFilaLegada' -count=1
# precedência flag > variável > branch
go test ./internal/layout -list 'TestPrecedenciaFlagVariavelBranch' | grep -F 'TestPrecedenciaFlagVariavelBranch'
go test ./internal/layout -run 'TestPrecedenciaFlagVariavelBranch' -count=1
# gate em branch por feature herda a correção
go test ./internal/gate -list 'TestGateEmBranchPorFeature' | grep -F 'TestGateEmBranchPorFeature'
go test ./internal/gate -run 'TestGateEmBranchPorFeature' -count=1
# marcador atual de ListInitiatives em branch por feature
go test ./internal/layout -list 'TestListInitiativesMarcaAtualEmBranchPorFeature' | grep -F 'TestListInitiativesMarcaAtualEmBranchPorFeature'
go test ./internal/layout -run 'TestListInitiativesMarcaAtualEmBranchPorFeature' -count=1
# flag e variável de escape; a flag vence a variável
go test ./cmd/forge-sdd -list 'TestQueueIniciativaFlagEEnv' | grep -F 'TestQueueIniciativaFlagEEnv'
go test ./cmd/forge-sdd -run 'TestQueueIniciativaFlagEEnv' -count=1
# queue list e next enxergam a fila numa branch por feature
go test ./cmd/forge-sdd -list 'TestQueueNextEmBranchPorFeature' | grep -F 'TestQueueNextEmBranchPorFeature'
go test ./cmd/forge-sdd -run 'TestQueueNextEmBranchPorFeature' -count=1
go vet ./... && go test ./...
```
- Cada linha `# negativo:` prova a recusa; sem ela o fix não passa (`TST-03`).

## Fora de escopo

Campo `Iniciativa` em `.forge`; estado por feature (`feat-106-07a`); edição das specs 106/107; `queue move` (`fix-104-17`).

## Decisões do dono (Gate 1, 2026-11-21)

DECISÃO 1: casar a iniciativa pelo prefixo numérico `<N>-*` único. DECISÃO 2: bastam `--iniciativa` e `FORGE_INICIATIVA`, sem campo em `.forge`. DECISÃO 3: ambiguidade (duas pastas com o mesmo número) FALHA com erro claro listando as pastas que casaram, sem cair na fila legada. Ordem: a `feat-106-07a` entra DEPOIS deste fix.

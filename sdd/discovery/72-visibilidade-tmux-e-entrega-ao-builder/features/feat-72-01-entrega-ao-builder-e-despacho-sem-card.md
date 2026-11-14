# feat-72-01 — Entrega ao Builder e Despacho sem Card

**Branch:** `feat/72-visibilidade-tmux-e-entrega-ao-builder`
**Fase:** 72-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
No uso real, a Discovery ficou sem próximo passo depois do split, tentou executar sozinha, editou `forge.json` e fabricou um `card-evidence.json`, porque o `dispatch` exige o arquivo mesmo com `require_existing_card: false`. Esta task faz o split entregar ao Builder e garante em Go que a Discovery não constrói.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão de código existente (`dispatch`, `queue`) e de skills, sem pacote novo nem concorrência (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Gate de despacho | `internal/pipeline/dispatch.go` (`loadEvidence`, chamado sempre) | Exige o arquivo de evidência mesmo com o card desligado | corrigir | com `require_existing_card: false` pula a evidência e o job nasce sem card |
| Comandos de fila | `cmd/forge-sdd/queue.go` (`next`, `complete`) | Funcionam em qualquer sessão | estender | guarda de modo pipeline: `FORGE_WORKER` ou `--local` |
| Skill do split | `.agents/skills/sdd-planejamento/references/split-features.md` | Termina em `queue enqueue` | estender | passo de `forge-sdd dispatch <iniciativa>` em modo pipeline, depois da confirmação humana |
| Skill do implementar e da concepção | `.agents/skills/sdd-build/references/implementar.md`, `sdd-concepcao` | `implementar` já diz que não constrói em modo pipeline | estender | a Discovery não edita `forge.json`, `card-evidence.json` nem o código da feature |
| Variável do worker | `internal/pipeline/runner.go` (`FORGE_WORKER`) | O runner define `FORGE_WORKER` para o agente | reusar | é o sinal de que o comando roda sob um worker |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-01`, `FLX-08`, `SEG-01`, `DUP-06`, `TST-02`, `TST-03`.
- Rules: `sdd-fluxo-e-portao.md`, `sdd-seguranca.md`.
- Gerais: `FLX-02`, `COD-02`, `TST-05`.
- Skill: `novo-subcomando-cli-sdd` (flag nova em subcomando existente).

## Escopo
1. `dispatch`: com `pm_mcp.require_existing_card: false`, não lê nem exige `card-evidence.json`; o job é criado com `card` vazio e a saída diz que o despacho foi sem card. Com `true`, o comportamento atual se mantém.
2. `forge-sdd queue next` e `forge-sdd queue complete`: com `.forge/config/forge.json` presente e sem `FORGE_WORKER` no ambiente, recusam com uma mensagem curta ("modo pipeline: o Builder constrói; use forge-sdd dispatch <iniciativa>"); `--local` libera para o uso consciente. Sem `forge.json` nada muda.
3. `split-features.md`: novo passo em modo pipeline, depois da confirmação humana que enfileira as tasks: rode `forge-sdd dispatch <iniciativa>` (gate e job), informe "entregue ao Builder" e pare; não implemente nesta sessão. Pendência do gate vai para a pessoa.
4. `implementar.md` e a skill de concepção: uma frase dizendo que, em modo pipeline, a Discovery nunca edita `forge.json`, `card-evidence.json` nem o código da feature.
5. Templates e goldens atualizados; os adaptadores seguem dentro do orçamento do `P-24`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./cmd/forge-sdd/... ./internal/scaffold/...
```
- `TestDispatch_SemCardObrigatorio`: card desligado, sem arquivo de evidência, job criado sem card; card ligado continua bloqueando sem evidência.
- `TestQueue_GuardaModoPipeline`: `next` e `complete` recusam sem `FORGE_WORKER` com `forge.json`; `--local` e `FORGE_WORKER` liberam; sem `forge.json` nada muda.
- Teste de conteúdo: `split-features.md` cita `forge-sdd dispatch` e a regra "não implemente nesta sessão".
- `forge-sdd gate` com `P-24` ✓; goldens atualizados (`TST-02`).

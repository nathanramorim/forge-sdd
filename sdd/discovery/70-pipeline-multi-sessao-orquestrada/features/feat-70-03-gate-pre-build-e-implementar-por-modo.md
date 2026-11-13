# feat-70-03 — Gate Pré-Build, `dispatch` e `/implementar` por Modo

**Branch:** `feat/70-pipeline-multi-sessao-orquestrada`
**Fase:** 70-03
**Tamanho:** `M`
**Depende de:** feat-70-01, feat-70-02
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
`/implementar <feature>` não pode enfileirar sem spec pronta, ADRs e rules mapeadas e um card existente. O Go faz a metade determinística; o agente, que alcança o MCP, grava a evidência do card. O comando continua com o comportamento da 4.0.0-beta quando não há `forge.json`.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** novo pacote `internal/pipeline` com uma função de despacho que compõe peças existentes (gate, contexto, barramento); sem concorrência própria (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Gate de spec | `internal/gate` (`ValidateFLX01`, P-02, P-03) | Existência da spec e tabela `O que já existe` | reusar | O gate pré-build chama as mesmas funções |
| IDs, rules e skills da task | `gate.BuildContext` | Resolve IDs, rules e skills com confinamento de caminho | reusar | Fonte das `rules_refs` do job |
| Validação da spec na fila | `internal/queue` (`ValidateSpecReuseCheck`) | P-02 e P-03 | reusar | Cada task passa por ela |
| Barramento | `internal/bus` (70-02) | `Enqueue` | reusar | Cria o job em `builder/pending` |
| Configuração | `internal/config` (70-01) | `LoadForgeConfig` | reusar | Modo e limites |
| MCP do gerenciador | `sdd/memory/mcps.md` | Só o agente alcança o MCP | criar contrato de evidência | `card-evidence.json` validado pelo Go (D-2) |
| Comando `/implementar` | `.agents/skills/sdd-build/references/implementar.md` | Executa a fila na sessão (4.0.0-beta) | estender | Modo pipeline com `forge.json` |
| Resolução de iniciativa por nome | `sdd-build` (`/implementar <nome>`) e `layout.ListInitiatives` | Busca em `sdd/discovery/*<nome>*` | reusar | Mesma regra de desambiguação |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-01`, `FLX-08`, `DUP-02`, `DUP-03`, `SEG-01`, `SEG-04`, `ARQ-10`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-fluxo-e-portao.md`, `sdd-seguranca.md`, `sdd-arquitetura.md`.
- Skill: `novo-subcomando-cli-sdd`.
- Gerais: `FLX-02`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. `forge-sdd dispatch <feature> [--card-evidence <arquivo>] [--dry-run]`. Sem `forge.json`: erro orientando `forge-sdd pipeline init`.
2. **Gate determinístico** (lista todas as pendências de uma vez): iniciativa única encontrada; `discovery.md`, `criteria.md`, `plan.md` e ao menos uma task em `features/`; cada task com as cinco seções obrigatórias; refs de ADR citadas existem em `adr_dirs` (padrão `sdd/spec/adr`) e rules citadas existem em `.agents/rules/`.
3. **Evidência de card** (`card-evidence.json`): `provider` dentro de `pm_mcp.supported`, `id`, `url`, `status`, `feature` igual à feature, `checked_at` dentro de `evidence_max_age_minutes` (padrão 30). Sem a evidência, ou vencida, bloqueia. Segredos no arquivo são recusados.
4. **Despacho:** monta o job (`feature_build`, `normal`, branch `feat/<id>-<nome>`, `base_branch`, `max_rework_attempts` da configuração) e enfileira em `builder/pending`. A saída instrui o agente a mover o card para `status_on_implement` via MCP.
5. `references/implementar.md` do `sdd-build` descreve os dois modos e os adaptadores `/implement`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./cmd/forge-sdd/...
go test ./internal/pipeline -run 'TestDispatch_SemCardBloqueia|TestDispatch_ComCardCriaJob'
```
- Bloqueio (cenário 1): sem evidência, evidência vencida, `feature` diferente, spec sem `O que já existe`, ADR ou rule inexistente; nada é criado em `builder/pending`; a mensagem lista as pendências.
- Despacho (cenário 2): job criado com `rules_refs` e `adr_refs` corretos e saída com a instrução do card.
- `--dry-run` não grava; nome de feature ambíguo lista as opções.
- Violações (`TST-03`): `feature` com `..`, evidência com segredo, `provider` fora da lista.
- Sem `forge.json`, o texto da skill e o comando mantêm o fluxo da 4.0.0-beta.

# feat-69-07 — Orçamento de Contexto e Constitution Enxuta

**Branch:** `feat/69-otimizacao-estrutura-tokens-e-paralelismo`
**Fase:** 69-07
**Tamanho:** `M`
**Depende de:** feat-69-06
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Um `/discovery` obriga a ler cerca de 65 KB e o Builder cerca de 35 KB por task. A Constitution (14 KB) escreve cada regra duas vezes (Regras 0 a 15 e IDs canônicos). Esta task fixa orçamentos verificáveis, entrega o comando que monta o contexto mínimo por task e fecha as divergências de documentação encontradas no discovery.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** um subcomando (`context`), checks de orçamento no gate e um teste; reaproveita o gate e o scaffold existentes (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Gate Camada A | `internal/gate/gate.go` (P-01..P-23) | Checks determinísticos em modo `observacao`/`bloqueio` | estender | Novos checks de orçamento em `observacao` |
| Medição de tokens | `forge-sdd session record`, `forge-sdd report` | Tokens de entrada por feature em `sdd/.metrics/` | reusar | Comparar antes e depois, sem ferramenta nova |
| Montagem de contexto | `.agents/rules/README.md`, frontmatter `globs` e `regras` | Rules carregadas manualmente | criar `context` | Subcomando resolve spec + IDs + rules + skill |
| Constitution | `sdd/memory/constitution.md` | 14 KB, regras 0 a 15 duplicadas pelos IDs | estender | Cada regra escrita uma vez, por ID |
| Teste de golden | `internal/scaffold/scaffold_test.go` | Compara templates com goldens | estender | `TestContextBudget` soma o conjunto de leitura |
| Divergências | `sdd/spec/concepts.md`, `ARQ-06`, mapa de rules | `internal/doctor`, `internal/telemetry`, `internal/lessons` inexistentes; `ARQ-06` sem `acp`, `attach`, `autopilot`, `governance`, `destroy`; `sdd-acp.md` fora do mapa | corrigir | Registrar na ADR 0003 |
| Procedimento | `.agents/skills/novo-subcomando-cli-sdd` | Aprovado | reusar | `DUP-05` |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-06`, `ARQ-10`, `GOV-01`, `GOV-04`, `FLX-05`, `FLX-06`, `DUP-06`, `TST-01`, `TST-03`, `TST-05`, `SEG-01`.
- Rules: `sdd-fluxo-e-portao.md`, `sdd-codigo-e-testes.md`, `sdd-reuso-e-duplicacao.md`.
- Skill: `novo-subcomando-cli-sdd`.
- Gerais: `FLX-01`, `COD-02`, `COD-05`.

## Escopo
1. `forge-sdd context <task>`: imprime a spec da task, as linhas dos IDs citados (extraídas da Constitution) e os caminhos das rules e da skill aplicáveis.
2. Constitution: remover a duplicação entre as Regras 0 a 15 e os IDs, mantendo cada regra uma vez; ≤ 9 KB; corrigir o mapa de rules (incluir `sdd-acp.md`) e `ARQ-06`.
3. Checks de gate em modo `observacao`: instruções raiz ≤ 1.536 bytes, `SKILL.md` ≤ 2.048 bytes, adaptadores ≤ 300 bytes.
4. `TestContextBudget`: soma dos arquivos de leitura de `/discovery` ≤ 22 KB e do Builder ≤ 15 KB nos templates.
5. Corrigir `sdd/spec/concepts.md` (pacotes e tipos inexistentes) e registrar tudo em `sdd/spec/adr/adr-0003-skills-por-estacao-e-orcamento.md` (`GOV-01`).
6. Medir: rodar `forge-sdd report` antes e depois e registrar os números no PR.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./...
go test ./internal/scaffold -run TestContextBudget
test "$(wc -c < sdd/memory/constitution.md)" -le 9216
forge-sdd context 69-07 | head -20
forge-sdd gate
```
- Violações (`TST-03`): orçamento estourado faz o check reportar falha; ID inexistente em `context` retorna erro enriquecido.
- Cada ID da Constitution anterior continua presente (teste compara a lista de IDs antes e depois).
- ADR 0003 cita `GOV-01` e as divergências corrigidas.

# fix-88-02 — Tetos do modo avançado desligados por padrão e todos no Settings

**Branch:** `fix/88-02-tetos-desligados-por-padrao-no-settings` (empilhada sobre `fix/88-01-teto-diario-nao-deve-parar-a-esteira`, PR 85; PR contra ela até o 85 ser mesclado)
**Origem:** Iniciativa 88, `feat-88-07` e `fix-88-01` (`internal/pipeline/safeguards.go`, `internal/config/forge.go`, `cmd/forge-sdd/menu.go`)
**Tamanho:** `S`
**Status:** `ready`

## Sintoma
Decisão do dono do produto depois do `fix-88-01`: nenhum teto deve parar a esteira sem ele pedir. Hoje o modo avançado nasce com o teto de tokens por job (2.000.000) e o de tempo por job (60 min) ligados, e os dois encerram o worker ("Esteira parada").

## Feature Correlacionada
- **Arquivo de origem:** `feat-88-07-salvaguardas-e-docs.md` e `fix-88-01-teto-diario-nao-deve-parar-a-esteira.md` (que já põe o teto diário no Settings, desligado).
- **Causa-raiz:** `DefaultForgeConfig` grava `Safeguards.MaxJobTokens: 2000000`; `JobOverCeiling` (`safeguards.go`) reaproveita `builder.job_timeout_minutes` (o timeout do runner, vigente em todos os modos) como teto que encerra a esteira. Não existe chave própria para o teto de tempo, então não dá para desligá-lo sem mexer no timeout do runner.
- **Evidência:** `internal/config/forge.go` (padrões), `internal/pipeline/safeguards.go:115-125`, `worker.go:227` e `deliver.go:332` (o timeout do runner, separado).

## Modelo do módulo
**Enxuto** (ARQ-11): ajuste em `internal/config`, `internal/pipeline/safeguards.go` e no Settings; sem módulo novo.

## Decisões do projeto aplicadas
ADR-0013 (salvaguardas do modo avançado), fix-88-01. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Teto de tokens por job | `Safeguards.MaxJobTokens`, `JobOverCeiling` | Padrão 2.000.000 | **mudar** o padrão para `0` (desligado) | Decisão do dono: tudo desligado |
| Teto de tempo que encerra a esteira | `JobOverCeiling` usa `builder.job_timeout_minutes` | Mistura o timeout do runner com a parada | **separar**: nova chave `safeguards.max_job_minutes` (0 = desligado); `builder.job_timeout_minutes` segue só como timeout do runner | Desligar a parada não pode tirar o timeout do runner |
| Chaves liga/desliga | menu Settings (feat-88-08, fix-88-01) | Card e teto diário | **estender** com tokens por job e tempo por job | RN-9 |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `COD-01`, `COD-03`, `DUP-01`, `TST-01`.

## Premissas (a confirmar no Gate 1)
- Sem nenhum teto ligado, só o `STOP` encerra a esteira avançada.
- `builder.job_timeout_minutes` (timeout do runner) não muda e continua valendo em todos os modos.
- `forge.json` existente com os valores gravados os mantém.

## Escopo
1. `DefaultForgeConfig`: `MaxJobTokens` = 0; novo `Safeguards.MaxJobMinutes` = 0 (com `omitempty` e validação).
2. `JobOverCeiling`: tempo usa `MaxJobMinutes` (0 = sem teto), não `builder.job_timeout_minutes`.
3. Settings: chaves "Teto de tokens por job" e "Teto de tempo por job" (desativado / valor), ao lado do teto diário.
4. Testes de regressão e atualização de `docs/pipeline.md` e do LEIA-ME.

## Fora do escopo
Mudar o `STOP`, o timeout do runner ou o modo normal.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/pipeline -run 'TestSalvaguardas|TestJobOverCeiling'
go test -v -race ./internal/config ./cmd/forge-sdd -run 'TestDefaultSafeguards|TestMenu_Settings'
go vet ./... && go test ./...
```
- Esteira avançada nova: nenhum teto ligado; job longo e log grande não encerram a esteira.
- Com `max_job_minutes` ligado, o job que passa do teto encerra a esteira com motivo no inbox.
- `builder.job_timeout_minutes` continua matando o job travado no runner.

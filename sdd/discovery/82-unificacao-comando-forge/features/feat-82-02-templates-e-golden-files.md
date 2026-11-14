# feat-82-02 — Atualização de Templates de Scaffold e Golden Files

**Branch:** `feat/82-unificacao-comando-forge`
**Fase:** 82-02
**Tamanho:** `M`
**Depende de:** 82-01
**Status:** `queued`

## Contexto & Motivação
Novos projetos inicializados pelo Forge-SDD geram arquivos de onboarding e automação através de templates (`internal/scaffold/templates/`). Vários templates ainda ensinam e referenciam `forge-sdd` (ex: `LEIA-ME.md.tmpl`, scripts de pipeline e prompts). Esta task atualiza esses templates para recomendar `forge` e atualiza a suíte de golden files em `internal/scaffold/testdata/golden/`.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica templates de scaffold distribuídos e golden files de teste (`DUP-10`, `TST-02`).

## Decisões do projeto aplicadas
- `ADR-0005`: Experiência unificada `forge`.
- `DUP-10`: Manutenção de golden files íntegros e dentro do orçamento de bytes.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Template LEIA-ME | `internal/scaffold/templates/.forge/LEIA-ME.md.tmpl` | Instruções `forge-sdd up`, etc. | estender | Padronizar para `forge up`, `forge status`, etc. |
| Scripts de up | `internal/scaffold/templates/` | Comandos `forge-sdd` | estender | Padronizar para `forge`. |
| Golden Files | `internal/scaffold/testdata/golden/` | Snapshots esperados dos templates | estender | Sincronizar com os novos templates atualizados. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-01`, `DUP-10`: Padronização canônica e integridade dos testes de scaffold.
- `TST-02`: Golden files testados e correspondentes aos templates reais.

## Critérios de Aceitação Executáveis
1. Todos os templates em `internal/scaffold/templates/` instruem o comando `forge` como comando primário.
2. `go test -v ./internal/scaffold/...` passa sem regressões ou divergências de golden files.

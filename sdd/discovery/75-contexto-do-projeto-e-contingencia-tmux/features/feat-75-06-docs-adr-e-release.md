# feat-75-06 — Docs, ADR e release

**Branch:** `feat/75-contexto-do-projeto-e-contingencia-tmux`
**Fase:** 75-06
**Tamanho:** `M`
**Depende de:** 75-01, 75-03, 75-05
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
Fechar a iniciativa: documentar o `brief`, a Constituição como primeiro passo e a contingência do tmux, registrar a decisão em ADR e fechar as notas de release.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** documentação, ADR e testes de ponta a ponta; sem código de produção novo (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Docs do pipeline | `README.md`, `docs/pipeline.md`, `.forge/LEIA-ME.md` (template), `site/public/guia-pipeline.html` | Descrevem menu, tmux, visão limpa | estender | `brief`, `/constitution` primeiro, instalação do tmux e contingência. |
| ADRs | `sdd/spec/adr/` (0001–0006, template) | Decisões anteriores | criar | Próxima ADR livre: contexto do projeto na Discovery e contingência do terminal. |
| Conceitos | `sdd/spec/concepts.md` | Tabela de conceitos reutilizáveis | estender | Linhas para `brief` e instalador do tmux (`DUP-10`). |
| Release notes | `sdd/releases/history.md` | Entrada `### Versão 4.2.0` | estender | Destaques reais da nova versão; bump a decidir (`4.2.1-beta` ou `4.3.0-beta`). |
| E2E | `internal/pipeline/e2e_test.go`, `pipeline_scaffold_test.go` | E2E do worker e do scaffold | estender | Projeto recém-gerado: `brief --check` rascunho e pronta; terminal com instalador falso. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-05`.
- Rules: `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md`.
- Gerais: `FLX-02`, `COD-02`.

## Escopo
1. README, `docs/pipeline.md`, `LEIA-ME.md.tmpl` e `guia-pipeline.html`: `forge-sdd brief` e `--check`; `/constitution` como passo 1; instalação do tmux com confirmação; contingência (Windows Terminal, janelas); WSL com tmux.
2. ADR nova (status `aprovado` só depois do Gate final) e linhas em `concepts.md`.
3. `sdd/releases/history.md`: destaques reais; versão nos arquivos de versão só com a decisão do bump (Regra 11: tag na branch, só com autorização).
4. E2E: scaffold novo, `brief --check` `rascunho`, povoar, `pronta`; terminal com instalador falso (instala, recusa, sem terminal, falha).
5. Goldens e orçamento `P-24` verdes.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./...
```
- `TestE2E_BriefProjetoNovo` e `TestE2E_TerminalSemTmux` passam.
- Teste de conteúdo: README, `docs/pipeline.md`, `LEIA-ME` e o guia citam `brief`, `/constitution` e a contingência do tmux.
- `forge-sdd gate` sem falhas; entrada de release com destaques reais (o workflow de release não cai no texto genérico).

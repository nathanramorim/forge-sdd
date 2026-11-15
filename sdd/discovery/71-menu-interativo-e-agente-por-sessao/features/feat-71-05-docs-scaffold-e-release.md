# feat-71-05 — Docs, Scaffold e Release 4.2.0

**Branch:** `feat/71-menu-interativo-e-agente-por-sessao`
**Fase:** 71-05
**Tamanho:** `M`
**Depende de:** 71-02, 71-03, 71-04
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Fecha a entrega: guia e README refletem o menu e o agente por sessão, o ADR registra as decisões e o que está a confirmar, e a versão sobe para `4.2.0-beta`.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** documentação, templates e goldens; sem código novo de domínio (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Guia no projeto | `internal/scaffold/templates/.forge/LEIA-ME.md.tmpl`, `docs/pipeline.md`, `README.md` | Passo a passo da 4.1 com `forge-sdd up --cli` | estender | passo a passo passa a começar por `forge` e pelo agente por sessão |
| Mensagem pós-instalação | `cmd/forge-sdd/main.go` (`printPipelineHint`) | Três passos do pipeline | estender | menciona `forge` e a escolha do agente |
| ADRs | `sdd/spec/adr/adr-0004-pipeline-multi-sessao.md` | Limites conhecidos do pipeline | estender com ADR novo | `adr-0005-menu-e-agente-por-sessao` registra D-1 a D-8 e a tabela de perfis |
| Conceitos e release notes | `sdd/spec/concepts.md`, `sdd/releases/history.md` | Linhas do pipeline; entrada `### Versão 4.1.0` | estender | linhas do menu e do seletor; entrada `### Versão 4.2.0 ... (Beta)` |
| Versão | `cmd/forge-sdd/main.go`, `internal/config/config.go`, `internal/acp/dispatcher.go`, `sdd/.sdd-version`, `sdd/.sddrc`, `npm/package.json` | `4.1.0-beta` | atualizar | bump para `4.2.0-beta` no commit de release (Regra 11) |

## Dependências novas
- Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-02`, `TST-05`.
- Rules: `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md`.

## Escopo
1. `.forge/LEIA-ME.md`, `docs/pipeline.md` e o README: `forge` como porta de entrada, o seletor por sessão, os níveis de permissão e a tabela de agentes com o que está confirmado e a confirmar; `forge-sdd up --discovery/--builder/--reviewer` para scripts.
2. `printPipelineHint` e a saída do `pipeline init` citam o menu.
3. ADR 0005; `concepts.md`; release notes `### Versão 4.2.0 — Menu Navegável e Agente por Sessão (Beta)` honestas: Gemini, Copilot e Codex não foram executados de ponta a ponta, só a montagem dos comandos.
4. Bump de versão para `4.2.0-beta` e goldens atualizados.
5. E2E: `go test ./cmd/forge-sdd -run TestE2E_MenuPipeline` percorre menu, seletor, checagem e geração dos três comandos com entrada simulada.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./... && forge-sdd gate
```
- Gate: `P-13`, `P-21`, `P-23`, `P-24` verdes.
- Teste de conteúdo: guia, README e release notes mencionam o menu e dizem o que não foi executado de ponta a ponta.
- `TestE2E_MenuPipeline` passa; goldens atualizados (`TST-02`).
- Release notes têm `### Versão 4.2.0` (exigido pelo workflow de release).

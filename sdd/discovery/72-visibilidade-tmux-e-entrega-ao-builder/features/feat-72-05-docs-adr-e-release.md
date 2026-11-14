# feat-72-05 — Docs, ADR 0006 e Release

**Branch:** `feat/72-visibilidade-tmux-e-entrega-ao-builder`
**Fase:** 72-05
**Tamanho:** `M`
**Depende de:** 72-01, 72-03, 72-04
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Fecha a entrega: o guia e o README refletem o tmux, a visão limpa, a pausa por limite de uso e a entrega do split ao Builder; o ADR registra as decisões; a entrada `4.2.0` das release notes passa a cobrir os discoveries 71 e 72.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** documentação, templates, goldens e um teste de ponta a ponta com agente falso (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Guia no projeto | `internal/scaffold/templates/.forge/LEIA-ME.md.tmpl`, `docs/pipeline.md`, `README.md` | Guia da 71 com o menu e o agente por sessão | estender | seções de tmux, o que cada painel mostra, limite de uso e "a Discovery não executa" |
| ADRs | `sdd/spec/adr/adr-0005-menu-e-agente-por-sessao.md` | Menu e perfis | estender com ADR novo | `adr-0006-visibilidade-tmux-e-entrega-ao-builder` registra D-1 a D-7 e o que está a confirmar |
| Release notes | `sdd/releases/history.md` | Entrada `### Versão 4.2.0` do discovery 71 | estender | a mesma entrada ganha os itens do 72, sem criar versão nova |
| Conceitos | `sdd/spec/concepts.md` | Linhas do pipeline, perfis e menu | estender | linhas de `LiveView`, `Requeue` e tmux |
| Teste de ponta a ponta | `internal/pipeline/e2e_test.go` | E2E do pipeline com agente falso | estender | worker com eventos de cada CLI e o caso de limite de uso |

## Dependências novas
- Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `FLX-05`, `FLX-07`, `GOV-01`, `TST-02`, `TST-05`.
- Rules: `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md`.

## Escopo
1. Guia (template, `docs/pipeline.md`, README): tmux (painéis, como sair e reconectar), o que cada painel mostra, `--raw`, o que o worker faz no limite de uso, e a regra de que a Discovery não executa o build; sem cópias divergentes entre os três.
2. ADR 0006 com D-1 a D-7, a guarda de modo, o que está confirmado e a confirmar (formatos de evento e mensagens de limite de uso por CLI) e o desvio, se houver.
3. `concepts.md` e a entrada `### Versão 4.2.0` das release notes (honesta: o que não foi executado de ponta a ponta continua dito; nada promete o que não está no código).
4. Versão permanece `4.2.0-beta`; goldens atualizados.
5. E2E: o worker com agente falso emite eventos de claude, gemini e codex e mostra as linhas esperadas; e um caso de limite de uso em que o job volta à fila, pausa e retoma.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./... && forge-sdd gate
```
- Gate: `P-13`, `P-21`, `P-23`, `P-24` verdes.
- Teste de conteúdo: guia, README e release notes citam tmux, visão limpa, limite de uso e a regra da Discovery, e dizem o que ficou a confirmar.
- `TestE2E_WorkerVisaoLimpaELimiteDeUso` passa; goldens atualizados (`TST-02`).
- Release notes têm `### Versão 4.2.0` (exigido pelo workflow de release).

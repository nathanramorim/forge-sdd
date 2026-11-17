# feat-90-07 — Modelo e esforço por sessão (absorve a iniciativa 85)

**Iniciativa:** 90-esteira-economica-e-autonoma
**Branch:** feat/90-esteira-economica-e-autonoma (a partir de `feat/88-modo-pipeline-avancado`)
**Status:** ready
**Depende de:** feat-90-02

## Problema
Builder e Reviewer usam `model: inherit`; não há como pôr modelo ou esforço menor nos papéis mecânicos. A 85 (arquivada) tinha a ideia, mas parou no Gate 1.

## Comportamento
1. `forge.json` ganha `models` e `efforts` (por papel `discovery`, `builder`, `reviewer`). Marcadores `{model}` e `{effort}` nos perfis expandem em zero ou mais argumentos, como `{permissions}`. `claude`: `--model`, `--effort low|medium|high|xhigh|max`; agy, gemini, codex e copilot: confirmar com `--help` no build, e só emitir flag que a CLI aceita.
2. Sem escolha, o argv é idêntico ao atual em todas as CLIs. Modelo fora de `^[A-Za-z0-9._:-]{1,80}$` ou esforço fora da lista da CLI é erro (`SEG-01`).
3. Settings (menu da 88) lista e grava por sessão, com sugestão econômica para Builder e Reviewer e "padrão da CLI" como primeira opção; oferece incluir o marcador em perfil gravado sem ele (como o `forge.json` deste repositório). O preflight avisa quando falta.

## Modelo do módulo
**Enxuto** (ARQ-11): `internal/config`, `internal/pipeline/runner.go`, `cmd/forge-sdd/menu.go`.

## Decisões do projeto aplicadas
ADR-0005 (D-1 marcadores, `CLIFor`), 88 RN-9 (Settings), ADR-0014.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Marcadores | `config.MarkerPermissions`, `expandArgs` | `{agent}`, `{name}`, `{prompt}`, `{permissions}` | estender | Mesmo padrão |
| Settings | `menu.go` (`settingKey`, `settingKeys`) | Chaves liga/desliga | estender com seleção | RN-9 |
| Discovery 85 | `sdd/archive/discovery/85-…` | Catálogo e presets | aproveitar, sem `agy models` nesta onda | Escopo |
| Preflight | `pipeline/preflight.go` | Confere CLI e arquivos | estender com o aviso | `DUP-01` |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`SEG-01`, `SEG-03`, `ARQ-01`, `DUP-01`, `TST-02`. Rules: `sdd-seguranca`, `sdd-arquitetura`, `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
go test ./internal/config ./internal/pipeline ./cmd/forge-sdd -run 'TestModeloPorSessao|TestEsforco|TestMenu_SettingsModelo'
go vet ./... && go test ./...
```
- Sem escolha, argv idêntico; `models.builder: sonnet` e `efforts.builder: low` geram `--model sonnet --effort low` no `claude`.
- Valores inválidos são erro; perfil sem `{model}` gera aviso.

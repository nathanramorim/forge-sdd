# feat-75-03 — Constituição primeiro

**Branch:** `feat/75-contexto-do-projeto-e-contingencia-tmux`
**Fase:** 75-03
**Tamanho:** `M`
**Depende de:** 75-01, 75-02
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
`/constitution` é opcional e só preenche `constitution.md` e `stack.md`. Para a Discovery "já entender logo depois de gerado", o primeiro comando num repositório novo precisa ser `/constitution`, e ele precisa povoar também módulos, fluxos, decisões, ADRs e segurança.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** textos do comando em três agentes, dicas de saída e dois itens de checagem; sem pacote novo (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Comando `/constitution` | `.agents/commands/constitution.md`, `.agents/skills/sdd-concepcao/references/constitution.md` e adaptadores | Idioma, nível, ferramentas, scan, preenche `constitution.md` e `stack.md` | estender | Passa a povoar `overview.md` (C4), `modules.md`, `flows.md`, `decisions.md`, uma ADR por decisão estrutural e a segurança, e a fechar com `brief --check`. |
| Dica pós-instalação | `cmd/forge-sdd/main.go` (`printPipelineHint`, `printCheatSheet`) | Fala do pipeline | estender | `/constitution` aparece como passo 1. |
| Menu | `cmd/forge-sdd/menu.go` | Pipeline ou estrutura padrão | estender | Mostra o passo 1 enquanto o contexto está em rascunho. |
| Checagem antes de abrir | `internal/pipeline/preflight.go` (`Preflight`, `LevelWarn`) | CLIs e arquivos de instrução | estender | Item `Constituição em rascunho` (aviso). |
| Diagnóstico | `cmd/forge-sdd/doctor.go` | Checagens de estrutura | estender | Item `Constituição: rascunho ou pronta`. |
| Estado do contexto | `internal/gate/brief.go` (75-01) | `Brief.Estado` | reusar | Única fonte da heurística (`DUP-01`). |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-01`, `COD-01`, `TST-02`, `TST-04`.
- Rules: `sdd-fluxo-e-portao.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-02`, `COD-02`, `TST-05`.
- Skill: `novo-subcomando-cli-sdd` (itens novos em comandos existentes).

## Escopo
1. `/constitution` (claude, gemini, openai e a cópia em `.agents/`): além de idioma, nível e ferramentas, escanear, preencher `constitution.md`, `stack.md`, `overview.md`, `modules.md`, `flows.md`, `decisions.md`, criar uma ADR por decisão estrutural e a seção de segurança (segredos, entradas, dependências); fechar com `forge-sdd brief --check` e só dar por concluído quando disser `pronta`.
2. `init` e menu: `/constitution` como passo 1; o passo do pipeline vem depois.
3. `Preflight` e `doctor`: Constituição em rascunho vira aviso (nunca bloqueio) com a ação `/constitution`, usando o estado do `brief`.
4. Goldens, cheat-sheet e orçamento `P-24` atualizados.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./internal/scaffold/... ./cmd/forge-sdd/...
```
- Teste de conteúdo: o comando `/constitution` nos três agentes cita módulos, fluxos, decisões, ADRs, segurança e `brief --check`.
- `TestInit_PrimeiroPasso`: a saída do `init` lista `/constitution` antes do pipeline; o menu idem em rascunho.
- `TestPreflight_ConstituicaoRascunho` e `TestDoctor_Constituicao`: aviso com a ação, sem bloquear; projeto `pronta` não avisa.
- `forge-sdd gate` sem falhas.

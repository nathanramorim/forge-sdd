# fix-118-02 — `forge doctor --harness` como hook de sessão falha com "Error:" vazio

**Branch da spec:** `fix/118-spec-guard-e-hook-de-sessao`; **da correção:** `fix/118-02-impl-doctor-harness-no-hook`, a partir da principal
**Origem:** Iniciativa 118, `hub-busize` (23/11/2026): ao abrir o Claude Code logo após atualizar apareceu `SessionStart:startup hook failed — Error:` sem nenhum texto.
**Tamanho:** `S`
**Status:** `spec`

## Sintoma

O `update` instala o hook `SessionStart` com `forge doctor --harness`. Quando alguma checagem dá `falha`, o comando sai com código 3, o relatório vai para **stdout** e o stderr recebe só `Error: ` (mensagem vazia de `exitCodeErr`) mais o bloco de uso do cobra. O Claude Code mostra o stderr: a pessoa vê "Error:" e nada do que corrigir. Reproduzido em projeto sem `sdd/.sddrc`: `forge doctor --harness` → código 3, stderr com 534 bytes de `Error: ` e `Usage:`.

## Feature Correlacionada

- **Arquivo de origem:** `cmd/forge-sdd/doctor.go` (`runDoctorHarness`) e `cmd/forge-sdd/brief.go` (`exitCodeErr`); `PersistentPreRunE` de `main.go` zera `SilenceUsage/SilenceErrors`.
- **Causa-raiz:** o relatório de falhas não é escrito em stderr e o cobra imprime erro vazio e uso para um código de saída "esperado".
- **Evidência:** execução acima; com projeto saudável, o mesmo comando sai 0 (`[ok]` em todas as linhas).

## Modelo do módulo

**Enxuto** (`ARQ-11`): ajuste de saída em `runDoctorHarness` e no tratamento de `exitCodeErr`.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-01` (`cmd/` só adapta), `SEG-05` (sem dado sensível na saída), `TST-02`, `TST-03`.

## Decisões do projeto aplicadas

- Em `--harness` com código ≠ 0, as linhas `[falha]` e `[aviso]` com a sugestão vão também para **stderr** (o canal que o hook mostra); o stdout continua igual (texto ou `--json`).
- `exitCodeErr` com `msg` vazia não imprime `Error:` nem o bloco de uso.
- Uma falha de `versao` (binário mais novo que o projeto) diz o comando que resolve: `forge update`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Silêncio de erro | `hook.go` (`SilenceUsage/SilenceErrors`) | só no `forge hook` | **reusar** o padrão para `exitCodeErr` |
| Relatório | `runDoctorHarness` | só stdout | **estender** |

## Comportamento esperado

1. `forge doctor --harness` com falha: stdout inalterado; stderr traz as linhas de falha e aviso com a sugestão; sem `Error:` vazio nem `Usage:`; código 3.
2. Projeto saudável: nada em stderr, código 0.
3. `--json`: o JSON segue só no stdout.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go test ./cmd/forge-sdd -run '^(TestDoctorHarnessFalhaEscreveEmStderr|TestDoctorHarnessSaudavelNaoEscreveEmStderr|TestExitCodeVazioNaoImprimeErroNemUso)$' -count=1
[ "$(go test ./cmd/forge-sdd -list '^(TestDoctorHarnessFalhaEscreveEmStderr|TestDoctorHarnessSaudavelNaoEscreveEmStderr|TestExitCodeVazioNaoImprimeErroNemUso)$' | grep -c '^Test')" -eq 3 ]
go test ./... -count=1
```

## Fora do escopo

- Descobrir qual checagem falhou no `hub-busize` na hora; o dono não guardou a saída, e hoje o relatório sai todo `[ok]`.

# feat-118-06b — `update --dry-run` completo: lista o plano inteiro, sem efeito

**Depende de:** feat-118-06a e das features 118-02, 118-03, 118-04, 118-05 (o plano lista o que elas fazem)
**Achado:** 8. **Tamanho:** `M`. **Status:** `spec`

## Objetivo

`forge update --dry-run` mostra tudo o que o `update` real faria, e nada é gravado (`ARQ-05`): hook `PreToolUse` (e os outros hooks e o MCP), arquivos que serão reescritos, arquivos preservados por edição do usuário (instruções, rules, script), campos de modo que seriam acrescentados na Constituição e o `forge-pronto.yml` quando `--ci`.

## Entradas

- `update --dry-run` (`cmd/forge-sdd/main.go:231-237`): chama `reportClaudeIntegration(..., dryRun)` e `printUpgradePlan` (`update_plan.go:13`), que só trata layout e obsoletos.
- `scaffold.Run`/`renderDir` com `cfg.DryRun` (`scaffold.go:186-190`), que já imprime `[DRY]` por destino sem consultar `shouldPreserve`.
- `SyncClaudeIntegration(dir, dryRun)` (`hooks.go:190`).

## Saídas

- Plano com seções: "Serão reescritos", "Serão preservados (editados por você)", "Serão acrescentados" (Constituição, `.sddrc`), "Hooks e MCP" e "Layout". Mesmo critério de decisão do `update` real (a **mesma** função, não uma cópia).
- Garantia testada: o `--dry-run` não altera nenhum arquivo (snapshot da árvore idêntico).

## Modelo do módulo

**Completo** (`ARQ-11`): `scaffold` (decisão por arquivo exposta como plano) e `cmd/forge-sdd` (impressão). Justificativa: o plano depende das decisões de várias features.

## Decisões do projeto aplicadas

- `ARQ-05` é o centro: nenhum efeito no disco.
- `DUP-06`: uma só função decide "reescrever, preservar ou acrescentar", usada pelo real e pelo dry-run; o `renderDir` com `DryRun` deixa de imprimir sem decidir.
- `TST-03`: conforme (plano completo) e negativo (árvore intocada).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Plano atual | `update_plan.go:13` `printUpgradePlan` | só layout e obsoletos | **estender** | achado 8 |
| Dry-run do scaffold | `scaffold.go:186-190` | lista tudo como `[DRY]`, ignora `shouldPreserve` | **estender** | honestidade do plano |
| Snapshot de árvore | `doctor_versao_test.go:99` `snapshotArquivos` | helper de teste | **reusar** | prova do "sem efeito" |
| Teste do dry-run | `update_plan_test.go:13,37` | cobre layout | **estender** | novos casos |
| Relato do hook | `agent_models.go:32` | já recebe `dryRun` | **reusar** | lista os hooks |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-05`, `DUP-06`, `DUP-03`, `DUP-02`, `ARQ-11`, `TST-02` (golden do texto do plano), `TST-03`.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**; os testes existentes `TestUpdateDryRun_ImprimeMovimentosENaoGravaNada` e `TestUpdateDryRun_SemLegadoInformaQueNadaHaAFazer` permanecem verdes.

```bash
set -o pipefail
go test ./cmd/forge-sdd -run '^(TestUpdateDryRunListaOHookPreToolUse|TestUpdateDryRunListaArquivosReescritos|TestUpdateDryRunListaArquivosPreservados|TestUpdateDryRunListaAcrescimosDaConstituicao|TestUpdateDryRunNaoAlteraNada)$' -count=1
go test ./internal/scaffold -run '^TestPlanoDoUpdateIgualAoRealSemGravar$' -count=1
go test ./cmd/forge-sdd -run '^(TestUpdateDryRun_ImprimeMovimentosENaoGravaNada|TestUpdateDryRun_SemLegadoInformaQueNadaHaAFazer)$' -count=1
[ "$(go test ./cmd/forge-sdd ./internal/scaffold -list '^(TestUpdateDryRunListaOHookPreToolUse|TestUpdateDryRunListaArquivosReescritos|TestUpdateDryRunListaArquivosPreservados|TestUpdateDryRunListaAcrescimosDaConstituicao|TestUpdateDryRunNaoAlteraNada|TestPlanoDoUpdateIgualAoRealSemGravar)$' | grep -c '^Test')" -eq 6 ]
go test ./cmd/forge-sdd ./internal/scaffold -count=1
```

## Fora do escopo

- JSON inválido, `--yes` e código de saída (118-06c).

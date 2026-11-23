# feat-118-06a — `doctor` sem aviso circular e confirmação do hook no `update`

**Depende de:** fix-106-14a (o `doctor` já confere o arquivo do script) e feat-118-05 (o aviso do `pronto?` depende do workflow opt-in)
**Achados:** 7 e 8 (parte do log real). **Tamanho:** `S`. **Status:** `spec`

## Objetivo

`forge doctor` só diz "ok" depois de conferir (RN-5): o job `pronto?` só é citado quando `.github/workflows/forge-pronto.yml` (ou o `ci.yml`) o contém; sem Claude, o aviso não manda rodar `forge update --upgrade` (que nunca instala o hook) e informa o que de fato vale. O `update` real **confirma** no log o hook `PreToolUse` instalado.

## Entradas

- `reportPRGuardCoverage` (`cmd/forge-sdd/doctor.go:363-398`): hoje procura o texto `pr_guard.py` no settings (linha 367), diz "rode 'forge update --upgrade'" (373) e cita `"pronto?"` sempre (396).
- `reportClaudeIntegration` (`cmd/forge-sdd/agent_models.go:32-47`) e `scaffold.SyncClaudeIntegration`.

## Saídas

- Mensagens do `doctor`: hook ausente em projeto **sem Claude** vira informação ("sem hook neste agente"), não pendência; com Claude, a correção é `forge update` (14b/14a), nunca o comando circular.
- Linha do job `pronto?` só quando o arquivo de workflow que o define existe.
- Log do `update` real lista "hook PreToolUse instalado" quando mesclou (e "já presente" quando não mudou).

## Modelo do módulo

**Enxuto** (`ARQ-11`): só `cmd/forge-sdd` (`doctor.go`, `agent_models.go`). Justificativa: ajuste de mensagens e de uma checagem, sem contrato novo.

## Decisões do projeto aplicadas

- `ARQ-05` e RN-5: o `doctor` lê, não grava.
- `DUP-06`: a checagem do arquivo do script é a do fix-106-14a; aqui só se reusa.
- `TST-03`: casos com e sem Claude, com e sem workflow.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Checagem do script | fix-106-14a (`os.Stat` em `doctor.go`) | spec pronta | **reusar** | evita duplicar |
| Cobertura do hook | `doctor.go:363-398` | aviso circular e texto fixo do `pronto?` | **estender** | achado 7 |
| Relato do `update` | `agent_models.go:32-47` (`changed`, `warns`) | já imprime `changed` | **reusar** | só garantir o hook na lista |
| Teste do doctor | `doctor_test.go`, `doctor_versao_test.go` | sem caso do aviso | **criar** casos | `TST-03` |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-05` (sem efeito colateral no `doctor` e no dry-run), `DUP-06`, `DUP-02`, `ARQ-11`, `TST-03`.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**.

```bash
set -o pipefail
go test ./cmd/forge-sdd -run '^(TestDoctorSemClaudeNaoMandaUpgrade|TestDoctorNaoCitaProntoSemWorkflow|TestDoctorCitaProntoComWorkflow|TestUpdateConfirmaOHookNoLog)$' -count=1
[ "$(go test ./cmd/forge-sdd -list '^(TestDoctorSemClaudeNaoMandaUpgrade|TestDoctorNaoCitaProntoSemWorkflow|TestDoctorCitaProntoComWorkflow|TestUpdateConfirmaOHookNoLog)$' | grep -c '^Test')" -eq 4 ]
go test ./cmd/forge-sdd -count=1
```

## Fora do escopo

- `--dry-run` completo (118-06b) e JSON inválido e `--yes` (118-06c).

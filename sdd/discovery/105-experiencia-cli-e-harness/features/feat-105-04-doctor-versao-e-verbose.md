# feat-105-04 — `doctor` compara `.sdd-version` com o binário; lista só com `--verbose`

**Iniciativa:** 105-experiencia-cli-e-harness (onda 1, task 1.4)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** nada
**Paralelo com:** feat-105-02, feat-105-03, feat-105-05
**Sequenciar com:** 106 (task 1.7/2.5/2.6: `doctor`, `hooks.go`, menus) e 104 (`doctor --harness`, `internal/doctor`): este PR altera `cmd/forge-sdd/doctor.go` e o pacote que gera o relatório; entra em série, rebase sobre o que a 106 e a 104 tiverem na `main`, resolvendo conflito sem editar arquivos de outra iniciativa. Se a 104 já tiver `internal/doctor`, estender lá (`ARQ-01`).

## Problema
`forge doctor` diz "em sincronia" comparando `.sddrc` com `.sdd-version`, não com o binário, e lista ~126 caminhos (O10).

## Comportamento
1. A checagem de versão compara a versão do projeto (`.sdd-version`/`.sddrc`) com a versão do binário em execução: igual = ok; binário mais antigo = falha com a sugestão de atualizar; projeto mais antigo = aviso com `forge update`.
2. A lista de caminhos/features só sai com `--verbose`; o padrão imprime resumo curto (contagens e pendências).
3. Sem efeito colateral (`ARQ-05`); mensagem em pt-BR; valores lidos de arquivo saneados (`SEG-01`).
4. Código de saída existente preservado.

## Modelo do módulo
**Estende** o diagnóstico existente; regra de comparação fica em `internal/` (função pura testável), `cmd/forge-sdd/doctor.go` só liga a flag `--verbose` (`ARQ-01`).

## Decisões do projeto aplicadas
ADR-0010 (comando único), ADR-0004; ADR-0015 (proposta da 104, não vale ainda); `/doctor` segue único.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Diagnóstico atual | `cmd/forge-sdd/doctor.go:90-96` | Compara `.sddrc` com `.sdd-version` | **Estender** | `DUP-01` |
| Comparação de versão | `internal/config` (`CompareVersions`) | Ignora sufixo de pré-lançamento (nota da feat-104-03) | **Reusar**; não alterar | `DUP-06` |
| Doctor da 104 | `internal/doctor` (só na branch da 104, `versaoMenor`) | Tratamento de `5.0.0-beta` | **Reusar se já na `main`**, senão função mínima local | Evitar duplicar (`DUP-01`) |
| Testes | `cmd/forge-sdd/doctor_test.go` | Testes do doctor | **Estender** | `TST-03` |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `SEG-01`, `TST-03`. Rules: `sdd-arquitetura`, `sdd-seguranca`, `sdd-testes`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go build ./... && go vet ./...
go test -list 'TestDoctorVersaoComparaComBinario' ./cmd/forge-sdd | grep -F 'TestDoctorVersaoComparaComBinario'
go test ./cmd/forge-sdd -run 'TestDoctorVersaoComparaComBinario'
go test -list 'TestDoctorListaSoComVerbose' ./cmd/forge-sdd | grep -F 'TestDoctorListaSoComVerbose'
go test ./cmd/forge-sdd -run 'TestDoctorListaSoComVerbose'
# negativo: projeto igual ao binário NÃO diz "em sincronia" quando o binário é mais antigo que o projeto
go test -list 'TestDoctorBinarioMaisAntigoFalha' ./cmd/forge-sdd | grep -F 'TestDoctorBinarioMaisAntigoFalha'
go test ./cmd/forge-sdd -run 'TestDoctorBinarioMaisAntigoFalha'
# negativo: versão com ESC/ANSI/CR no `.sdd-version` não chega à saída
go test -list 'TestDoctorVersaoSaneada' ./cmd/forge-sdd | grep -F 'TestDoctorVersaoSaneada'
go test ./cmd/forge-sdd -run 'TestDoctorVersaoSaneada'
go test ./...
```
- Sem `--verbose` a saída não contém a lista de caminhos (teste de contagem de linhas com limite explícito, não `grep -c`).
- `doctor` não escreve no disco (`git status --porcelain` igual antes e depois, coberto em teste).

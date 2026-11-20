# ref-07 — remover-duplicados

## Comportamento
`RemoverDuplicados(xs []string) []string` mantém a primeira ocorrência e a ordem.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestRemoverDuplicadosNegativo' ./texto | grep -F 'TestRemoverDuplicadosNegativo' >/dev/null
# negativo: ordem alterada na saída reprova o teste
go test ./texto -run 'TestRemoverDuplicadosNegativo'
```

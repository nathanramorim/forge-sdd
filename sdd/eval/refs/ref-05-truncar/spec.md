# ref-05 — truncar

## Comportamento
`Truncar(s string, n int) string` corta em n runas e acrescenta reticências.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestTruncarNegativo' ./texto | grep -F 'TestTruncarNegativo' >/dev/null
# negativo: corte no meio de uma runa reprova o teste
go test ./texto -run 'TestTruncarNegativo'
```

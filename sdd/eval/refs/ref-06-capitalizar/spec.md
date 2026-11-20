# ref-06 — capitalizar

## Comportamento
`Capitalizar(s string) string` deixa a primeira letra de cada palavra em maiúscula.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestCapitalizarNegativo' ./texto | grep -F 'TestCapitalizarNegativo' >/dev/null
# negativo: palavra iniciada por acento sem maiúscula reprova o teste
go test ./texto -run 'TestCapitalizarNegativo'
```

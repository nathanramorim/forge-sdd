# ref-03 — contar-palavras

## Comportamento
`ContarPalavras(s string) int` conta palavras separadas por qualquer espaço em branco.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestContarPalavrasNegativo' ./texto | grep -F 'TestContarPalavrasNegativo' >/dev/null
# negativo: espaços duplos contados como palavra vazia reprovam o teste
go test ./texto -run 'TestContarPalavrasNegativo'
```

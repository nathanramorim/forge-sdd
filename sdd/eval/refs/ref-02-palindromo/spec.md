# ref-02 — palindromo

## Comportamento
`Palindromo(s string) bool` ignora caixa e espaços.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestPalindromoNegativo' ./texto | grep -F 'TestPalindromoNegativo' >/dev/null
# negativo: "Ana" com caixa mista tratado como não palíndromo reprova o teste
go test ./texto -run 'TestPalindromoNegativo'
```

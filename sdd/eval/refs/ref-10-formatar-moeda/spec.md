# ref-10 — formatar-moeda

## Comportamento
`FormatarMoeda(centavos int64) string` formata em reais com separador de milhar.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestFormatarMoedaNegativo' ./texto | grep -F 'TestFormatarMoedaNegativo' >/dev/null
# negativo: valor negativo sem sinal reprova o teste
go test ./texto -run 'TestFormatarMoedaNegativo'
```

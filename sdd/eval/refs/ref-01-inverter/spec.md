# ref-01 — inverter

## Comportamento
`Inverter(s string) string` devolve a string de trás para frente, respeitando runas UTF-8.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestInverterNegativo' ./texto | grep -F 'TestInverterNegativo' >/dev/null
# negativo: "ação" invertida byte a byte reprova o teste
go test ./texto -run 'TestInverterNegativo'
```

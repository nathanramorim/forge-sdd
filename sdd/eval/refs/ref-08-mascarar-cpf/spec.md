# ref-08 — mascarar-cpf

## Comportamento
`MascararCPF(cpf string) string` mostra só os dois últimos dígitos.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestMascararCPFNegativo' ./texto | grep -F 'TestMascararCPFNegativo' >/dev/null
# negativo: CPF com menos de 11 dígitos aceito sem erro reprova o teste
go test ./texto -run 'TestMascararCPFNegativo'
```

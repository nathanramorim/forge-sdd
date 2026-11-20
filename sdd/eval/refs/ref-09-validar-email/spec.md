# ref-09 — validar-email

## Comportamento
`ValidarEmail(s string) bool` exige usuário, um único @ e domínio com ponto.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestValidarEmailNegativo' ./texto | grep -F 'TestValidarEmailNegativo' >/dev/null
# negativo: endereço com dois @ aceito reprova o teste
go test ./texto -run 'TestValidarEmailNegativo'
```

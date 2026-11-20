# ref-04 — slug

## Comportamento
`Slug(s string) string` gera minúsculas, remove acentos e troca espaços por hífen.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test -list 'TestSlugNegativo' ./texto | grep -F 'TestSlugNegativo' >/dev/null
# negativo: acento mantido no slug reprova o teste
go test ./texto -run 'TestSlugNegativo'
```

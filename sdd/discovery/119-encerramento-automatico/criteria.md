# Critérios — 119

```bash
set -o pipefail
go build ./... && go vet ./...
go test ./... -count=1
```

Detalhe em `fixes/fix-119-01-encerramento-de-entrega-concluida.md`.

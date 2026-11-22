#!/usr/bin/env bash
# CI local primeiro (fix-106-10): mesmos passos do .github/workflows/ci.yml.
# Passou tudo => grava .forge/runtime/ci-local.json (head_sha + hora). Falha nao grava selo.
# SEG-01: sem rede e sem escrita fora do repositorio.
# Uso: scripts/ci-local.sh [tudo|test]   (test = sem lint e sem selo; o job lint do GitHub usa a action).
set -euo pipefail
modo="${1:-tudo}"
raiz="$(git rev-parse --show-toplevel)"
cd "$raiz"
rm -f .forge/runtime/ci-local.json

go build ./...
go vet ./...
go test ./... -count=1
if [ "$modo" = "tudo" ] && command -v golangci-lint >/dev/null 2>&1; then
  golangci-lint run
elif [ "$modo" = "tudo" ]; then
  echo "aviso: golangci-lint nao instalado; lint pulado (o CI do GitHub roda)" >&2
fi
GOOS=windows go build ./... && GOOS=windows go vet ./...
GOOS=darwin go build ./... && GOOS=darwin go vet ./...

[ "$modo" = "tudo" ] || exit 0
mkdir -p .forge/runtime
head_sha="$(git rev-parse HEAD)"
printf '{"head_sha": "%s", "gerado_em": "%s"}\n' "$head_sha" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > .forge/runtime/ci-local.json
echo "ci-local: ok; selo gravado para $head_sha"

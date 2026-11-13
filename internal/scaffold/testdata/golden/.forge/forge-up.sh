#!/usr/bin/env bash
# Abre as sessões forge-discovery, forge-builder e forge-reviewer do pipeline Forge SDD.
# Uso: .forge/forge-up.sh [--cli claude|agy] [--dry-run]
# A lógica fica em `forge-sdd up`; este script só garante o diretório e o binário.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
if ! command -v forge-sdd >/dev/null 2>&1; then
  echo "forge-sdd não está no PATH. Instale com: npm install -g @nathanramorim/forge-sdd@beta" >&2
  exit 1
fi
exec forge-sdd up "$@"

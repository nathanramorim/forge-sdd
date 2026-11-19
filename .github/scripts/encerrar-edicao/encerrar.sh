#!/usr/bin/env bash
set -euo pipefail

repo="${1:-.}"
if [ ! -d "$repo" ]; then
  echo "erro: pasta do repositório inválida: $repo" >&2
  exit 1
fi

repo="$(cd "$repo" && pwd)"
script_dir="$(cd "$(dirname "$0")" && pwd)"

remover_marcados() {
  local arquivo="$1"
  [ -f "$arquivo" ] || return 0
  if ! grep -q 'edicao:inicio' "$arquivo"; then
    return 0
  fi
  local tmp
  tmp="$(mktemp)"
  awk '
    /edicao:inicio/ { skip = 1; next }
    /edicao:fim/ { skip = 0; next }
    !skip { print }
  ' "$arquivo" > "$tmp"
  mv "$tmp" "$arquivo"
}

trocar_cabecalho() {
  local readme="$1"
  local modelo="$2"
  [ -f "$readme" ] || return 0
  if ! grep -q 'edicao:inicio' "$readme"; then
    return 0
  fi
  if [ ! -f "$modelo" ]; then
    echo "erro: modelo de cabeçalho ausente: $modelo" >&2
    exit 1
  fi
  local tmp
  tmp="$(mktemp)"
  awk -v modelo="$modelo" '
    /edicao:inicio/ {
      while ((getline linha < modelo) > 0) print linha
      close(modelo)
      skip = 1
      next
    }
    /edicao:fim/ { skip = 0; next }
    !skip { print }
  ' "$readme" > "$tmp"
  mv "$tmp" "$readme"
}

remover_marcados "$repo/docs/index.html"
remover_marcados "$repo/docs/wiki.html"
trocar_cabecalho "$repo/README.md" "$script_dir/README.cabecalho.md"
trocar_cabecalho "$repo/npm/README.md" "$script_dir/npm.README.cabecalho.md"

rm -rf "$repo/docs/edicao" "$repo/.github/scripts/encerrar-edicao"

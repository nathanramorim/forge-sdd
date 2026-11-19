#!/bin/bash
# uso: render.sh <svg> <png> <largura> <altura> [escala]
# Só no macOS, com o Google Chrome instalado. Gera PNG com fundo transparente.
set -euo pipefail
svg="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
png="$2"; w="$3"; h="$4"; s="${5:-1}"
perfil="$(mktemp -d)"
perl -e 'alarm 40; exec @ARGV' "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
  --headless --disable-gpu --hide-scrollbars --user-data-dir="$perfil" \
  --force-device-scale-factor="$s" --default-background-color=00000000 \
  --screenshot="$png" --window-size="$w,$h" "file://$svg" >/dev/null 2>&1 || true
rm -rf "$perfil"
test -s "$png" && ls -la "$png" | awk '{print $5, $9}'

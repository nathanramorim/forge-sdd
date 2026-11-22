#!/usr/bin/env bash
# Deploy da LP (fix-106-11): um lugar so (DUP-01), usado na maquina e pelo workflow manual.
# Uso: scripts/lp-deploy.sh [--dry-run | --status]
#   (sem flag)  regera as novidades, confere com --check e publica com o Wrangler a partir de site/
#   --dry-run   imprime os comandos; nao chama a rede nem o wrangler
#   --status    compara o ultimo commit que tocou docs/ com o ultimo deploy registrado
# SEG-04: o token vem do ambiente (CLOUDFLARE_API_TOKEN) e nunca e impresso.
# Sem escrita fora do repositorio: o registro vai para .forge/runtime/lp-deploy.json.
set -euo pipefail
modo="${1:-deploy}"
raiz="$(git rev-parse --show-toplevel)"
cd "$raiz"
registro=".forge/runtime/lp-deploy.json"
wrangler="wrangler@4.165.0"

case "$modo" in
  --status)
    ultimo="$(git log -1 --format=%H -- docs 2>/dev/null || true)"
    feito=""
    if [ -f "$registro" ]; then
      feito="$(sed -n 's/.*"docs_commit"[^"]*"\([0-9a-f]*\)".*/\1/p' "$registro" | head -n1)"
    fi
    if [ -z "$ultimo" ]; then echo "LP: nenhum commit em docs/"; exit 0; fi
    if [ "$ultimo" = "$feito" ]; then
      echo "LP em dia (docs/ em $ultimo)"
    else
      echo "LP desatualizada: docs/ mudou em $ultimo e o ultimo deploy registrado foi ${feito:-nenhum}. Rode scripts/lp-deploy.sh"
    fi
    exit 0
    ;;
  --dry-run)
    echo "python3 scripts/novidades_lp.py"
    echo "python3 scripts/novidades_lp.py --check"
    echo "(cd site && npx $wrangler deploy)"
    echo "dry-run: nada executado, sem rede"
    exit 0
    ;;
  deploy) ;;
  *) echo "uso: $0 [--dry-run|--status]" >&2; exit 2 ;;
esac

python3 scripts/novidades_lp.py
python3 scripts/novidades_lp.py --check || { echo "recusado: releases.json diverge do historico" >&2; exit 1; }
(cd site && npx "$wrangler" deploy)

mkdir -p .forge/runtime
printf '{"docs_commit": "%s", "head_sha": "%s", "gerado_em": "%s"}\n' \
  "$(git log -1 --format=%H -- docs)" "$(git rev-parse HEAD)" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$registro"
echo "LP publicada: https://forge-sdd.org (confira no ar o video e as imagens que a pagina carrega)"

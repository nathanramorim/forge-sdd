#!/usr/bin/env bash
# Deploy da LP (fix-106-11): um lugar so (DUP-01), usado na maquina e pelo workflow manual.
# Uso: scripts/lp-deploy.sh [--dry-run | --status] [--conta <id>]
#   (sem flag)  regera as novidades, confere com --check e publica com o Wrangler a partir de site/
#   --dry-run   imprime os comandos; nao chama a rede nem o wrangler
#   --status    compara o ultimo commit que tocou docs/ com o ultimo deploy registrado
# Conta do Cloudflare: sempre explicita, nunca a do ambiente (um shell com outro XDG_CONFIG_HOME ou
#   CLOUDFLARE_ACCOUNT_ID publicaria a LP na conta errada). Ordem: --conta, LP_CLOUDFLARE_ACCOUNT_ID,
#   .forge/config/lp-deploy.json ({"account_id": "..."}, fora do git). Sem conta, o deploy recusa.
# SEG-04: o token vem do ambiente (CLOUDFLARE_API_TOKEN) e nunca e impresso.
# Sem escrita fora do repositorio: o registro vai para .forge/runtime/lp-deploy.json.
set -euo pipefail
modo="deploy"
conta_arg=""
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run|--status) modo="$1"; shift ;;
    --conta) [ $# -ge 2 ] || { echo "uso: --conta <id>" >&2; exit 2; }; conta_arg="$2"; shift 2 ;;
    *) echo "uso: $0 [--dry-run|--status] [--conta <id>]" >&2; exit 2 ;;
  esac
done
raiz="$(git rev-parse --show-toplevel)"
cd "$raiz"
registro=".forge/runtime/lp-deploy.json"
wrangler="wrangler@4.165.0"
arquivo_conta=".forge/config/lp-deploy.json"

# resolve_conta imprime o id da conta (32 hex) ou nada; o formato e validado por quem chama.
resolve_conta() {
  if [ -n "$conta_arg" ]; then printf '%s' "$conta_arg"; return; fi
  if [ -n "${LP_CLOUDFLARE_ACCOUNT_ID:-}" ]; then printf '%s' "$LP_CLOUDFLARE_ACCOUNT_ID"; return; fi
  if [ -f "$arquivo_conta" ]; then tr -d '\n' < "$arquivo_conta" | sed -n 's/.*"account_id"[^"]*"\([^"]*\)".*/\1/p' | head -n1; fi
}
mascara() { printf '****%s' "${1: -4}"; }

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
    conta="$(resolve_conta)"
    if [[ "$conta" =~ ^[0-9a-f]{32}$ ]]; then
      echo "conta do Cloudflare: $(mascara "$conta") (CLOUDFLARE_ACCOUNT_ID so neste comando)"
    else
      echo "conta do Cloudflare: nao definida; o deploy recusa (use --conta, LP_CLOUDFLARE_ACCOUNT_ID ou $arquivo_conta)"
    fi
    echo "(cd site && npx $wrangler deploy)"
    echo "dry-run: nada executado, sem rede"
    exit 0
    ;;
  deploy) ;;
  *) echo "uso: $0 [--dry-run|--status]" >&2; exit 2 ;;
esac

python3 scripts/novidades_lp.py
python3 scripts/novidades_lp.py --check || { echo "recusado: releases.json diverge do historico" >&2; exit 1; }
conta="$(resolve_conta)"
if ! [[ "$conta" =~ ^[0-9a-f]{32}$ ]]; then
  echo "recusado: conta do Cloudflare nao definida ou invalida. Use --conta <id>, LP_CLOUDFLARE_ACCOUNT_ID ou $arquivo_conta ({\"account_id\": \"...\"}, fora do git)." >&2
  exit 1
fi
ident="$(cd site && npx "$wrangler" whoami 2>/dev/null)" || ident=""
if ! grep -qF -- "$conta" <<<"$ident"; then
  echo "recusado: a conta $(mascara "$conta") nao esta entre as contas do login atual do Wrangler (confira XDG_CONFIG_HOME e 'wrangler whoami')." >&2
  exit 1
fi
(cd site && CLOUDFLARE_ACCOUNT_ID="$conta" npx "$wrangler" deploy)

mkdir -p .forge/runtime
printf '{"docs_commit": "%s", "head_sha": "%s", "gerado_em": "%s"}\n' \
  "$(git log -1 --format=%H -- docs)" "$(git rev-parse HEAD)" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$registro"
echo "LP publicada: https://forge-sdd.org (confira no ar o video e as imagens que a pagina carrega)"

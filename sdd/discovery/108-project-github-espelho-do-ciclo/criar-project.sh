#!/usr/bin/env bash
# RASCUNHO da iniciativa 108. NÃO foi executado por quem o escreveu.
# Cria o Project de usuário (privado), o campo "Etapa" com as 7 opções, as issues de issues/*.md
# (sem duplicar, procurando pelo título) e os cartões, definindo a etapa pelo indice.csv.
#
# Uso:
#   criar-project.sh [--dry-run]    (padrão) só imprime o que faria; não chama o gh
#   criar-project.sh --executar     cria de verdade (exige gh autenticado com o escopo "project")
#
# Por que um campo novo "Etapa" e não o "Status" padrão: o gh não tem "field-edit" (gh project --help);
# editar as opções do "Status" exigiria a API GraphQL (a verificar). field-create SINGLE_SELECT existe.
# Nunca imprime token. Não usa "closes", "fixes" nem "resolves" em nenhum texto.
set -euo pipefail

REPO="nathanramorim/forge-sdd"
DONO="@me"
TITULO_PROJETO="Forge SDD — ciclo"
CAMPO="Etapa"
OPCOES="Backlog,Spec,Spec-validation,Todo,Building,Validation,Done"
AQUI="$(cd "$(dirname "$0")" && pwd)"
INDICE="$AQUI/issues/indice.csv"

MODO="dry-run"
for arg in "$@"; do
  case "$arg" in
    --dry-run) MODO="dry-run" ;;
    --executar) MODO="executar" ;;
    *) echo "argumento desconhecido: $arg (use --dry-run ou --executar)" >&2; exit 2 ;;
  esac
done

[ -f "$INDICE" ] || { echo "falta $INDICE" >&2; exit 1; }

# A verificar antes de executar: flags --format/-q de "gh project view" e "item-add", e o formato de
# "field-list --format json" (.fields[].options[]). Sem jq: toda leitura de JSON usa o -q do próprio gh.

# Lista fechada das etapas: qualquer valor fora dela é recusado antes de chamar o gh.
etapa_valida() {
  case ",$OPCOES," in *",$1,"*) return 0 ;; *) return 1 ;; esac
}

if [ "$MODO" = "dry-run" ]; then
  echo "DRY-RUN: nenhuma chamada ao gh será feita."
  echo "DRY-RUN: criaria o Project '$TITULO_PROJETO' (dono $DONO), depois --visibility PRIVATE e link com $REPO"
  echo "DRY-RUN: criaria o campo $CAMPO (SINGLE_SELECT) com as opções: $OPCOES"
  # Linha 1 é o cabeçalho; os títulos não têm vírgula (verificado ao gerar o índice).
  tail -n +2 "$INDICE" | while IFS=, read -r chave titulo coluna tipo pasta pr pai; do
    etapa_valida "$coluna" || { echo "etapa inválida em $chave: $coluna" >&2; exit 1; }
    [ -f "$AQUI/issues/$chave.md" ] || { echo "falta issues/$chave.md" >&2; exit 1; }
    echo "DRY-RUN: issue '$titulo' (se não existir) -> etapa $coluna${pr:+; PR $pr}$([ "$coluna" = Validation ] && [ -n "$pr" ] && echo " vira cartão próprio em Validation")"
  done
  echo "DRY-RUN: fim. Para criar de verdade: $0 --executar"
  exit 0
fi

# ---- modo executar ----
command -v gh >/dev/null 2>&1 || { echo "gh não encontrado" >&2; exit 1; }
gh auth status >/dev/null 2>&1 || { echo "gh não autenticado (rode gh auth login)" >&2; exit 1; }
gh auth status 2>&1 | grep -q "project" || { echo "o token do gh não tem o escopo project (gh auth refresh -s project)" >&2; exit 1; }

# (a) Project privado, sem duplicar: procura pelo título.
numero="$(gh project list --owner "$DONO" --format json -q ".projects[] | select(.title==\"$TITULO_PROJETO\") | .number" | head -n1 || true)"
if [ -z "$numero" ]; then
  numero="$(gh project create --owner "$DONO" --title "$TITULO_PROJETO" --format json -q .number)"
  echo "Project criado: número $numero"
else
  echo "Project já existe: número $numero"
fi
gh project edit "$numero" --owner "$DONO" --visibility PRIVATE >/dev/null
gh project link "$numero" --owner "$DONO" --repo "$REPO" >/dev/null 2>&1 || echo "aviso: link com $REPO não aplicado (talvez já exista)"

project_id="$(gh project view "$numero" --owner "$DONO" --format json -q .id)"

# (b) Campo "Etapa" com as 7 opções na ordem do pedido, se ainda não existir.
campo_id="$(gh project field-list "$numero" --owner "$DONO" --format json -q ".fields[] | select(.name==\"$CAMPO\") | .id" || true)"
if [ -z "$campo_id" ]; then
  gh project field-create "$numero" --owner "$DONO" --name "$CAMPO" --data-type SINGLE_SELECT --single-select-options "$OPCOES" >/dev/null
  campo_id="$(gh project field-list "$numero" --owner "$DONO" --format json -q ".fields[] | select(.name==\"$CAMPO\") | .id")"
  echo "Campo $CAMPO criado"
else
  echo "Campo $CAMPO já existe"
fi

opcao_id() {
  gh project field-list "$numero" --owner "$DONO" --format json \
    -q ".fields[] | select(.name==\"$CAMPO\") | .options[] | select(.name==\"$1\") | .id"
}

tail -n +2 "$INDICE" | while IFS=, read -r chave titulo coluna tipo pasta pr pai; do
  etapa_valida "$coluna" || { echo "etapa inválida em $chave: $coluna" >&2; exit 1; }
  corpo="$AQUI/issues/$chave.md"
  [ -f "$corpo" ] || { echo "falta $corpo" >&2; exit 1; }

  # (c) Cria a issue só se o título ainda não existir.
  numero_issue="$(gh issue list --repo "$REPO" --state all --limit 500 --json number,title -q ".[] | select(.title==\"$titulo\") | .number" | head -n1 || true)"
  if [ -z "$numero_issue" ]; then
    url="$(gh issue create --repo "$REPO" --title "$titulo" --body-file "$corpo")"
    echo "issue criada: $titulo -> $url"
  else
    url="https://github.com/$REPO/issues/$numero_issue"
    echo "issue já existe: $titulo -> $url"
  fi

  # (d) Adiciona a issue e o PR ao Project; (e) define a Etapa de cada cartão.
  alvos="$url"
  # Só PR de código (etapa Validation) vira cartão próprio; PR de spec (#120 a #122) é compartilhado e fica só citado no corpo da issue.
  if [ -n "$pr" ] && [ "$coluna" = "Validation" ]; then alvos="$alvos https://github.com/$REPO/pull/${pr#\#}"; fi
  opcao="$(opcao_id "$coluna")"
  for alvo in $alvos; do
    item="$(gh project item-add "$numero" --owner "$DONO" --url "$alvo" --format json -q .id)"
    gh project item-edit --id "$item" --project-id "$project_id" --field-id "$campo_id" --single-select-option-id "$opcao" >/dev/null
    echo "cartão: $alvo -> $coluna"
  done
done

# (f) Relatório final (contagens vêm do próprio Project; nenhum token é impresso).
echo "Relatório: Project $numero ($TITULO_PROJETO), campo $CAMPO com 7 opções."
gh project item-list "$numero" --owner "$DONO" --format json -q ".items | length" | sed 's/^/cartões no Project: /'

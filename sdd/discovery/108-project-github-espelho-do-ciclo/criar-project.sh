#!/usr/bin/env bash
# Iniciativa 108. Cria (ou completa) o Project de usuário "Forge SDD — ciclo", sem duplicar nada.
#
# Uso:
#   criar-project.sh [--dry-run]    (padrão) só imprime o que faria; não chama o gh
#   criar-project.sh --executar     cria de verdade (gh autenticado com o escopo "project")
#
# O campo da etapa é o "Status" do Project. Verificado na API real em 20/11/2026: a mutation GraphQL
# updateProjectV2Field SUBSTITUI as opções do Status; createProjectV2View com BOARD_LAYOUT cria o quadro,
# que agrupa pelo Status por padrão (o agrupamento não é configurável pela API). O gh não tem "field-edit".
# ATENÇÃO: a API gera ids NOVOS para as opções ao substituí-las; por isso, depois de atualizar, o laço
# final reatribui o Status de todos os cartões pelo indice.csv (idempotente).
# Nunca imprime token. Não usa "closes", "fixes" nem "resolves" em nenhum texto.
set -euo pipefail

REPO="nathanramorim/forge-sdd"
DONO="nathanramorim"          # "@me" falha no gh project link; usar o login
TITULO_PROJETO="Forge SDD — ciclo"
CAMPO="Status"
VISAO="Quadro do ciclo"
# nome:COR na ordem do pedido (cores aceitas pela API: GRAY BLUE GREEN YELLOW ORANGE RED PINK PURPLE)
OPCOES_COR="Backlog:GRAY Spec:BLUE Spec-validation:PURPLE Todo:YELLOW Building:ORANGE Validation:PINK Done:GREEN"
CORES_OK=" GRAY BLUE GREEN YELLOW ORANGE RED PINK PURPLE "
# Etiquetas: lista fechada nome|cor|descrição (constantes; nada interpolado de fora).
ETIQUETAS=(
  "código|4f8ef7|Muda código (Go, scripts, testes, templates)"
  "adr|9b7bff|Registro de decisão de arquitetura"
  "rule|f5b04a|Mexe em regra: Constituição ou .agents/rules"
  "documentação|99a2b8|Muda documentação ou texto entregue"
  "spec|4fd1b5|Ainda é só especificação, sem construção"
)
NOMES_ETIQUETA=",código,adr,rule,documentação,spec,"
CAMPOS_VISIVEIS=("Title" "Labels" "Linked pull requests")
for e in "${ETIQUETAS[@]}"; do
  IFS='|' read -r en ec ed <<< "$e"
  [[ "$ec" =~ ^[0-9a-f]{6}$ ]] || { echo "cor de etiqueta inválida: $ec" >&2; exit 1; }
  [ -n "$ed" ] && [ "${#ed}" -le 100 ] || { echo "descrição de etiqueta inválida: $en" >&2; exit 1; }
done
# Etiquetas do índice (separadas por +) precisam estar na lista fechada.
etiquetas_validas() {
  [ -z "$1" ] && return 0
  local x
  for x in ${1//+/ }; do
    case "$NOMES_ETIQUETA" in *",$x,"*) ;; *) return 1 ;; esac
  done
}
# Prefixo do título (função pura, mesma regra de PrefixoDoTitulo): conteúdo = código, adr, rule presentes
# (nessa ordem) unidos por +; sem nenhum, "doc" se houver documentação; com spec, começa por "spec · ".
# Etiqueta fora da lista fechada é recusada (return 1). Sem etiqueta, sem prefixo.
prefixo_do_titulo() {
  local et="$1" x tem_spec=0 tem_doc=0 tem_cod=0 tem_adr=0 tem_rule=0 conteudo="" p=""
  [ -z "$et" ] && return 0
  for x in ${et//+/ }; do
    case "$x" in
      código) tem_cod=1 ;; adr) tem_adr=1 ;; rule) tem_rule=1 ;;
      documentação) tem_doc=1 ;; spec) tem_spec=1 ;;
      *) echo "etiqueta fora da lista fechada: $x" >&2; return 1 ;;
    esac
  done
  [ "$tem_cod" = 1 ] && conteudo="código"
  [ "$tem_adr" = 1 ] && conteudo="${conteudo:+$conteudo+}adr"
  [ "$tem_rule" = 1 ] && conteudo="${conteudo:+$conteudo+}rule"
  [ -z "$conteudo" ] && [ "$tem_doc" = 1 ] && conteudo="doc"
  if [ "$tem_spec" = 1 ]; then p="spec"; [ -n "$conteudo" ] && p="spec · $conteudo"; else p="$conteudo"; fi
  [ -n "$p" ] && printf '[%s] ' "$p"
  return 0
}
AQUI="$(cd "$(dirname "$0")" && pwd)"
INDICE="$AQUI/issues/indice.csv"

OPCOES=""
for par in $OPCOES_COR; do
  nome="${par%%:*}"; cor="${par##*:}"
  [[ "$nome" =~ ^[A-Za-z-]+$ ]] || { echo "nome de etapa inválido: $nome" >&2; exit 1; }
  case "$CORES_OK" in *" $cor "*) ;; *) echo "cor inválida: $cor" >&2; exit 1 ;; esac
  OPCOES="${OPCOES:+$OPCOES,}$nome"
done

MODO="dry-run"
for arg in "$@"; do
  case "$arg" in
    --dry-run) MODO="dry-run" ;;
    --executar) MODO="executar" ;;
    *) echo "argumento desconhecido: $arg (use --dry-run ou --executar)" >&2; exit 2 ;;
  esac
done

[ -f "$INDICE" ] || { echo "falta $INDICE" >&2; exit 1; }

# Lista fechada das etapas: qualquer valor fora dela é recusado antes de chamar o gh.
etapa_valida() {
  case ",$OPCOES," in *",$1,"*) return 0 ;; *) return 1 ;; esac
}

if [ "$MODO" = "dry-run" ]; then
  echo "DRY-RUN: nenhuma chamada ao gh será feita."
  echo "DRY-RUN: criaria o Project '$TITULO_PROJETO' (dono $DONO) se não existir, privado e ligado a $REPO"
  echo "DRY-RUN: garantiria as opções do $CAMPO via updateProjectV2Field (só se diferirem): $OPCOES_COR"
  echo "DRY-RUN: garantiria o Board '$VISAO' via createProjectV2View (BOARD_LAYOUT), sem duplicar"
  echo "DRY-RUN: mostraria no Board os campos: ${CAMPOS_VISIVEIS[*]} (updateProjectV2View, visibleFieldIds buscados pelo nome)"
  for e in "${ETIQUETAS[@]}"; do
    IFS='|' read -r en ec ed <<< "$e"
    echo "DRY-RUN: gh label create --force '$en' --color $ec --description '$ed'"
  done
  # Linha 1 é o cabeçalho; os títulos não têm vírgula (verificado ao gerar o índice).
  tail -n +2 "$INDICE" | while IFS=, read -r chave titulo coluna tipo pasta pr pai etiq; do
    etiq="${etiq%$'\r'}"
    etapa_valida "$coluna" || { echo "etapa inválida em $chave: $coluna" >&2; exit 1; }
    [ -f "$AQUI/issues/$chave.md" ] || { echo "falta issues/$chave.md" >&2; exit 1; }
    etiquetas_validas "$etiq" || { echo "etiqueta inválida em $chave: $etiq" >&2; exit 1; }
    pref="$(prefixo_do_titulo "$etiq")" || exit 1
    echo "DRY-RUN: issue '${pref}${titulo}' (procura por título base ou prefixado; renomeia se achar; cria se não) -> $CAMPO $coluna${pr:+; PR $pr}$([ "$coluna" = Validation ] && [ -n "$pr" ] && echo " vira cartão próprio em Validation")${etiq:+; etiquetas: ${etiq//+/, }}"
  done
  echo "DRY-RUN: fim. Para criar de verdade: $0 --executar"
  exit 0
fi

# ---- modo executar ----
command -v gh >/dev/null 2>&1 || { echo "gh não encontrado" >&2; exit 1; }
gh auth status >/dev/null 2>&1 || { echo "gh não autenticado (rode gh auth login)" >&2; exit 1; }
gh auth status 2>&1 | grep -q "project" || { echo "o token do gh não tem o escopo project (gh auth refresh -s project)" >&2; exit 1; }

# (0) Etiquetas, idempotente (--force atualiza se já existir). Valores vêm da lista fechada acima.
for e in "${ETIQUETAS[@]}"; do
  IFS='|' read -r en ec ed <<< "$e"
  gh label create "$en" --repo "$REPO" --color "$ec" --description "$ed" --force >/dev/null
  echo "etiqueta garantida: $en"
done

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
campo_id="$(gh project field-list "$numero" --owner "$DONO" --format json -q ".fields[] | select(.name==\"$CAMPO\") | .id")"
[ -n "$campo_id" ] || { echo "campo $CAMPO não encontrado" >&2; exit 1; }

# (b) Garante as 7 opções do Status (só atualiza se os nomes ou a ordem diferirem).
atuais="$(gh project field-list "$numero" --owner "$DONO" --format json -q "[.fields[] | select(.name==\"$CAMPO\") | .options[].name] | join(\",\")")"
if [ "$atuais" != "$OPCOES" ]; then
  # Mutation 1: updateProjectV2Field. JSON montado só com nomes e cores validados acima (sem texto livre).
  lista=""
  for par in $OPCOES_COR; do
    nome="${par%%:*}"; cor="${par##*:}"
    lista="${lista:+$lista,}{\"name\":\"$nome\",\"color\":\"$cor\",\"description\":\"Etapa $nome do ciclo\"}"
  done
  corpo="$(mktemp)"
  printf '{"query":"mutation($input: UpdateProjectV2FieldInput!){updateProjectV2Field(input:$input){projectV2Field{... on ProjectV2SingleSelectField{id}}}}","variables":{"input":{"fieldId":"%s","singleSelectOptions":[%s]}}}' "$campo_id" "$lista" > "$corpo"
  gh api graphql --input "$corpo" >/dev/null
  rm -f "$corpo"
  echo "Opções do $CAMPO atualizadas (ids novos; os cartões serão reatribuídos abaixo)"
else
  echo "Opções do $CAMPO já corretas"
fi

# (c) Garante o Board "Quadro do ciclo" (lista as views antes; não duplica).
visoes="$(gh api graphql -f query='query($id:ID!){node(id:$id){... on ProjectV2{views(first:50){nodes{name}}}}}' -f id="$project_id" -q '.data.node.views.nodes[].name')"
if printf '%s\n' "$visoes" | grep -qxF "$VISAO"; then
  echo "Board '$VISAO' já existe"
else
  # Mutation 2: createProjectV2View (BOARD_LAYOUT agrupa pelo Status por padrão).
  gh api graphql -f query='mutation($p:ID!,$n:String!){createProjectV2View(input:{projectId:$p,name:$n,layout:BOARD_LAYOUT}){projectV2View{id}}}' -f p="$project_id" -f n="$VISAO" >/dev/null
  echo "Board '$VISAO' criado"
fi

# (c2) Campos visíveis do Board: ids buscados pelo nome na listagem (nada hardcoded).
visao_id="$(gh api graphql -f query='query($id:ID!){node(id:$id){... on ProjectV2{views(first:50){nodes{id name}}}}}' -f id="$project_id" -q ".data.node.views.nodes[] | select(.name==\"$VISAO\") | .id")"
ids_json=""
for nome_campo in "${CAMPOS_VISIVEIS[@]}"; do
  fid="$(gh project field-list "$numero" --owner "$DONO" --format json -q ".fields[] | select(.name==\"$nome_campo\") | .id")"
  [ -n "$fid" ] || { echo "campo $nome_campo não encontrado" >&2; exit 1; }
  ids_json="${ids_json:+$ids_json,}\"$fid\""
done
corpo_v="$(mktemp)"
printf '{"query":"mutation($input: UpdateProjectV2ViewInput!){updateProjectV2View(input:$input){projectV2View{id}}}","variables":{"input":{"viewId":"%s","configuration":{"visibleFieldIds":[%s]}}}}' "$visao_id" "$ids_json" > "$corpo_v"
gh api graphql --input "$corpo_v" >/dev/null
rm -f "$corpo_v"
echo "Board '$VISAO' mostra: ${CAMPOS_VISIVEIS[*]}"

opcao_id() {
  gh project field-list "$numero" --owner "$DONO" --format json \
    -q ".fields[] | select(.name==\"$CAMPO\") | .options[] | select(.name==\"$1\") | .id"
}

# (d) Issues sem duplicar, cartões e Status de cada um (reatribui também os cartões já existentes).
tail -n +2 "$INDICE" | while IFS=, read -r chave titulo coluna tipo pasta pr pai etiq; do
    etiq="${etiq%$'\r'}"
  etapa_valida "$coluna" || { echo "etapa inválida em $chave: $coluna" >&2; exit 1; }
  corpo="$AQUI/issues/$chave.md"
  [ -f "$corpo" ] || { echo "falta $corpo" >&2; exit 1; }

  pref="$(prefixo_do_titulo "$etiq")" || exit 1
  titulo_final="${pref}${titulo}"
  # Acha pelo título base OU por qualquer título que termine em "] <base>" (prefixo antigo ou novo).
  achada="$(gh issue list --repo "$REPO" --state all --limit 500 --json number,title -q ".[] | select(.title==\"$titulo\" or (.title | endswith(\"] $titulo\"))) | \"\(.number)|\(.title)\"" | head -n1 || true)"
  if [ -z "$achada" ]; then
    url="$(gh issue create --repo "$REPO" --title "$titulo_final" --body-file "$corpo")"
    echo "issue criada: $titulo_final -> $url"
  else
    numero_issue="${achada%%|*}"; titulo_atual="${achada#*|}"
    url="https://github.com/$REPO/issues/$numero_issue"
    if [ "$titulo_atual" != "$titulo_final" ]; then
      gh issue edit "$numero_issue" --repo "$REPO" --title "$titulo_final" >/dev/null
      echo "issue renomeada: '$titulo_atual' -> '$titulo_final'"
    else
      echo "issue já existe: $titulo_final -> $url"
    fi
  fi

  alvos="$url"
  # Só PR de código (Validation) vira cartão próprio; PR de spec (#120 a #122) é compartilhado.
  if [ -n "$pr" ] && [ "$coluna" = "Validation" ]; then alvos="$alvos https://github.com/$REPO/pull/${pr#\#}"; fi
  opcao="$(opcao_id "$coluna")"
  etiquetas_validas "$etiq" || { echo "etiqueta inválida em $chave: $etiq" >&2; exit 1; }
  lista_et="${etiq//+/,}"
  for alvo in $alvos; do
    if [ -n "$lista_et" ]; then
      case "$alvo" in
        */pull/*) gh pr edit "$alvo" --add-label "$lista_et" >/dev/null ;;
        *) gh issue edit "$alvo" --add-label "$lista_et" >/dev/null ;;
      esac
    fi
    item="$(gh project item-add "$numero" --owner "$DONO" --url "$alvo" --format json -q .id)"
    gh project item-edit --id "$item" --project-id "$project_id" --field-id "$campo_id" --single-select-option-id "$opcao" >/dev/null
    echo "cartão: $alvo -> $CAMPO $coluna"
  done
done

# (e) Relatório (nenhum token é impresso).
echo "Relatório: Project $numero ($TITULO_PROJETO), campo $CAMPO com 7 opções, Board '$VISAO'."
gh project item-list "$numero" --owner "$DONO" --limit 200 --format json -q ".items | length" | sed 's/^/cartões no Project: /'

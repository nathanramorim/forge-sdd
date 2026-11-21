# fix-104-15 — bots dos diagramas do README com cores e marcas diferentes das da LP

**Branch:** `fix/104-bots-com-cores-da-lp`
**Origem:** Iniciativa 104 (PR #118, `3c8a8d5`, bots dos diagramas do README)
**Tamanho:** `S`
**Status:** `spec`

## Problema

Os três bots desenhados em `docs/img/esteira-3-sessoes.svg` e `docs/img/agentes.svg` não batem com os personagens da LP atual (3 sessões). Quem vê o README e depois a LP encontra o Discovery com marca de outro personagem, o Builder na cor do Reviewer e o Reviewer na cor do antigo Líder.

## Evidência

Mapeamento oficial (`origin/main:docs/index.html`, `const SESSIONS`): `01. Discovery` `#8aa4ff` `diamond`; `02. Builder` `#4fd1b5` `wave`; `03. Reviewer` `#ff8a5c` `tri`.

Estado atual nos SVGs (ordem do fluxo, lida em `<ellipse cx="13" ... fill=...>` dos dois arquivos):

| Papel | Hoje | Correto |
|---|---|---|
| Discovery | `#8aa4ff` `bar` (errada: `bar` era o Planejador, `#c9a7ff`) | `#8aa4ff` `diamond` |
| Builder | `#ff8a5c` `tri` | `#4fd1b5` `wave` |
| Reviewer | `#f2a65a` `ring` (do Líder da LP antiga) | `#ff8a5c` `tri` |

Estado (orbe e olhos): a LP define o estado padrão de cada sessão (`defaultState` em `docs/index.html` e `STATE_DEFS`, ~linha 1600). O `bots.py` fixava orbe e olhos em `#4fd1b5` com olhos cápsula para todos os bots, o que não é cor de nenhum estado da LP.

| Papel | Estado | Orbe | Olhos |
|---|---|---|---|
| Discovery | `approval` | `#f5b04a` | `round` (círculos `cx=45`/`75 cy=86 r=6.5`) |
| Builder | `working` | `#4f8ef7` | `cap` (retângulos `x=41.5`/`71.5 y=78 w=7 h=16 rx=3.5`) |
| Reviewer | `searching` | `#38c6e8` | `round` |

`docs/diagramas/bots.py` tem `MARCAS` com `bar`, `ring`, `wave`, `tri`; não tem `diamond`. O `.astro` (`site/src/components/MascotFallbackSvg.astro`, linha 65) desenha `diamond` como `<polygon points="60,45.5 66.5,52 60,58.5 53.5,52" fill="rgba(18,24,48,0.6)" />`.

## Feature Correlacionada

- Origem: `docs/diagramas/bots.py` e os dois SVG/PNG de `docs/img/`, criados no `3c8a8d5` (#118).
- Causa-raiz: as cores e marcas das chamadas do `bots.py` foram escolhidas à mão a partir da LP antiga (5 personagens) em vez de lidas de `const SESSIONS`; nenhum critério comparava o diagrama ao mapeamento da LP.
- Iniciativa de origem: 104 (mesma do #118). Numeração: `fix-104-15` não existe em `origin/main`, nas branches `fix/*` nem no PR #131 (`fix-97-01`, iniciativa 97). O `fix-104-14` está no PR #130 (`fix/104-readme-esteira-duplicada`), ainda fora da `main`: a numeração depende de o #130 não ter reservado o 15.

## Decisões do projeto aplicadas

- DUP-01 e DUP-03: reusar `bots.py` e o diagram-studio; só acrescentar a marca `diamond`.
- P-03: tabela "O que já existe" abaixo.
- P-21: o `status.md` da 104 não é tocado (o #130 também o altera; colisão evitada).
- ADR-0004 (3 sessões): o mapeamento segue as 3 sessões da LP atual.
- Fora do escopo: Edição Missões (`docs/edicao/**`, `docs/design/edicao-missoes/**`) e código Go.
- Segurança: nada de segredo nem caminho de usuário.

## Correção

1. `docs/diagramas/bots.py`: adicionar em `MARCAS` a chave `diamond` com `<polygon points="60,45.5 66.5,52 60,58.5 53.5,52" fill="rgba(18,24,48,0.6)"/>`. Não remover as outras marcas. Acrescentar `ESTADOS` (`approval`, `working`, `searching`, com orbe e olhos como na tabela de estado) e o argumento `estado` por bot: `cor:marca:estado:cx:y[:escala]`. A entrada é validada (cor `^#[0-9a-fA-F]{6}$`, marca em `MARCAS`, estado em `ESTADOS`); valor desconhecido encerra com mensagem e código diferente de 0.
2. Regerar cada diagrama (plugin 0.2.0; largura 1280; os JSON de `docs/diagramas/` não mudam, os bots são injetados depois):
   - `uv run ~/.claude/plugins/cache/diagram-studio/diagram-studio/0.2.0/scripts/diagrama.py docs/diagramas/<nome>.json --sem-assinatura --saida docs/img/<nome>`
   - `uv run --with resvg-py docs/diagramas/bots.py docs/img/<nome>.svg 1280 <bots>`
3. Bots, sempre na ordem do fluxo (Discovery, Builder, Reviewer):
   - `esteira-3-sessoes` (escala 0.6, y 112; posições iguais às atuais): `"#8aa4ff:diamond:approval:340:112" "#4fd1b5:wave:working:740:112" "#ff8a5c:tri:searching:940:112"`
   - `agentes` (escala 0.5, y 168; posições iguais às atuais): `"#8aa4ff:diamond:approval:240:168:0.5" "#4fd1b5:wave:working:640:168:0.5" "#ff8a5c:tri:searching:1040:168:0.5"`
   (cx = translate x atual + 60 vezes a escala: 304+36, 704+36, 904+36; 210+30, 610+30, 1010+30.)
4. O PNG sai a 2560 px de largura (1280 x 2), como hoje. `README.md` referencia as imagens por caminho e o `alt` não cita cor: não muda.
5. Abrir cada PNG e conferir os três bots (Discovery azul com losango, Builder verde-água com onda, Reviewer laranja com triângulo, sem sobreposição com o texto).

Arquivos que o Builder altera (todos em `origin/main`): `docs/diagramas/bots.py`, `docs/img/esteira-3-sessoes.svg`, `docs/img/esteira-3-sessoes.png`, `docs/img/agentes.svg`, `docs/img/agentes.png`. Não altera: `docs/diagramas/*.json`, `README.md`, `npm/README.md`.

## Seguimento (fora do escopo, para o dono decidir)

- `forge-sdd-avatar.png` e os 3 diagramas da assinatura da 106 existem só em `origin/feat/106-pr-draft-revisao-e-ci`, em `sdd/discovery/106-pr-draft-revisao-e-ci/diagramas/`; a assinatura usa um bot `#8aa4ff` com marca `bar` (a marca da LP para o Discovery é `diamond`).
- Deck comercial do hub-busize, `~/Downloads/deck-forge-ci-so-quando-pronto/` (fora do git), usa esse avatar.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Mapeamento oficial | `docs/index.html` `const SESSIONS` | 3 sessões com cor e marca | Reusar como fonte | É a verdade do produto |
| Desenho da marca `diamond` | `site/src/components/MascotFallbackSvg.astro:65` | Polígono de 4 pontos | Copiar para `MARCAS` | Mesmo desenho do `.astro` |
| Injetor de bots | `docs/diagramas/bots.py` | Sem `diamond` | Estender | DUP-03 |
| JSON dos diagramas | `docs/diagramas/esteira-3-sessoes.json`, `agentes.json` | Sem bots (injetados depois) | Reusar sem mudança | Bots não estão no JSON |
| Teste Go que lê as imagens | `*_test.go` com `esteira-3-sessoes` ou `img/agentes` | Nenhum | Sem golden a ajustar | `grep` no repositório |
| README | `README.md:80` e `:93` | Imagens por caminho, `alt` sem cor | Não mudar | Só o conteúdo das imagens muda |

## Dependências novas

Nenhuma. `uv`, `resvg_py` e o plugin diagram-studio 0.2.0 já foram usados no #118.

## Rules e IDs da Constituição Aplicáveis

DUP-01, DUP-03, GOV-01, SEG-07 (sem segredo), P-03, P-21, P-25.

## Descrição de negócio (release notes)

- **Sem mudança visível para o produto:** os bots dos diagramas do README passam a ter as mesmas cores dos personagens da LP.

## A confirmar

- O dono quer tratar o avatar e a assinatura da 106 e o deck do hub-busize numa iniciativa própria? Esta spec não os altera.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# a marca diamond existe no bots.py, igual ao .astro, e as outras seguem definidas
grep -F '<polygon points="60,45.5 66.5,52 60,58.5 53.5,52" fill="rgba(18,24,48,0.6)"/>' docs/diagramas/bots.py >/dev/null
grep -F '"wave":' docs/diagramas/bots.py >/dev/null
grep -F '"tri":' docs/diagramas/bots.py >/dev/null
# em cada SVG: cores dos bots na ordem Discovery, Builder, Reviewer, e marcas diamond, wave, tri
for n in esteira-3-sessoes agentes; do
  test "$(grep -o '<ellipse cx="13" cy="90" rx="7" ry="11" fill="#[0-9a-f]*"' docs/img/$n.svg | grep -o '#[0-9a-f]*' | tr '\n' ' ')" = '#8aa4ff #4fd1b5 #ff8a5c '
  test "$(grep -oE 'points="60,45.5 66.5,52 60,58.5 53.5,52"|M51 53 Q55.5 47|points="60,46.5 66,57 54,57"' docs/img/$n.svg | tr '\n' '|')" = 'points="60,45.5 66.5,52 60,58.5 53.5,52"|M51 53 Q55.5 47|points="60,46.5 66,57 54,57"|'
  # negativo: a cor do Líder antigo e as marcas bar e ring não aparecem
  test "$(grep -oF '#f2a65a' docs/img/$n.svg | wc -l)" -eq 0
  test "$(grep -oF '<rect x="52" y="49.7" width="16" height="4.6" rx="2.3"' docs/img/$n.svg | wc -l)" -eq 0
  test "$(grep -oF '<circle cx="60" cy="52" r="4.6" fill="none"' docs/img/$n.svg | wc -l)" -eq 0
  # P-25: orbes dos estados da LP presentes (3 de cada: halo, orbe e olhos de cada bot)
  test "$(grep -oF '#f5b04a' docs/img/$n.svg | wc -l)" -eq 3
  test "$(grep -oF '#38c6e8' docs/img/$n.svg | wc -l)" -eq 3
  # o orbe (círculo da antena) tem a cor do estado: exatamente um por cor, e #4fd1b5 só no corpo do Builder
  test "$(grep -oF 'r="7.8" fill="#f5b04a"' docs/img/$n.svg | wc -l)" -eq 1
  test "$(grep -oF 'r="7.8" fill="#4f8ef7"' docs/img/$n.svg | wc -l)" -eq 1
  test "$(grep -oF 'r="7.8" fill="#38c6e8"' docs/img/$n.svg | wc -l)" -eq 1
  # negativo: nenhum bot usa o orbe antigo #4fd1b5
  test "$(grep -oF 'r="7.8" fill="#4fd1b5"' docs/img/$n.svg | wc -l)" -eq 0
  # olhos: 2 cápsulas (Builder) e 4 círculos (Discovery e Reviewer)
  test "$(grep -oF '<rect x="41.5" y="78" width="7" height="16" rx="3.5"/>' docs/img/$n.svg | wc -l)" -eq 1
  test "$(grep -oF '<circle cx="45" cy="86" r="6.5"/>' docs/img/$n.svg | wc -l)" -eq 2
  # PNG com o dobro da largura do SVG (1280 x 2)
  file docs/img/$n.png | grep -F '2560 x' >/dev/null
done
# negativo: o bots.py recusa estado, marca e cor desconhecidos (código diferente de 0)
cp docs/img/agentes.svg /tmp/p25-bots.svg
for a in "#8aa4ff:tri:nada:1:1" "#8aa4ff:nada:working:1:1" "nada:tri:working:1:1"; do
  if uv run --with resvg-py docs/diagramas/bots.py /tmp/p25-bots.svg 1280 "$a" >/dev/null 2>&1; then exit 1; fi
done
# negativo: Edição Missões, README e JSON de origem intactos
test "$(git diff --name-only origin/main -- docs/edicao docs/design/edicao-missoes README.md npm/README.md docs/diagramas/esteira-3-sessoes.json docs/diagramas/agentes.json | wc -l)" -eq 0
# o README continua apontando para as duas imagens
grep -F 'docs/img/esteira-3-sessoes.svg' README.md >/dev/null
grep -F 'docs/img/agentes.svg' README.md >/dev/null
# nenhum golden depende das imagens: o teste dos marcadores da edição segue verde
go test -list 'TestScaffold_DocsEdicao' ./internal/scaffold | grep -F 'TestScaffold_DocsEdicao' >/dev/null
go test ./internal/scaffold -run 'TestScaffold_DocsEdicao' -count=1
```

Verificação visual (manual): abrir `docs/img/esteira-3-sessoes.png` e `docs/img/agentes.png` e conferir os três bots.

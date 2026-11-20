# feat-101-07 — Edição alinhada aos pipelines: 3 missões, 3 robôs

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** M
**Depende de:** feat-101-06 (robôs 2D), fix-101-01 (`missoes.json`)

## Objetivo

O Nathan apontou que a visão de 5 estações (com o Arquiteto) já não existe: o produto agora são pipelines, as 3 sessões Discovery, Builder e Reviewer do README. A edição foi construída em cima dos 5. Esta task a alinha ao produto atual: 3 missões e 3 robôs na equipe, mais o fecho "Missão cumprida".

## Decisões registradas (Nathan, 19/11/2026)

- "Tema com 3 missões e 3 robôs": Discovery é a Missão 1 (o plano), Builder a Missão 2 (a obra) e Reviewer a Missão 3 (o vigia), com o fecho "Missão cumprida" para a sua aprovação.
- O texto da LP (os 5 cartões da seção de estações) **não muda** nesta task; muda só o que a edição põe por cima. Atualizar o conteúdo da LP para as 3 sessões é outra decisão.

## Modelo do módulo

**Enxuto** (ARQ-11): `missoes.json`, o gerador anexo, o JS e o CSS da camada, os dois READMEs e o histórico. Nenhum Go.

## O que muda

1. **`missoes.json`**, a fonte única, passa a ter 3 missões e o fecho:

   | Missão | Sessão | Estações da LP que ela cobre |
   |---|---|---|
   | Missão 1: o plano | Discovery | Concepção, Planejamento |
   | Missão 2: a obra | Builder | Construção |
   | Missão 3: o vigia | Reviewer | Governança |
   | Missão cumprida (fecho) | Você aprova | Operação |

2. **LP:** cada cartão ganha o rótulo da missão que cobre a sua estação (o do Líder diz "Missão cumprida"). A placa de neon lê o texto do fecho no JSON.
3. **Equipe do topo:** 3 robôs (Discovery, Builder, Reviewer). Os retratos dos cartões continuam um por estação.
4. **Banner, espécime, documento do design system e os dois READMEs:** passam a mostrar as 3 missões e o fecho, gerados da fonte única.
5. **Textos da iniciativa e do histórico:** "5 estações como missões" vira as 3 missões e o fecho.

## Fora do escopo

Reescrever o conteúdo da seção de estações da LP, o hero e os textos da wiki; os retratos dos cartões; nome e vigência da edição.

## Decisões do projeto aplicadas

- **Discovery 101:** substitui a decisão do Gate 1 que mapeava 5 estações em 5 missões; registrada na própria discovery.
- **Critérios 1, 3 e 7 da iniciativa** (`criteria.md`); o 3 passa a falar das 3 missões e do fecho.
- **FLX-02:** branch única da iniciativa.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Fonte única das missões | `docs/design/edicao-missoes/missoes.json` | 5 missões | estender: 3 missões, `estacoes` por missão e `fecho` |
| Leitura do JSON | `docs/edicao/edicao.js`, `rotularCartoes` e placa | Mapeia estação a missão por lista fixa | estender: usa `estacoes` do JSON |
| Geradores | `anexos/gerador-banner/*.py` | Montam as missões a partir do JSON | estender com `titulos_missoes()` em `tokens.py` |
| Equipe | `personagens.py`, `ORDEM_EQUIPE` | 5 robôs | estender: 3 |
| Textos que citam 5 missões | READMEs, `history.md`, `criteria.md`, specs 101-01 e fix-101-01 | Várias | atualizar |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11**, **DUP-02**, **COD-01**, **COD-04**, **COD-05**, **SEG-07**, **TST-01**, **TST-05**, **FLX-01**, **FLX-02**.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
base=$(git merge-base HEAD origin/main)

# 1. A fonte única tem 3 missões e o fecho, cada estação da LP em exatamente uma
node -e '
const m = require("./docs/design/edicao-missoes/missoes.json");
if (m.missoes.length !== 3 || !m.fecho || !m.fecho.titulo) { console.error("estrutura"); process.exit(1); }
const todas = [].concat(...m.missoes.map(x => x.estacoes), m.fecho.estacoes);
for (const e of ["Operação","Concepção","Planejamento","Construção","Governança"])
  if (todas.filter(x => x === e).length !== 1) { console.error("estação sem missão única:", e); process.exit(1); }
console.log("ok");'

# 2. READMEs, banner e espécime com o mesmo texto da fonte única
node -e '
const fs = require("fs");
const m = require("./docs/design/edicao-missoes/missoes.json");
const titulos = m.missoes.map(x => x.titulo).concat(m.fecho.titulo);
for (const f of ["README.md", "npm/README.md", "docs/img/banner-v5-missoes.svg", "docs/design/edicao-missoes/especime.svg"]) {
  const t = fs.readFileSync(f, "utf8");
  for (const x of titulos) if (!t.includes(x)) { console.error(f, "sem:", x); process.exit(1); }
}
for (const f of ["README.md", "npm/README.md"]) if (/Missão 4|Missão 5|a rota/.test(fs.readFileSync(f, "utf8"))) { console.error("resto de missão antiga:", f); process.exit(1); }'

# 3. A camada lê as estações do JSON, sem lista fixa de missões, e a equipe tem 3 robôs
grep -q 'estacoes' docs/edicao/edicao.js
test "$(grep -c -i -E 'o plano|a obra|o vigia|cumprida' docs/edicao/edicao.js docs/edicao/edicao.css | awk -F: '{s+=$2} END{print s}')" = "0"
( cd sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner && python3 personagens.py /tmp/eq-ensaio >/dev/null && cmp /tmp/eq-ensaio/equipe.svg ../../../../../docs/edicao/equipe.svg; r=$?; rm -rf __pycache__ /tmp/eq-ensaio; exit $r )
test "$(grep -o '<ellipse cx="[0-9]*" cy="324"' docs/edicao/equipe.svg | wc -l | tr -d ' ')" = "3"

# 4. Os geradores reproduzem banner e espécime
( cd sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner && python3 banner.py /tmp/b-ensaio.svg >/dev/null && cmp /tmp/b-ensaio.svg ../../../../../docs/img/banner-v5-missoes.svg && python3 specimen.py /tmp/s-ensaio.svg >/dev/null && cmp /tmp/s-ensaio.svg ../../../../../docs/design/edicao-missoes/especime.svg; r=$?; rm -rf __pycache__ /tmp/b-ensaio.svg /tmp/s-ensaio.svg; exit $r )

# 5. Marca de terceiros e HTML intacto fora dos marcadores
test "$(grep -rn -i -E 'gta|grand theft|vice city|pricedown|rockstar|take-two|lucia|jason' docs/edicao | wc -l | tr -d ' ')" = "0"
for f in docs/index.html docs/wiki.html; do
  diff <(git show "$base:$f") <(sed '/edicao:inicio/,/edicao:fim/d' "$f") >/dev/null || { echo "mudou fora dos marcadores: $f"; exit 1; }
done

# 6. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

Topo com 3 robôs, placa "Missão cumprida" e cartões com o rótulo certo; banner e README com as 3 missões e o fecho. Peça registrada em `revisao-pi.md`.

## Status

- [x] Implementado em feat-101-07 (19/11/2026): critérios 1 a 5 com saída 0 e `forge gate` em bloqueio sem pendências.

# fix-101-01 — Alinhar README, banner e design system à Edição Missões

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** M
**Ordem:** primeira task da fila. O README está público com o defeito.

## Objetivo

O PR #113 publicou o tema no README sem conhecer as decisões desta iniciativa. Este fix deixa README, banner, notas da versão e design system dentro do que o Gate 1 aprovou: homenagem ao clima, com o nome "Edição Missões", sem "GTA VI" escrito, sem letreiro parecido e sem interface do jogo.

## Feature Correlacionada

- **Origem:** PR #113, `feat(readme): versão 5.0.0, novidades e edição GTA VI` (commit `58d96db`). Arquivos: `README.md`, `npm/README.md`, `sdd/releases/history.md`, `docs/img/banner-v5-gta-vi.svg`, `docs/img/banner-v5-gta-vi.png` e a pasta `docs/design/gta-vi/`.
- **Causa-raiz:** a entrega foi feita fora da iniciativa. A busca por decisões olhou só o `main`, e a discovery 101 estava apenas na branch `feat/101-tema-missoes-v5` (PR #111, em rascunho).
- **Evidência do defeito**, no `main` (commit `58d96db`):

| Critério da 101 | O que está publicado |
|---|---|
| 1: sem numeral nem letreiro parecido | O banner empilha "forge / sdd" em letras de bloco sobre um V em degradê, na composição do logotipo do jogo |
| 1: sem fonte que imite o título | `tokens.json`, `tokens.css` e o documento põem as duas fontes do conceito como primeira opção das pilhas |
| 1: sem interface do jogo | Banner e espécime trazem estrelas de procurado, dinheiro, radar e botões de console |
| 2: "GTA VI" só em legenda de post | `grep -n -i "gta vi" README.md npm/README.md sdd/releases/history.md` devolve o link "edição GTA VI", a linha da tabela de versões e o item das notas. O banner também escreve "edição GTA VI" |
| 3: as 5 estações como missões | Não aparecem em nenhuma peça |

## Modelo do módulo

**Enxuto** (ARQ-11): só arquivos de documentação e imagem. Nenhum pacote Go novo e nenhum arquivo `.go` de produção muda.

## Decisões do projeto aplicadas

- **Discovery 101, Gate 1 (19/11/2026):** homenagem ao clima; sem logo, letreiro, "VI", personagens, trilha ou interface do jogo; "GTA VI" nunca no banner nem no README; o "5" como destaque gráfico próprio; tipografia da marca.
- **Segunda rodada de decisões (19/11/2026):** do design system ficam a paleta do pôr do sol, o neon e as molduras; o nome é "Edição Missões".
- **Critérios 1, 2, 3 e 5 da iniciativa** (`criteria.md`).
- **FLX-02:** trabalho na branch única da iniciativa, `feat/101-tema-missoes-v5`.
- **Release:** a entrada `### Versão 5.0.0` de `sdd/releases/history.md` mantém o formato que o `release.yml` extrai.

## O que muda

1. **Nomes.** `docs/design/gta-vi/` vira `docs/design/edicao-missoes/`. `docs/img/banner-v5-gta-vi.svg` e `.png` viram `banner-v5-missoes.svg` e `.png`. As variáveis `--gta-*` viram `--missoes-*`. O título do documento passa a ser "Design system da Edição Missões".
2. **Texto.** "edição GTA VI" vira "Edição Missões" em `README.md`, em `npm/README.md` e na entrada da 5.0.0 de `sdd/releases/history.md` (título e item).
3. **Banner.**
   - Sai: o texto "GTA VI", o V com as palavras empilhadas, as letras de bloco, as estrelas de procurado e os botões de console (△ ○ ✕, L1, R1).
   - Entra: a marca `forge-sdd` na tipografia da marca (Montserrat 800), o "5" como destaque gráfico em degradê, o selo "Edição Missões" e as 5 missões no lugar das abas.
   - Fica: pôr do sol, sol listrado, palmeiras, silhueta da cidade, barras pretas e linha neon.
   - O PNG do npm é gerado de novo a partir do SVG.
4. **Design system.**
   - As pilhas tipográficas adotadas passam a ser as da marca (Montserrat e JetBrains Mono). Os nomes das fontes do conceito saem de `tokens.json`, `tokens.css`, `identidade.json` e do documento.
   - Saem do documento e do espécime os componentes que são interface do jogo: estrelas e dinheiro, radar, roda de seleção, slots de ação, discos de vida e energia e dicas com botões de console.
   - Ficam cores, degradês, moldura, linha neon, abas em chip, lista de opções e painel de descrição.
   - A origem continua creditada em texto comum, em uma linha com o link, seguida do aviso "conteúdo original, sem ligação com a Rockstar/Take-Two". O campo `origem.titulo` sai do `tokens.json`; o link basta.
5. **Missões em fonte única.** Novo `docs/design/edicao-missoes/missoes.json`:

   ```json
   {
     "edicao": "Edição Missões",
     "promessa": "A IA executa, o plano segura, você aprova.",
     "missoes": [
       { "numero": 1, "estacao": "Concepção", "titulo": "Missão 1: o plano" },
       { "numero": 2, "estacao": "Planejamento", "titulo": "Missão 2: a rota" },
       { "numero": 3, "estacao": "Construção", "titulo": "Missão 3: a obra" },
       { "numero": 4, "estacao": "Governança", "titulo": "Missão 4: o vigia" },
       { "numero": 5, "estacao": "Operação", "titulo": "Missão 5: missão cumprida" }
     ]
   }
   ```

   `README.md` e `npm/README.md` ganham um bloco curto com as 5 missões, com o mesmo texto.
6. **Marcadores.** Tudo o que é da edição nos dois READMEs (banner, selos, link e bloco das missões) fica entre `<!-- edicao:inicio -->` e `<!-- edicao:fim -->`. A feat-101-03 usa esses marcadores para devolver o layout padrão.
7. **Revisão de propriedade intelectual.** Novo `sdd/discovery/101-tema-missoes-v5/revisao-pi.md`, com uma linha por peça (banner, espécime, README): o que foi conferido, por quem e quando.

## Fora do escopo

LP e wiki (feat-101-01), mascotes (feat-101-02), retorno agendado (feat-101-03), imagem de compartilhamento e posts.

## Como regenerar as imagens

O gerador usado no #113 está em `sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner/` (Python puro, sem dependências):

```bash
cd sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner
python3 documento.py
python3 banner.py   ../../../../../docs/img/banner-v5-missoes.svg
python3 specimen.py ../../../../../docs/design/edicao-missoes/especime.svg
./render.sh ../../../../../docs/img/banner-v5-missoes.svg ../../../../../docs/img/banner-v5-missoes.png 1280 420 1.5
```

`tokens.py` guarda os valores, `parts.py` as peças (palmeira, sol, linha neon) e `documento.py` gera `tokens.json`, `tokens.css`, `identidade.json` e o documento do design system. As missões são lidas de `missoes.json`. O letreiro em bloco e o V do #113 saíram do gerador neste fix. O `render.sh` só roda no macOS, com o Chrome instalado.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Decisões sobre o tema | `sdd/discovery/101-tema-missoes-v5/discovery.md` e `criteria.md` | Gate 1 aprovado e 7 critérios | reusar: são a régua deste fix |
| Banner e espécime | `docs/img/banner-v5-gta-vi.svg`, `docs/design/gta-vi/especime.svg` | Peças do #113 | estender: renomear e redesenhar o que fere a 101 |
| Gerador das peças | `sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner/` | `banner.py`, `specimen.py`, `parts.py`, `tokens.py`, `render.sh` | reusar |
| Marca na tipografia da marca | `docs/img/banner.svg` | Logo hexagonal e nome em Montserrat 800 | reusar como referência |
| Texto das missões | `discovery.md`, seção "Proposta" | As 5 missões e a promessa | criar `missoes.json` como fonte única |
| Marcadores de edição | `README.md`, `npm/README.md` | Não existem | criar |
| Registro de revisão de PI | pasta da iniciativa | Não existe | criar `revisao-pi.md` |
| Teste que lê o README | `internal/scaffold/pipeline_scaffold_test.go`, `TestScaffold_DocsCitamBriefConstituicaoETmux` | Exige "brief", "/constitution" e "tmux" no README | reusar: o fix não remove esses termos |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11:** modelo enxuto, declarado acima.
- **DUP-02:** tabela "O que já existe" preenchida; as missões passam a ter uma fonte única.
- **COD-01:** cada arquivo com um papel (tokens, missões, documento, imagem).
- **COD-04:** nomes e arquivos antigos saem neste mesmo fix; nada de cópia antiga ao lado da nova.
- **COD-05:** textos em pt-BR, nível padrão.
- **SEG-07:** nenhuma dependência nova.
- **TST-01 e TST-05:** `go test ./...` e `go vet ./...` no portão.
- **FLX-01 e FLX-02:** spec antes do código; branch única da iniciativa.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
# 1. Nenhuma marca de terceiros fora da linha de crédito
test "$(grep -rn -i -E "gta|grand theft|vice city|pricedown|rockstar|take-two" \
  --include="*.md" --include="*.svg" --include="*.json" --include="*.css" --include="*.frag" \
  README.md npm/README.md sdd/releases/history.md docs/img docs/design \
  | grep -v -E "sem ligação com a Rockstar/Take-Two|behance\.net/gallery/224512205" | wc -l | tr -d ' ')" = "0"

# 2. Nomes antigos saem, nomes novos entram
test ! -e docs/design/gta-vi && test ! -e docs/img/banner-v5-gta-vi.svg && test ! -e docs/img/banner-v5-gta-vi.png
test -f docs/design/edicao-missoes/tokens.css && test -f docs/design/edicao-missoes/missoes.json
test -f docs/img/banner-v5-missoes.svg && test -f docs/img/banner-v5-missoes.png
test "$(grep -c -- '--gta-' docs/design/edicao-missoes/tokens.css)" = "0"

# 3. Os READMEs apontam para o banner novo e usam o nome da edição
grep -q 'docs/img/banner-v5-missoes.svg' README.md
grep -q 'banner-v5-missoes.png' npm/README.md
for f in README.md npm/README.md sdd/releases/history.md; do
  grep -q 'Edição Missões' "$f" || { echo "sem o nome da edição: $f"; exit 1; }
done
test "$(grep -c '^### Versão 5.0.0' sdd/releases/history.md)" = "1"

# 4. As missões (3 e o fecho, desde a feat-101-07) saem da fonte única e aparecem iguais nos dois READMEs
node -e '
const fs = require("fs");
const m = JSON.parse(fs.readFileSync("docs/design/edicao-missoes/missoes.json", "utf8"));
if (m.missoes.length !== 3 || !m.fecho) { console.error("esperava 3 missões e o fecho"); process.exit(1); }
for (const f of ["README.md", "npm/README.md"]) {
  const t = fs.readFileSync(f, "utf8");
  for (const x of m.missoes.concat(m.fecho)) if (!t.includes(x.titulo)) { console.error(f, "sem:", x.titulo); process.exit(1); }
}
console.log("ok");'

# 5. Trechos da edição entre marcadores, com abertura e fechamento casados
for f in README.md npm/README.md; do
  a=$(grep -c 'edicao:inicio' "$f"); b=$(grep -c 'edicao:fim' "$f")
  test "$a" -ge 1 && test "$a" = "$b" || { echo "marcadores: $f"; exit 1; }
done

# 6. Imagens válidas e sem o numeral antigo
xmllint --noout docs/img/banner-v5-missoes.svg docs/design/edicao-missoes/especime.svg
test "$(grep -c 'id="gV"' docs/img/banner-v5-missoes.svg)" = "0"

# 7. Links e imagens locais do README existem
for p in $(grep -o -E '(src|href)="(docs|sdd)/[^"#]+"' README.md | sed -E 's/.*="([^"]+)"/\1/'); do
  test -e "$p" || { echo "link quebrado: $p"; exit 1; }
done

# 8. Revisão de propriedade intelectual registrada (banner, espécime e README)
test "$(grep -c -i -E 'banner|espécime|README' sdd/discovery/101-tema-missoes-v5/revisao-pi.md)" -ge 3

# 9. Portão do projeto
go test ./... && go vet ./...
```

## Status

- [x] Implementado em `fix-101-01` (19/11/2026): critérios 1 a 9 com saída 0 e `forge gate` em modo bloqueio sem pendências.

## Conferência visual (Gate 2)

- O banner na largura do README, nos temas claro e escuro do GitHub.
- O "5" não lembra o numeral do jogo e nenhuma peça usa interface do jogo.
- Os selos continuam com `nathanramorim/forge-sdd` e o texto do banner vem do projeto (passo 7 da skill `readme-comercial`).

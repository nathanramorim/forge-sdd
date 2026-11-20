# feat-101-04 — Visual mais intenso da Edição Missões na LP

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** M
**Depende de:** feat-101-01 (camada e porteiro de vigência) e feat-101-02 (mascotes)

## Objetivo

Depois de ver a camada da feat-101-01 funcionando, o Nathan achou que "não lembra o estilo". Esta task carrega o clima no visual da LP: o topo vira um pôr do sol em tela cheia, os cartões ganham contorno grosso em mosaico, entra uma placa de neon com palavras nossas e o fundo passa do quase preto para um degradê ameixa. Tudo some junto com a edição, em 23/11/2026, 0h de Brasília.

## Decisões registradas (Nathan, 19/11/2026)

- "Mais clima, dentro da 101": a dose maior vale, mas as regras de marca da iniciativa continuam. Fora: nome do jogo, logo, letreiro parecido, personagens, interface do jogo.
- Elementos escolhidos: pôr do sol em tela cheia (sol listrado, palmeiras, skyline art déco), cartões em mosaico com contorno grosso, placa de neon com palavras nossas, fundo em degradê rosa e laranja.

## Modelo do módulo

**Enxuto** (ARQ-11): só `docs/edicao/` (CSS, um SVG e poucas linhas no JS) e o gerador anexo. Nenhum pacote Go novo.

## Como funciona

1. **Arte do topo.** `docs/edicao/hero-arte.svg`, gerado por `hero.py` no gerador anexo (mesmas peças do banner: céu, sol listrado, palmeiras, skyline). Sem texto e sem rosto. A base da arte se funde com a cor do fundo da página.
2. **Topo em tela cheia.** `.hero` ganha a arte como fundo, com a largura da janela (via pseudo-elemento, sem mudar o HTML), altura mínima maior e sombra no texto para manter a leitura.
3. **Placa de neon.** Novo elemento `.edicao-placa`, criado pelo `edicao.js`, com a última missão em caixa alta ("MISSÃO CUMPRIDA"), lida de `missoes.json` (sem cópia do texto). Tubo em branco com brilho rosa e coral, levemente inclinada. A cintilação só roda com `prefers-reduced-motion: no-preference`.
4. **Cartões em mosaico.** `.agent-card` e `.step-node` com contorno de 3 px, cantos de 6 px e sombra dura deslocada, na cor do degradê assinatura (a do agente no cartão ativo).
5. **Fundo.** Degradê do preto da moldura para o ameixa, só com tokens do design system. As variáveis de superfície passam a usar `ameixa`, mantendo texto branco com contraste acima de 4,5:1.

### Fora do que a 101 permite (não entra)

Nome do jogo ou das empresas, logo ou numeral parecido, letreiro em letras de bloco, personagens, rostos, interface de jogo (estrelas, radar, dinheiro, botões de console).

## Fora do escopo

Wiki (segue com a camada atual), textos e ordem da LP, mascotes, workflow de encerramento.

## Decisões do projeto aplicadas

- **Discovery 101, Gate 1:** homenagem ao clima; mosaico de painéis com contorno grosso e placa com palavras nossas estavam na proposta original.
- **Critérios 1, 2, 5, 6 e 7 da iniciativa** (`criteria.md`).
- **Vigência:** tudo sob `html[data-edicao="missoes"]`; o PR de encerramento apaga `docs/edicao/` inteira, inclusive o SVG novo.
- **FLX-02:** branch única da iniciativa.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Peças do pôr do sol | `sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner/parts.py` | `palm`, `striped_sun`, `skyline`, `grad` | reusar no novo `hero.py` |
| Camada da edição | `docs/edicao/edicao.css`, `edicao.js` | Variáveis, chips, faixa do 5.0, rótulos das missões | estender |
| Texto das missões | `docs/design/edicao-missoes/missoes.json` | 5 missões | reusar: a placa lê daqui |
| Tokens de cor | `docs/design/edicao-missoes/tokens.css` | `ameixa`, `vinho`, degradês | reusar; nenhuma cor nova |
| Estilo dos cartões | `docs/index.html`: `.agent-card`, `.step-node` | Borda de 1 px e raio de 18 px | estender por sobrescrita, sem editar o bloco |
| Marcadores no HTML | `docs/index.html`, `docs/wiki.html` | Já existem | reusar; o HTML não muda |

## Dependências novas

Nenhuma. Sem fonte, imagem de terceiros ou biblioteca.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11:** modelo enxuto.
- **DUP-02:** peças e texto reusados; nada duplicado.
- **COD-01:** arte no gerador, estilo no CSS, comportamento no JS.
- **COD-04:** tudo removível pela remoção de `docs/edicao/`.
- **COD-05:** textos em pt-BR.
- **SEG-07:** nenhuma dependência nova.
- **TST-01 e TST-05:** `go test ./...` e `go vet ./...`.
- **FLX-01 e FLX-02:** spec antes do código; branch da iniciativa.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
base=$(git merge-base HEAD origin/main)

# 1. Arte válida e reprodutível pelo gerador
xmllint --noout docs/edicao/hero-arte.svg
( cd sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner && python3 hero.py /tmp/hero-ensaio.svg >/dev/null && cmp /tmp/hero-ensaio.svg ../../../../../docs/edicao/hero-arte.svg; r=$?; rm -rf __pycache__ /tmp/hero-ensaio.svg; exit $r )

# 2. Toda regra de primeiro nível só vale com a marca no <html>
test "$(grep -E '^[^@ }/*][^{]*\{' docs/edicao/edicao.css | grep -v -c 'html\[data-edicao="missoes"\]')" = "0"

# 3. Topo usa a arte; placa e contorno de mosaico existem; o texto da placa vem do JSON
grep -q 'hero-arte.svg' docs/edicao/edicao.css
grep -q 'edicao-placa' docs/edicao/edicao.js && grep -q 'edicao-placa' docs/edicao/edicao.css
test "$(grep -c -i 'cumprida' docs/edicao/edicao.js docs/edicao/edicao.css | awk -F: '{s+=$2} END{print s}')" = "0"
grep -q 'agent-card' docs/edicao/edicao.css && grep -q 'step-node' docs/edicao/edicao.css

# 4. O HTML não muda fora dos marcadores
for f in docs/index.html docs/wiki.html; do
  diff <(git show "$base:$f") <(sed '/edicao:inicio/,/edicao:fim/d' "$f") >/dev/null || { echo "mudou fora dos marcadores: $f"; exit 1; }
done

# 5. Contraste do texto sobre o fundo e o cartão (mínimo 4,5:1) e cintilação respeitando o movimento reduzido
node -e '
const t = require("./docs/design/edicao-missoes/tokens.json");
const L = h => { const v = [1,3,5].map(i => parseInt(h.slice(i,i+2),16)/255).map(c => c<=.03928 ? c/12.92 : ((c+.055)/1.055)**2.4); return .2126*v[0]+.7152*v[1]+.0722*v[2]; };
const R = (a,b) => { const [x,y] = [L(a),L(b)].sort((p,q)=>q-p); return (x+.05)/(y+.05); };
const ameixa = t.cor.superficie.ameixa.valor, moldura = t.cor.superficie["preto-moldura"].valor;
for (const [n,c,f] of [["branco/ameixa","#ffffff",ameixa],["muted/ameixa","#cdbdd3",ameixa],["branco/moldura","#ffffff",moldura]])
  if (R(c,f) < 4.5) { console.error("contraste baixo:", n, R(c,f).toFixed(2)); process.exit(1); }
console.log("ok");'
grep -q 'prefers-reduced-motion: no-preference' docs/edicao/edicao.css

# 6. Nenhuma marca de terceiros
test "$(grep -rn -i -E 'gta|grand theft|vice city|pricedown|rockstar|take-two' docs/edicao | wc -l | tr -d ' ')" = "0"

# 7. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

- LP com `?edicao=on` e `?edicao=off`: desligada, igual à de hoje.
- Topo legível sobre a arte; contorno e sombra dos cartões sem cortar o conteúdo; placa sem encostar no texto.
- Peça registrada em `sdd/discovery/101-tema-missoes-v5/revisao-pi.md`.

## Status

- [x] Implementado em feat-101-04 (19/11/2026): critérios 1 a 7 com saída 0 e `forge gate` em bloqueio sem pendências.

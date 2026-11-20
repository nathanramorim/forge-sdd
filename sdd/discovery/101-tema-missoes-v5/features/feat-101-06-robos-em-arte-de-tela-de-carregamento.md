# feat-101-06 — Robôs como personagens em arte de tela de carregamento

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** G
**Depende de:** feat-101-05 (figurino) e feat-101-04 (topo em tela cheia)

## Objetivo

O Nathan achou que os mascotes 3D "não lembram nada" do GTA. O que o reconhecimento pede é o estilo de ilustração, e não esferas 3D com acessórios: figura inteira em pose, contorno grosso, cor chapada com luz de contorno rosa. Esta task redesenha os cinco agentes como personagens 2D nesse estilo, usados no avatar de cada cartão e em uma formação ("a equipe") no topo da LP, durante a edição.

## Limite de marca (da 101, mantido)

São os robôs do forge-sdd: cabeça redonda com visor, antena e a marca de cada estação. O estilo vem da técnica (contorno grosso, cor chapada, luz de contorno), não de personagem algum. Nenhuma pose, roupa ou composição reproduz arte ou personagem do jogo (em especial a pose de casal encostado em carro da arte de capa). Sem nome, logo, texto ou interface do jogo; armas e charuto continuam de brinquedo.

## Modelo do módulo

**Enxuto** (ARQ-11): um gerador Python no anexo da iniciativa, seis SVG em `docs/edicao/` e regras de CSS. Nenhum Go.

## Como funciona

1. **Gerador.** `personagens.py` (no gerador anexo) desenha cada robô em um quadro de 200 × 300: cabeça com visor e olhos da estação, antena, marca na testa, torso com a roupa da estação, braços e o item na mão. Cada figura tem contorno de 5 px em `#151515`, sombreamento em dois tons e luz de contorno em `rosa-choque` no lado direito.
2. **Retratos.** Cinco SVG, um por estação, recortados em cabeça e ombros (`docs/edicao/personagens/<estacao>.svg`).
3. **Equipe.** `docs/edicao/equipe.svg`: os cinco de corpo inteiro, em linha irregular (alturas e poses diferentes), para o topo da LP.
4. **Aplicação.** Com a edição vigente, o avatar de cada cartão mostra o retrato (o canvas 3D fica oculto); o topo mostra a equipe diante do pôr do sol. Fora da vigência, nada muda.

### Quem veste o quê

| Estação | Marca | Roupa e itens |
|---|---|---|
| Líder | `ring` | terno claro com camisa rosa, chapéu panamá, charuto, corrente de ouro, arminha |
| PM & Arquiteto | `diamond` | camisa florida, óculos de aro âmbar, brinco, arminha |
| Planejador | `bar` | regata branca, boné de aba reta, corrente de ouro, porrete |
| Builders | `wave` | macacão azul-petróleo, bandana, óculos rosa, arminha |
| Governança | `tri` | colete com distintivo, boné de vigia, porrete |

## Fora do escopo

Animação, som, texto dentro das imagens, qualquer elemento do jogo, os mascotes 3D (ficam como estão fora da edição).

## Decisões do projeto aplicadas

- **Discovery 101, Gate 1:** sem personagens do jogo; critérios 1 e 7 da iniciativa.
- **Decisão do dono (19/11/2026):** os robôs precisam lembrar o estilo dos personagens; ele rejeitou o 3D com acessórios.
- **Vigência:** tudo sob `html[data-edicao="missoes"]`; o PR de encerramento apaga `docs/edicao/` inteira.
- **FLX-02:** branch única da iniciativa.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Figurino por estação | `docs/edicao/edicao-mascotes.js`, `acessoriosPorMarca` | Itens por marca | reusar como referência do desenho 2D |
| Peças do gerador | `anexos/gerador-banner/parts.py`, `tokens.py` | Cores, degradês, `grad` | reusar |
| Avatar do cartão | `docs/index.html`, `.avatar-stage` e canvas 3D | Canvas por estação | estender por CSS (ocultar o canvas, fundo com o retrato) |
| Topo em tela cheia | `docs/edicao/edicao.css`, `.hero::before` | Arte do pôr do sol | estender com a equipe em `.hero::after` |
| Ordem das estações | `docs/edicao/edicao.js`, `estacaoDoCartao` | Líder, PM, Planejador, Builders, Governança | reusar |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11**, **DUP-02**, **COD-01**, **COD-04**, **COD-05**, **SEG-07**, **TST-01**, **TST-05**, **FLX-01**, **FLX-02**.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
base=$(git merge-base HEAD origin/main)

# 1. Seis SVG válidos e reproduzíveis pelo gerador
for f in lider pm planejador builders governanca; do xmllint --noout docs/edicao/personagens/$f.svg; done
xmllint --noout docs/edicao/equipe.svg
( cd sdd/discovery/101-tema-missoes-v5/anexos/gerador-banner && python3 personagens.py /tmp/personagens-ensaio >/dev/null && for f in personagens/lider.svg personagens/pm.svg personagens/planejador.svg personagens/builders.svg personagens/governanca.svg equipe.svg; do cmp "/tmp/personagens-ensaio/$f" "../../../../../docs/edicao/$f" || exit 1; done; r=$?; rm -rf __pycache__ /tmp/personagens-ensaio; exit $r )

# 2. Sem texto, imagem embutida ou rede nas figuras
test "$(grep -c -E '<text|<image|https?://|data:' docs/edicao/equipe.svg docs/edicao/personagens/*.svg | awk -F: '{s+=$2} END{print s}')" = "0"

# 3. CSS: retrato no avatar, equipe no topo, tudo sob a marca da edição
grep -q 'personagens/lider.svg' docs/edicao/edicao.css && grep -q 'equipe.svg' docs/edicao/edicao.css
test "$(grep -E '^[^@ }/*][^{]*\{' docs/edicao/edicao.css | grep -v -c 'html\[data-edicao="missoes"\]')" = "0"

# 4. Marca de terceiros e HTML intacto fora dos marcadores
test "$(grep -rn -i -E 'gta|grand theft|vice city|pricedown|rockstar|take-two|lucia|jason' docs/edicao | wc -l | tr -d ' ')" = "0"
for f in docs/index.html docs/wiki.html; do
  diff <(git show "$base:$f") <(sed '/edicao:inicio/,/edicao:fim/d' "$f") >/dev/null || { echo "mudou fora dos marcadores: $f"; exit 1; }
done

# 5. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

Cada retrato legível a 78 px e a equipe legível no topo; nada cortado; nenhuma pose lembra a arte de capa. Peça registrada em `revisao-pi.md`.

## Status

- [x] Implementado em feat-101-06 (19/11/2026): critérios 1 a 5 com saída 0 e `forge gate` em bloqueio sem pendências.

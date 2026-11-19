# feat-101-01 — LP e wiki na Edição Missões, com vigência e retorno automático

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** M
**Depende de:** fix-101-01 (nomes do design system e `missoes.json`)

## Objetivo

Durante a edição de lançamento da 5.0.0, `forge-sdd.org` (landing page e wiki) aparece com o tema da Edição Missões. À 0h de segunda-feira, 23/11/2026, no horário de Brasília, as páginas voltam sozinhas ao layout de hoje, sem deploy e sem ação de ninguém.

## Decisões registradas (Nathan, 19/11/2026)

- A feature nasce dentro da 101, como a segunda etapa prevista no Gate 1 (site).
- A identidade segue a 101: nome "Edição Missões", sem "GTA VI" escrito, sem letreiro parecido e com o "5" como destaque próprio.
- Do design system ficam a paleta do pôr do sol, o neon e as molduras.
- O tema é temporário e parametrizado: "ficar assim só essa semana, remover na segunda-feira dia 23/11/2026 0h, de forma automática voltar ao layout anterior".
- O conteúdo e a estrutura da LP 5.0.0 não mudam.

## Modelo do módulo

**Enxuto** (ARQ-11): site estático, sem backend e sem pacote Go novo. Entram dois arquivos em `docs/edicao/`, poucas linhas marcadas em `docs/index.html` e `docs/wiki.html` e um teste no pacote que já testa a documentação.

## Como funciona

1. **Camada isolada.** Tudo da edição fica em `docs/edicao/`: `edicao.js` (parâmetros e porteiro) e `edicao.css` (tema). `docs/index.html` e `docs/wiki.html` só ganham as linhas que carregam esses dois arquivos, entre os marcadores `edicao:inicio` e `edicao:fim`.
2. **Parâmetro único.** `docs/edicao/edicao.js` é o único lugar com o identificador e a data de fim:

   ```js
   window.forgeEdicao = {
     id: "missoes",
     fim: "2026-11-23T00:00:00-03:00",
     vigente(agoraMs) { /* true enquanto agoraMs < fim; false se a data for inválida */ }
   };
   ```

   Mudar a data é editar essa linha.
3. **Porteiro.** O script roda no `<head>`, sem `defer` nem `async`, antes da primeira pintura. Com a edição vigente, marca `<html data-edicao="missoes">`. Fora da vigência, não marca nada.
4. **Tema.** Toda regra de `edicao.css` começa por `html[data-edicao="missoes"]`. A camada troca as variáveis de cor da LP e da wiki pelos tokens do design system e acrescenta: barras com linha neon no topo e no rodapé, aba ativa em chip, pôr do sol no fundo do topo da página, o selo "Edição Missões" e o "5" em destaque, na tipografia da marca.
5. **Missões.** Com a edição vigente, cada cartão da seção `#estacoes` ganha o rótulo da sua missão. O texto vem de `docs/design/edicao-missoes/missoes.json`, lido em tempo de execução. A camada não guarda cópia do texto.
6. **Pré-visualização.** `?edicao=on` e `?edicao=off` forçam o estado só naquela visita, para conferir antes e depois da virada. Qualquer outro valor é ignorado.

### Comportamento esperado

| Situação | Resultado |
|---|---|
| Antes de 23/11/2026 0h (Brasília) | `data-edicao="missoes"` no `<html>` antes da primeira pintura; tema aplicado |
| A partir de 23/11/2026 0h (Brasília) | Nenhuma marca; LP e wiki iguais às de hoje |
| JavaScript desligado | Layout de hoje |
| Data ausente ou inválida | Layout de hoje |
| `?edicao=on` ou `?edicao=off` | Força o estado só naquela visita |

A virada usa o relógio do visitante, comparado em tempo absoluto: o fuso de quem visita não muda a hora da troca.

### Missão de cada cartão

| Cartão da LP | Estação | Missão |
|---|---|---|
| 02. PM & Arquiteto | Concepção | Missão 1: o plano |
| 03. O Planejador | Planejamento | Missão 2: a rota |
| 04. Builders | Construção | Missão 3: a obra |
| 05. Governança | Governança | Missão 4: o vigia |
| 01. O Líder | Operação | Missão 5: missão cumprida |

**A confirmar no Gate 1:** a LP abre com o Líder (cartão 01), que nesta tabela fica com a Missão 5. A alternativa é deixar o Líder sem rótulo de missão.

## Fora do escopo

- Agentes rebeldes nos mascotes (feat-101-02).
- Retorno do README e remoção da camada (feat-101-03).
- README, banner e design system (fix-101-01).
- Imagem de compartilhamento e posts.
- Mudar textos, seções ou a ordem da LP; mexer no Worker ou no DNS.

## Decisões do projeto aplicadas

- **Discovery 101, Gate 1 (19/11/2026):** homenagem ao clima, sem logo, letreiro, "VI", personagens, trilha ou interface do jogo; o site é a segunda etapa.
- **Critérios 1, 2, 3 e 6 da iniciativa** (`criteria.md`).
- **Hospedagem:** `forge-sdd.org` é a pasta `docs/`, publicada pelo Worker `lp-forgesdd` (`site/wrangler.jsonc`). `docs/edicao/` sobe junto com o resto.
- **FLX-02:** branch única da iniciativa, `feat/101-tema-missoes-v5`.
- **Módulos e fluxos do CLI:** nenhum é tocado.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Tokens do tema | `docs/design/edicao-missoes/tokens.css` (hoje em `docs/design/gta-vi/`, renomeado pelo fix-101-01) | Superfícies, texto, acentos e degradês medidos | reusar |
| Variáveis de tema da LP | `docs/index.html`, bloco `:root` (linhas 13 a 38) | `--bg`, `--bg2`, `--bg3`, `--accent`, `--ok`, `--warn`, `--line`, `--f-body`, `--f-mono` | estender: a camada sobrescreve as variáveis, sem editar o bloco |
| Variáveis de tema da wiki | `docs/wiki.css`, bloco `:root` | `--bg`, `--accent`, `--text`, `--text2`, `--border` | estender, da mesma forma |
| Cores fixas fora das variáveis | `docs/index.html` | Cerca de 14 valores em hexadecimal, a maioria a cor de cada estação | reusar como estão |
| Troca de tema por data ou atributo | `docs/`, `site/` | Não existe | criar em `docs/edicao/edicao.js` |
| Texto das missões | `docs/design/edicao-missoes/missoes.json` (fix-101-01) | 5 missões e a promessa | reusar a fonte única |
| Cartões das estações | `docs/index.html`, seção `#estacoes`, lista de estações do script | 5 cartões montados por script | reusar: o rótulo entra pela camada, sem editar a lista |
| Teste que lê a documentação | `internal/scaffold/pipeline_scaffold_test.go` | `TestScaffold_DocsCitamBriefConstituicaoETmux` | estender com `TestScaffold_DocsEdicao` |
| Publicação da LP | `site/wrangler.jsonc`, `.github/workflows/deploy-lp.yml`, `docs/.assetsignore` | Worker de assets publica `docs/` | reusar |

## Dependências novas

Nenhuma. Fontes (Montserrat e JetBrains Mono) e Three.js já são carregadas pela LP.

**Dependência de entrega:** o deploy automático só funciona depois do PR #112 (`fix(lp)`), que corrige o workflow `Deploy LP`.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11:** modelo enxuto, declarado acima.
- **DUP-02:** tabela "O que já existe" preenchida; missões e tokens vêm de fonte única.
- **COD-01:** a edição vive em arquivos próprios; nada do tema fica espalhado no CSS da LP.
- **COD-04:** a camada é removível pelos marcadores; a remoção acontece na feat-101-03.
- **COD-05:** textos em pt-BR, nível padrão.
- **SEG-01:** `?edicao` só aceita `on` e `off`; data inválida desliga o tema.
- **SEG-07:** nenhuma dependência nova.
- **TST-01 e TST-05:** teste Go novo, `go test ./...` e `go vet ./...` no portão.
- **FLX-01 e FLX-02:** spec antes do código; branch única da iniciativa.
- Rules: `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
base=$(git merge-base HEAD origin/main)

# 1. Camada isolada e data em um único lugar
test -f docs/edicao/edicao.css && test -f docs/edicao/edicao.js
test "$(grep -rl '2026-11-23' docs README.md npm/README.md .github 2>/dev/null)" = "docs/edicao/edicao.js"

# 2. Marcadores casados nas duas páginas
for f in docs/index.html docs/wiki.html; do
  a=$(grep -c 'edicao:inicio' "$f"); b=$(grep -c 'edicao:fim' "$f")
  test "$a" -ge 1 && test "$a" = "$b" || { echo "marcadores: $f"; exit 1; }
done

# 3. Fora dos marcadores, as páginas não mudam; a folha da wiki não muda
for f in docs/index.html docs/wiki.html; do
  diff <(git show "$base:$f") <(sed '/edicao:inicio/,/edicao:fim/d' "$f") >/dev/null \
    || { echo "mudou fora dos marcadores: $f"; exit 1; }
done
git diff --quiet "$base" -- docs/wiki.css

# 4. O porteiro carrega no <head>, sem defer nem async
for f in docs/index.html docs/wiki.html; do
  awk '/edicao\/edicao\.js/{a=NR} /<\/head>/{b=NR} END{exit !(a && b && a<b)}' "$f"
  test "$(grep 'edicao/edicao.js' "$f" | grep -c -E 'defer|async')" = "0"
done

# 5. Vigência: ligada um segundo antes da virada, desligada na virada e com data inválida
node -e '
const vm = require("vm"), fs = require("fs");
const raiz = { setAttribute() {}, removeAttribute() {} };
const ctx = { window: {}, document: { documentElement: raiz }, location: { search: "" } };
vm.runInNewContext(fs.readFileSync("docs/edicao/edicao.js", "utf8"), ctx);
const e = ctx.window.forgeEdicao;
const antes = e.vigente(Date.parse("2026-11-22T23:59:59-03:00"));
const depois = e.vigente(Date.parse("2026-11-23T00:00:00-03:00"));
e.fim = "sem data";
const invalida = e.vigente(Date.parse("2026-11-20T12:00:00-03:00"));
if (antes !== true || depois !== false || invalida !== false) { console.error({ antes, depois, invalida }); process.exit(1); }
console.log("ok");'

# 6. Toda regra de primeiro nível da camada só vale com a marca no <html>
test "$(grep -E '^[^@ }/*][^{]*\{' docs/edicao/edicao.css | grep -v -c 'html\[data-edicao="missoes"\]')" = "0"

# 7. As missões vêm da fonte única, sem cópia na camada
grep -q 'missoes.json' docs/edicao/edicao.js
test "$(grep -c 'missão cumprida' docs/edicao/edicao.js)" = "0"

# 8. Nenhuma marca de terceiros nas páginas nem na camada
test "$(grep -rn -i -E 'gta|grand theft|vice city|pricedown|rockstar|take-two' docs/edicao docs/index.html docs/wiki.html | wc -l | tr -d ' ')" = "0"

# 9. Portão do projeto
go test ./internal/scaffold/ -run 'TestScaffold_DocsEdicao' -count=1
go test ./... && go vet ./...
```

`TestScaffold_DocsEdicao` confere, em Go, que `docs/edicao/edicao.js` tem uma única data no formato RFC 3339 e que os marcadores estão casados em `docs/index.html` e `docs/wiki.html`.

## Conferência visual (Gate 2)

- LP e wiki com `?edicao=on` e com `?edicao=off`, em 1280 px e em 390 px de largura.
- Com `?edicao=off`, as páginas são iguais às publicadas hoje.
- Texto branco sobre superfícies escuras; rótulo sobre chip sempre com contorno ou sombra escura (tabela de contraste do design system).
- Nenhuma animação nova roda com `prefers-reduced-motion: reduce`.
- Peça registrada em `sdd/discovery/101-tema-missoes-v5/revisao-pi.md`.

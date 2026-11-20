# feat-101-05 — Visual de malandro de cidade tropical nos mascotes

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** M
**Depende de:** feat-101-02 (acessórios dos mascotes)

## Objetivo

O Nathan aprovou o visual da LP e pediu que os robôs "tenham que parecer personagem do GTA". Esta task dá a cada mascote um figurino de malandro de cidade tropical (chapéu panamá, charuto, corrente e brinco de ouro, gola florida, boné de aba reta), por cima das bandanas, óculos e armas de brinquedo que já existem.

## Limite de marca (da 101, mantido)

Os robôs continuam sendo os mascotes do forge-sdd. O estilo vem de arquétipos genéricos de figurino. Nenhum personagem do jogo (nem semelhança com Lucia, Jason ou outros), nenhum logo, texto ou interface do jogo. O charuto é de brinquedo, sem fumaça, e as armas seguem de brinquedo.

## Modelo do módulo

**Enxuto** (ARQ-11): só `docs/edicao/edicao-mascotes.js`, em geometria pura do Three.js. Nenhum arquivo novo e nenhum Go.

## Figurino por estação

| Cartão | Marca | Itens |
|---|---|---|
| 01. O Líder | `ring` | `panama`, `charuto`, `corrente`, `arminha` |
| 02. PM & Arquiteto | `diamond` | `oculos`, `colarinho`, `brinco`, `arminha` |
| 03. O Planejador | `bar` | `bonereto`, `corrente`, `porrete` |
| 04. Builders | `wave` | `bandana`, `oculos`, `brinco`, `arminha` |
| 05. Governança | `tri` | `bone`, `corrente`, `porrete` |

Itens novos: `panama` (chapéu de aba larga com faixa), `charuto` (cilindro marrom com ponta em brasa, no canto da boca), `corrente` (colar de ouro com pingente em losango, sob a cabeça), `brinco` (argola dourada na orelha esquerda), `colarinho` (gola com flores), `bonereto` (boné de aba reta). Os itens antigos ficam como estão.

## Fora do escopo

Rostos, corpo, roupas fora da cabeça e do pescoço, animações novas, sons, qualquer arquivo externo.

## Decisões do projeto aplicadas

- **Discovery 101, Gate 1:** sem personagem do jogo; critérios 1 e 7 da iniciativa.
- **Decisão do dono (19/11/2026):** os robôs devem lembrar o clima dos personagens, dentro dessas regras.
- **Vigência:** os itens só existem com a marca da edição e somem com `docs/edicao/`.
- **FLX-02:** branch única da iniciativa.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Sistema de acessórios | `docs/edicao/edicao-mascotes.js`: `acessoriosPorMarca`, `vestirMascote`, criadores por tipo | Bandana, óculos, boné, arminha e porrete | estender com novos tipos |
| Boné | `criarBone` (aba curva) | Um só modelo | estender com variante de aba reta |
| Ponto de extensão no mascote | `docs/index.html`, linha marcada em `createForgeMascot3D` | Chama `vestirMascote` | reusar; o HTML não muda |
| Paleta | constante `COR` no mesmo arquivo | Tons do design system | reusar; ouro e marrom entram como materiais do brinquedo |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11**, **DUP-02**, **COD-01**, **COD-04**, **COD-05**, **SEG-07**, **TST-01**, **TST-05**, **FLX-01**, **FLX-02**.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
base=$(git merge-base HEAD origin/main)

# 1. Cada marca tem o figurino da tabela, com arma de brinquedo
node -e '
const vm = require("vm"), fs = require("fs");
const ctx = { window: { forgeEdicao: {} } };
vm.runInNewContext(fs.readFileSync("docs/edicao/edicao-mascotes.js", "utf8"), ctx);
const a = ctx.window.forgeEdicao.acessoriosPorMarca;
const esperado = { ring: ["panama","charuto","corrente","arminha"], diamond: ["oculos","colarinho","brinco","arminha"], bar: ["bonereto","corrente","porrete"], wave: ["bandana","oculos","brinco","arminha"], tri: ["bone","corrente","porrete"] };
for (const m in esperado) for (const i of esperado[m]) if (!(a[m] || []).includes(i)) { console.error("falta", m, i); process.exit(1); }
console.log("ok");'

# 2. Cada item novo tem criador e é ligado em vestirMascote
for t in panama charuto corrente brinco colarinho bonereto; do
  grep -q "\"$t\"" docs/edicao/edicao-mascotes.js || { echo "sem $t"; exit 1; }
done

# 3. Sem arquivo externo, rede ou marca de terceiros; HTML intacto fora dos marcadores
test "$(grep -c -i -E '\.(glb|gltf|obj|fbx|png|jpe?g|webp|mp3|wav)|https?://' docs/edicao/edicao-mascotes.js)" = "0"
test "$(grep -rn -i -E 'gta|grand theft|vice city|pricedown|rockstar|take-two|lucia|jason' docs/edicao | wc -l | tr -d ' ')" = "0"
for f in docs/index.html docs/wiki.html; do
  diff <(git show "$base:$f") <(sed '/edicao:inicio/,/edicao:fim/d' "$f") >/dev/null || { echo "mudou fora dos marcadores: $f"; exit 1; }
done

# 4. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

Capturas dos 5 cartões ampliadas: cada item legível, nada cobrindo os olhos e nada cortado pela borda do avatar. Peça registrada em `revisao-pi.md`.

## Status

- [x] Implementado em feat-101-05 (19/11/2026): critérios 1 a 4 com saída 0 e `forge gate` em bloqueio sem pendências.

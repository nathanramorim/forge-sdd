# feat-101-02 — Agentes rebeldes nas 5 missões

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** M
**Depende de:** feat-101-01 (camada da edição e porteiro de vigência)

## Objetivo

Enquanto a edição estiver vigente, os mascotes 3D dos agentes da landing page aparecem como "rebeldes": cada um ganha um visual da edição e uma arminha ou um porrete de brinquedo. Fora da vigência, os mascotes são exatamente os de hoje.

## Decisões registradas (Nathan, 19/11/2026)

- Pedido: "considere colocar os agentes de IA rebelde com arminhas purrete sem perder qualidade e nas fases".
- Leitura adotada: arminhas e porrete, nos 5 cartões da seção de estações (as fases). **A confirmar no Gate 1:** "purrete" é porrete?
- Os personagens são os mascotes do próprio forge-sdd. Nenhum personagem do jogo entra (critério 1 da iniciativa).

## Modelo do módulo

**Enxuto** (ARQ-11): sem pacote Go novo. Entra um arquivo em `docs/edicao/` e uma linha marcada dentro da função que já desenha o mascote.

## Como funciona

1. **Acessórios na camada da edição.** Novo `docs/edicao/edicao-mascotes.js`, carregado junto com a camada. Ele declara os acessórios de cada estação e sabe desenhá-los em geometria do Three.js, com os mesmos materiais e a mesma luz do mascote.
2. **Um ponto de extensão.** `createForgeMascot3D` ganha uma única linha, entre marcadores `edicao:inicio` e `edicao:fim`, que entrega o grupo do mascote à camada quando a edição está vigente. A função não é copiada nem reescrita.
3. **Acessórios por estação.** O contrato é uma tabela exposta em `window.forgeEdicao.acessoriosPorMarca`, indexada pela marca que cada estação já usa:

| Cartão da LP | Marca | Visual | Na mão |
|---|---|---|---|
| 01. O Líder | `ring` | bandana | arminha |
| 02. PM & Arquiteto | `diamond` | óculos escuros | arminha |
| 03. O Planejador | `bar` | bandana | porrete |
| 04. Builders | `wave` | óculos escuros | arminha |
| 05. Governança | `tri` | boné de vigia | porrete |

   Os outros dois mascotes da página (emulador e seção de governança) usam as marcas `ring` e `tri` e seguem a mesma tabela.
4. **Tom.** Arminhas e porrete em estilo de brinquedo: formas arredondadas, cores do design system, sem disparo, sem mira no visitante e sem dano ou sangue.
5. **Estados.** Os acessórios acompanham os estados que o mascote já tem (ocioso, trabalhando, concluído, bloqueado) e usam o mesmo laço de animação. Nenhum renderizador novo é criado.

### Barra de qualidade

- Silhueta legível no tamanho do cartão (78 × 86 px) e em telas de 390 px de largura.
- Curvas com pelo menos 16 segmentos; nenhuma face piscando por sobreposição.
- Acessório com no máximo 45% da altura do corpo, sem cobrir os olhos nem a marca da estação.
- A animação do cartão ativo continua fluida: sem laço extra de desenho e sem aumentar o limite de `devicePixelRatio`.

## Fora do escopo

- Personagens novos, ilustrações 2D, sons e modelos 3D externos.
- Mudar textos ou a ordem dos cartões.
- Tema, vigência e rótulos de missão (feat-101-01).

## Decisões do projeto aplicadas

- **Discovery 101, Gate 1 (19/11/2026):** homenagem ao clima, sem personagens nem interface do jogo.
- **Critérios 1 e 7 da iniciativa** (`criteria.md`).
- **Vigência da edição:** os acessórios só aparecem com `html[data-edicao="missoes"]` (feat-101-01) e somem sozinhos na virada.
- **FLX-02:** branch única da iniciativa, `feat/101-tema-missoes-v5`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Mascote 3D | `docs/index.html`, `createForgeMascot3D(mountEl, agentHex, initialState, markType)` | Corpo, olhos, sombra, marca por estação e estados | estender com um ponto de extensão marcado |
| Onde os mascotes nascem | `docs/index.html`: cartões das estações, emulador (`emuMascot`) e governança (`govMascot`) | 3 chamadas da mesma função | reusar: nenhuma chamada muda |
| Marcas das estações | `docs/index.html`, lista de estações (`ring`, `diamond`, `bar`, `wave`, `tri`) | 5 marcas | reusar como chave da tabela de acessórios |
| Three.js | `docs/index.html`, `<head>` (r128, por CDN) | Já carregada | reusar |
| Sinal de edição vigente | `window.forgeEdicao` e `html[data-edicao]` (feat-101-01) | Porteiro de data | reusar |
| Modelos, texturas ou desenhos de arma | `docs/`, `site/public/` | Não existem | criar em geometria, dentro de `docs/edicao/` |

## Dependências novas

Nenhuma. Sem modelos, texturas, imagens ou bibliotecas novas.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11:** modelo enxuto, declarado acima.
- **DUP-01 e DUP-02:** o mascote continua com uma única função dona; a edição só acrescenta acessórios.
- **COD-01:** acessórios em arquivo próprio da camada.
- **COD-04:** o ponto de extensão fica entre marcadores e sai com a camada na feat-101-03.
- **COD-05:** textos em pt-BR, nível padrão.
- **SEG-07:** nenhuma dependência nova.
- **TST-01 e TST-05:** `go test ./...` e `go vet ./...` no portão.
- **FLX-01 e FLX-02:** spec antes do código; branch única da iniciativa.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
base=$(git merge-base HEAD origin/main)

# 1. Acessórios na camada; o mascote continua com uma única função
test -f docs/edicao/edicao-mascotes.js
test "$(grep -c 'function createForgeMascot3D' docs/index.html)" = "1"
test "$(grep -c 'WebGLRenderer' docs/edicao/edicao-mascotes.js)" = "0"

# 2. Fora dos marcadores, a página não muda
diff <(git show "$base:docs/index.html") <(sed '/edicao:inicio/,/edicao:fim/d' docs/index.html) >/dev/null

# 3. As 5 marcas têm visual e arminha ou porrete (o arquivo carrega sem o Three.js)
node -e '
const vm = require("vm"), fs = require("fs");
const ctx = { window: { forgeEdicao: {} } };
vm.runInNewContext(fs.readFileSync("docs/edicao/edicao-mascotes.js", "utf8"), ctx);
const a = ctx.window.forgeEdicao.acessoriosPorMarca;
for (const m of ["ring", "diamond", "bar", "wave", "tri"]) {
  const itens = (a && a[m]) || [];
  if (itens.length < 2 || !itens.some(x => x === "arminha" || x === "porrete")) { console.error("incompleto:", m, itens); process.exit(1); }
}
console.log("ok");'

# 4. Sem arquivo externo: nada de modelo, imagem, som ou endereço de rede
test "$(grep -c -i -E '\.(glb|gltf|obj|fbx|png|jpe?g|webp|mp3|wav)|https?://' docs/edicao/edicao-mascotes.js)" = "0"

# 5. Nenhuma marca de terceiros
test "$(grep -rn -i -E 'gta|grand theft|vice city|pricedown|rockstar|take-two' docs/edicao | wc -l | tr -d ' ')" = "0"

# 6. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

- Capturas dos 5 cartões com `?edicao=on`, em 1280 px e em 390 px, comparadas com `?edicao=off`.
- Cada item da barra de qualidade conferido e anotado.
- Com `?edicao=off`, os mascotes são iguais aos de hoje.
- Peça registrada em `sdd/discovery/101-tema-missoes-v5/revisao-pi.md`.

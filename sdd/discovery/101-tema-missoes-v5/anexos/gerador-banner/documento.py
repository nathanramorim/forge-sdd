"""Gera tokens.json, tokens.css, identidade.json e o README do design system. Uso: python3 documento.py"""
import json
import os

from tokens import FONTS, GRADIENTS, PASTA, TOKENS, c, missoes

AVISO = "Conteúdo original, sem ligação com a Rockstar/Take-Two."
URL = "https://www.behance.net/gallery/224512205"


def luminancia(h):
    h = h.lstrip('#')
    r, g, b = (int(h[i:i + 2], 16) / 255 for i in (0, 2, 4))
    f = lambda v: v / 12.92 if v <= 0.03928 else ((v + 0.055) / 1.055) ** 2.4
    return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b)


def contraste(a, b):
    la, lb = sorted((luminancia(a), luminancia(b)), reverse=True)
    return (la + 0.05) / (lb + 0.05)


def escrever(nome, texto):
    with open(os.path.join(PASTA, nome), "w", encoding="utf-8") as f:
        f.write(texto)
    print("ok", nome)


def tokens_json():
    doc = {
        "nome": "Edição Missões",
        "versao": "2.0.0",
        "origem": {
            "autor": "Matías Rodríguez",
            "url": URL,
            "publicado": "2025-04-25",
            "medicao": "Cores amostradas das pranchas publicadas, em 2026-11-19. Tolerância de ±2 por canal.",
            "aviso": AVISO,
        },
        "cor": {g: {n: {"valor": v, "uso": u} for n, (v, u) in itens.items()} for g, itens in TOKENS.items()},
        "degrade": {n: {"direcao": d, "uso": u, "paradas": [{"posicao": p, "cor": cor} for p, cor in st]} for n, (d, u, st) in GRADIENTS.items()},
        "tipografia": FONTS,
    }
    return json.dumps(doc, ensure_ascii=False, indent=2) + "\n"


def tokens_css():
    css = ["/* Design system da Edição Missões (forge-sdd 5.0). Gerado por documento.py junto com tokens.json: altere tokens.py/documento.py e regenere. */", ":root {"]
    for g, itens in TOKENS.items():
        css += [f"  --missoes-{g}-{n}: {v};" for n, (v, _) in itens.items()] + [""]
    for n, (d, _, st) in GRADIENTS.items():
        ang = "180deg" if d == "vertical" else "90deg"
        css.append(f"  --missoes-degrade-{n}: linear-gradient({ang}, " + ", ".join(f"{cor} {p}%" for p, cor in st) + ");")
    css.append("")
    css += [f"  --missoes-fonte-{n}: {f['pilha']};" for n, f in FONTS.items()]
    css.append("}")
    return "\n".join(css) + "\n"


def identidade_json():
    doc = {
        "fundo": c("superficie", "preto-moldura"),
        "brilho": c("superficie", "vinho"),
        "texto": c("texto", "branco"),
        "suave": c("texto", "inativo"),
        "acento": [c("acento", "pessego"), c("acento", "coral"), c("acento", "rosa-choque")],
        "painel": c("superficie", "botao-escuro"),
        "borda": c("superficie", "painel"),
        "terminal_fundo": c("superficie", "preto-fundo"),
        "terminal_borda": c("superficie", "faixa-inativa"),
        "ok": c("texto", "ativo"),
        "fonte": FONTS["marca"]["pilha"],
        "fonte_mono": FONTS["mono"]["pilha"],
        "logo": "logo.svg.frag",
    }
    return json.dumps(doc, ensure_ascii=False, indent=2) + "\n"


def tabela_cores(grupo):
    linhas = ["| Token | Cor | Uso |", "|---|---|---|"]
    linhas += [f"| `{n}` | `{v}` | {u} |" for n, (v, u) in TOKENS[grupo].items()]
    return "\n".join(linhas)


def tabela_degrades():
    linhas = ["| Token | Direção | Paradas | Uso |", "|---|---|---|---|"]
    for n, (d, u, st) in GRADIENTS.items():
        linhas.append(f"| `{n}` | {d} | " + " → ".join(f"`{cor}`" for _, cor in st) + f" | {u} |")
    return "\n".join(linhas)


def tabela_contraste():
    t = lambda g, n: TOKENS[g][n][0]
    pares = [
        ("`branco` sobre `preto-moldura`", "#ffffff", t("superficie", "preto-moldura"), "Livre"),
        ("`branco` sobre `vinho`", "#ffffff", t("superficie", "vinho"), "Livre"),
        ("`branco` sobre `painel`", "#ffffff", t("superficie", "painel"), "Livre"),
        ("`ativo` sobre `preto-moldura`", t("texto", "ativo"), t("superficie", "preto-moldura"), "Livre"),
        ("`preto-moldura` sobre `ambar`", t("superficie", "preto-moldura"), t("acento", "ambar"), "Livre"),
        ("`aviso` sobre `painel`", t("texto", "aviso"), t("superficie", "painel"), "Livre"),
        ("`inativo` sobre `preto-moldura`", t("texto", "inativo"), t("superficie", "preto-moldura"), "Livre"),
        ("`branco` sobre `ambar`", "#ffffff", t("acento", "ambar"), "Só em negrito e com contorno ou sombra escura"),
        ("`ativo` sobre `faixa-inativa`", t("texto", "ativo"), t("superficie", "faixa-inativa"), "Decorativo"),
        ("`inativo` sobre `faixa-inativa`", t("texto", "inativo"), t("superficie", "faixa-inativa"), "Decorativo: não use para texto essencial"),
    ]
    linhas = ["| Combinação | Contraste | Leitura |", "|---|---|---|"]
    for nome, a, b, leitura in pares:
        valor = f"{contraste(a, b):.1f}".replace(".", ",")
        linhas.append(f"| {nome} | {valor}:1 | {leitura} |")
    chip = [contraste("#ffffff", cor) for _, cor in GRADIENTS["chip"][2]]
    faixa = f"{min(chip):.1f}:1 a {max(chip):.1f}:1".replace(".", ",")
    linhas.insert(10, f"| `branco` sobre `chip` | {faixa} | Só em negrito e com contorno ou sombra escura |")
    return "\n".join(linhas)


def readme():
    m = missoes()
    promessa = m["promessa"]
    missoes_md = "\n".join(["| Missão | Estação |", "|---|---|"] + [f"| {x['titulo']} | {x['estacao']} |" for x in m["missoes"]])
    fontes = "\n".join(["| Papel | Uso | Pilha |", "|---|---|---|"] + [f"| `{n}` | {f['uso']} | `{f['pilha']}` |" for n, f in FONTS.items()])
    return f"""# Design system da Edição Missões

<p align="center">
  <img src="especime.svg" alt="Espécime do design system da Edição Missões: tipografia, paleta, degradês, cores e componentes" width="100%">
</p>

Identidade visual da edição de lançamento do forge-sdd 5.0.0. É uma homenagem ao clima de pôr do sol, neon e palmeiras. A paleta e os degradês foram medidos em um conceito de interface publicado no Behance; a marca, a tipografia, as ilustrações e os textos são do próprio forge-sdd.

As regras de marca desta edição estão em `sdd/discovery/101-tema-missoes-v5/`.

## Origem e limites

| | |
|---|---|
| **Referência de cor** | [Conceito de interface de Matías Rodríguez no Behance]({URL}), publicado em 25/04/2025. {AVISO} |
| **O que foi medido** | Cores e degradês, amostrados por script nas pranchas publicadas, em 19/11/2026. Tolerância de ±2 por canal. |
| **O que é nosso** | Marca, tipografia, ilustrações, textos e as 5 missões. |
| **O que não entra** | Logotipos, letreiros, numerais, personagens, fontes e interface de jogos. |

## Princípios

1. **Pôr do sol sobre preto.** A cor vem da arte e dos degradês. A interface em si é preta e branca.
2. **Neon fino.** Uma linha de 2 px marca o limite das barras. Ela não vira borda grossa nem moldura de cartão.
3. **Branco para ler, cor para estado.** O texto é branco. Âmbar, rosa e degradê aparecem só em seleção, foco e seção ativa.
4. **Um sinal por estado.** Degradê `chip` é seção ativa, `ambar` é linha selecionada e `rosa-foco` é foco.

## As 5 missões

{missoes_md}

A promessa da edição: "{promessa}"

O texto vem de [`missoes.json`](missoes.json), a fonte única. README, site e posts usam exatamente esse texto.

## Cores

### Superfícies

{tabela_cores("superficie")}

### Texto

{tabela_cores("texto")}

### Acentos

{tabela_cores("acento")}

### Apoio

{tabela_cores("apoio")}

## Degradês

{tabela_degrades()}

## Tipografia

{fontes}

A edição usa a tipografia da marca. O destaque da versão é o "5.0" em Montserrat 800, preenchido com o degradê `assinatura`.

## Componentes

| Componente | Anatomia | Tokens |
|---|---|---|
| **Moldura** | Barra superior e inferior em preto, cada uma com cerca de 9% a 11% da altura. | `preto-moldura`, `preto-fundo` |
| **Linha neon** | Colada na borda interna de cada barra, com 2 px e um brilho suave da mesma cor. | `assinatura` |
| **Abas** | Fila de rótulos na barra. A aba ativa é um chip em degradê com rótulo branco; as inativas são só o rótulo. | `chip` |
| **Destaque da versão** | O "5.0" em degradê, na tipografia da marca, à esquerda do bloco de título. | `assinatura` |
| **Bloco de título** | Selo da edição, título em caixa alta, a promessa e um parágrafo curto, sobre um véu preto. | `preto-fundo`, `branco` |
| **Lista de opções** | Linhas escuras translúcidas. A linha ativa é âmbar sólido, com o valor em caixa alta à direita. | `ambar`, `preto-fundo` |
| **Painel de descrição** | Barra de título escura com rótulo centralizado; corpo translúcido; aviso em amarelo. | `painel`, `aviso` |
| **Fundos** | Degradê horizontal atrás das molduras; pôr do sol com sol listrado, palmeiras e silhueta de cidade. | `prancha`, `ceu`, `sol-crepusculo`, `sol-orquidea`, `sol-ambar` |
| **Topo em tela cheia** | Pôr do sol com a largura da janela atrás do título da LP, gerado por `hero.py`; a base se funde com o fundo ameixa. | `ceu`, `ameixa` |
| **Placa de neon** | Última missão em caixa alta, tubo branco com brilho rosa e coral, levemente inclinada. | `branco`, `coral`, `rosa-choque` |
| **Cartão em mosaico** | Contorno de 3 px, cantos de 6 px e sombra dura deslocada na cor do degradê assinatura. | `coral`, `rosa-choque` |

## Contraste

Medido pela fórmula da WCAG 2. O mínimo recomendado é 4,5:1 para texto comum e 3:1 para texto grande.

{tabela_contraste()}

As quatro últimas linhas ficam abaixo do recomendado. No banner e no espécime, os rótulos sobre chip e sobre âmbar levam contorno escuro.

## Arquivos

| Arquivo | Para que serve |
|---|---|
| [`tokens.json`](tokens.json) | Tokens de cor, degradê e tipografia, com uso e origem |
| [`tokens.css`](tokens.css) | Os mesmos tokens como variáveis CSS (`--missoes-*`) |
| [`missoes.json`](missoes.json) | As 5 missões e a promessa, em fonte única |
| [`identidade.json`](identidade.json) e `logo.svg.frag` | Identidade no formato do readme-studio, para gerar novos banners |
| [`especime.svg`](especime.svg) | Folha de espécime no topo desta página |
| [`../../img/banner-v5-missoes.svg`](../../img/banner-v5-missoes.svg) | Banner do README na edição (há uma cópia em PNG para o npm) |

Exemplo com as variáveis CSS:

```css
@import url("tokens.css");

.barra {{
  background: var(--missoes-superficie-preto-moldura);
  border-bottom: 2px solid;
  border-image: var(--missoes-degrade-assinatura) 1;
}}

.aba[aria-selected="true"] {{
  background: var(--missoes-degrade-chip);
  color: var(--missoes-texto-branco);
  font-family: var(--missoes-fonte-marca);
}}
```

## O que evitar

- Escrever nomes de jogos ou de empresas de terceiros em banner, README ou site.
- Imitar logotipos, letreiros ou numerais, ou usar personagens, fontes e interface de jogos.
- Engrossar a linha neon ou usá-la como borda de cartão.
- Colorir texto corrido: cor é para estado.
"""


if __name__ == "__main__":
    escrever("tokens.json", tokens_json())
    escrever("tokens.css", tokens_css())
    escrever("identidade.json", identidade_json())
    escrever("README.md", readme())

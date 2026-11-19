"""Fonte dos tokens do design system da Edição Missões (cores medidas nas pranchas de referência)."""
import json
import os

RAIZ = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", "..", "..", ".."))
PASTA = os.path.join(RAIZ, "docs", "design", "edicao-missoes")

TOKENS = {
    "superficie": {
        "preto-moldura": ("#151515", "Barras superior e inferior das molduras"),
        "preto-fundo": ("#060506", "Fundo interno das molduras, atrás da arte"),
        "preto-lateral": ("#141414", "Painel lateral escuro"),
        "botao-escuro": ("#232123", "Botão escuro sobre a barra"),
        "disco": ("#333334", "Disco de ícone"),
        "grafite": ("#343434", "Contorno escuro de texto grande"),
        "vinho": ("#6b193c", "Fundo de página de especificação"),
        "ameixa": ("#451b3b", "Sombras e silhuetas sobre o pôr do sol"),
        "painel": ("#563f5e", "Painel de descrição, translúcido sobre a arte"),
        "faixa-inativa": ("#775882", "Item de menu inativo, em faixa translúcida"),
    },
    "texto": {
        "branco": ("#ffffff", "Títulos e corpo sobre superfícies escuras"),
        "inativo": ("#a48cad", "Rótulo de item inativo"),
        "ativo": ("#dfa351", "Rótulo do item de navegação ativo"),
        "aviso": ("#e6d364", "Linha de aviso no painel de descrição"),
    },
    "acento": {
        "pessego": ("#ffaa6e", "Início do degradê assinatura"),
        "coral": ("#fe818d", "Meio do degradê assinatura"),
        "rosa-choque": ("#f061af", "Fim do degradê assinatura"),
        "orquidea": ("#e446c4", "Topo do sol orquídea"),
        "violeta": ("#824cc8", "Transição do sol crepúsculo"),
        "indigo": ("#4053c6", "Topo do sol crepúsculo"),
        "ambar": ("#e28642", "Linha ativa das listas e sol âmbar"),
        "rosa-foco": ("#fe7694", "Foco e contorno do item selecionado"),
        "laranja-quente": ("#dd9a48", "Realce quente em ícones"),
    },
    "apoio": {
        "salvia": ("#87a77c", "Formas orgânicas de fundo"),
        "salvia-escura": ("#677c5e", "Bloco de apoio"),
        "lavanda": ("#9396c0", "Formas claras de fundo"),
        "malva": ("#8774a2", "Meio das formas de fundo"),
        "azul-baia": ("#5c7caf", "Formas azuis de fundo"),
        "azul-claro": ("#769ac6", "Realce azul"),
        "azul-noite": ("#434a68", "Elemento apagado"),
        "lilas": ("#a485aa", "Fundo lilás, topo"),
        "rosa-antigo": ("#c283ac", "Fundo lilás, base"),
        "mar": ("#5fa2b9", "Mar"),
        "mar-fundo": ("#5397ac", "Mar profundo"),
        "flamingo": ("#be50db", "Realce violeta vivo"),
    },
}

GRADIENTS = {
    "assinatura": ("horizontal", "Linha neon das molduras e destaque do 5.0", [(0, "#ffaa6e"), (20, "#fda96a"), (35, "#fd967d"), (51, "#fe818d"), (66, "#fd749b"), (82, "#f668aa"), (100, "#f061af")]),
    "chip": ("horizontal", "Aba ou seção ativa", [(0, "#f7a164"), (25, "#f88b7a"), (50, "#f77191"), (75, "#f05e9e"), (100, "#e850af")]),
    "prancha": ("horizontal", "Fundo atrás das molduras e cabeçalhos", [(0, "#fea96a"), (25, "#ff927f"), (37, "#ff818c"), (50, "#f87795"), (62, "#e57999"), (75, "#cd7c9e"), (100, "#a780a6")]),
    "ceu": ("vertical", "Céu do pôr do sol", [(0, "#5764b0"), (30, "#845d9e"), (62, "#ca84a9"), (82, "#e9949c"), (100, "#ffb06d")]),
    "sol-crepusculo": ("vertical", "Sol 1 da paleta", [(0, "#4053c6"), (12, "#824cc8"), (25, "#a849c7"), (37, "#e747c9"), (50, "#f161ae"), (62, "#f66e9f"), (75, "#f0798f"), (87, "#d16f58"), (100, "#cc7743")]),
    "sol-orquidea": ("vertical", "Sol 2 da paleta", [(0, "#e446c4"), (37, "#ee4fbc"), (50, "#f35caf"), (75, "#f269a5"), (100, "#e8698e")]),
    "sol-ambar": ("vertical", "Sol 3 da paleta", [(0, "#de8640"), (25, "#e28642"), (62, "#e1884e"), (100, "#cd7745")]),
}

FONTS = {
    "marca": {
        "uso": "Marca, títulos, rótulos e corpo. Títulos em caixa alta espaçada, peso 700 a 800; corpo em peso 500 a 600",
        "pilha": "Montserrat, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif",
    },
    "mono": {
        "uso": "Comandos de terminal",
        "pilha": "'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
    },
}


def c(group, name):
    return TOKENS[group][name][0]


def missoes():
    with open(os.path.join(PASTA, "missoes.json"), encoding="utf-8") as f:
        return json.load(f)

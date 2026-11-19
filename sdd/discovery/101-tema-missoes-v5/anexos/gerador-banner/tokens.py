"""Fonte única dos tokens do design system GTA VI (medidos nas pranchas do conceito)."""

TOKENS = {
    "superficie": {
        "preto-moldura": ("#151515", "Barras superior e inferior das molduras de tela"),
        "preto-fundo": ("#060506", "Fundo interno das molduras (atrás da arte)"),
        "preto-lateral": ("#141414", "Lateral do menu de pausa"),
        "tecla": ("#232123", "Teclas LB/RB e botões circulares"),
        "disco": ("#333334", "Disco dos ícones de vida e energia"),
        "grafite": ("#343434", "Contorno do letreiro de encerramento"),
        "vinho": ("#6b193c", "Fundo da prancha de especificação"),
        "ameixa": ("#451b3b", "Sombras e silhuetas sobre o pôr do sol"),
        "painel": ("#563f5e", "Painel de descrição (translúcido sobre a arte)"),
        "faixa-inativa": ("#775882", "Item de menu inativo (faixa translúcida)"),
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
        "rosa-vida": ("#fe7694", "Vida, foco e contorno do item selecionado"),
        "laranja-energia": ("#dd9a48", "Energia / estamina"),
    },
    "apoio": {
        "salvia": ("#87a77c", "Ondas da prancha de componentes"),
        "salvia-escura": ("#677c5e", "Slots de ação"),
        "lavanda": ("#9396c0", "Ondas claras"),
        "malva": ("#8774a2", "Meio das ondas"),
        "azul-baia": ("#5c7caf", "Ondas azuis"),
        "azul-claro": ("#769ac6", "Realce azul"),
        "estrela-apagada": ("#434a68", "Estrela de procurado apagada"),
        "lilas": ("#a485aa", "Fundo da prancha de encerramento (topo)"),
        "rosa-antigo": ("#c283ac", "Fundo da prancha de encerramento (base)"),
        "mar": ("#5fa2b9", "Mar"),
        "mar-fundo": ("#5397ac", "Mar profundo"),
        "flamingo": ("#be50db", "Flamingos"),
    },
}

GRADIENTS = {
    "assinatura": ("horizontal", "Linha neon das molduras", [(0, "#ffaa6e"), (20, "#fda96a"), (35, "#fd967d"), (51, "#fe818d"), (66, "#fd749b"), (82, "#f668aa"), (100, "#f061af")]),
    "chip": ("horizontal", "Aba ou seção ativa", [(0, "#f7a164"), (25, "#f88b7a"), (50, "#f77191"), (75, "#f05e9e"), (100, "#e850af")]),
    "prancha": ("horizontal", "Fundo atrás das molduras e cabeçalhos", [(0, "#fea96a"), (25, "#ff927f"), (37, "#ff818c"), (50, "#f87795"), (62, "#e57999"), (75, "#cd7c9e"), (100, "#a780a6")]),
    "ceu": ("vertical", "Céu (capa no topo, tela de ajustes na base)", [(0, "#5764b0"), (30, "#845d9e"), (62, "#ca84a9"), (82, "#e9949c"), (100, "#ffb06d")]),
    "sol-crepusculo": ("vertical", "Sol 1 da paleta", [(0, "#4053c6"), (12, "#824cc8"), (25, "#a849c7"), (37, "#e747c9"), (50, "#f161ae"), (62, "#f66e9f"), (75, "#f0798f"), (87, "#d16f58"), (100, "#cc7743")]),
    "sol-orquidea": ("vertical", "Sol 2 da paleta", [(0, "#e446c4"), (37, "#ee4fbc"), (50, "#f35caf"), (75, "#f269a5"), (100, "#e8698e")]),
    "sol-ambar": ("vertical", "Sol 3 da paleta", [(0, "#de8640"), (25, "#e28642"), (62, "#e1884e"), (100, "#cd7745")]),
}

FONTS = {
    "display": {
        "conceito": "Pricedown",
        "uso": "Botões, abas, menus em jogo e dinheiro",
        "pilha": "Pricedown, 'Arial Black', Impact, 'Helvetica Neue', sans-serif",
    },
    "titulo-corpo": {
        "conceito": "GTAArtDeco",
        "uso": "Títulos (caixa alta, espaçados) e corpo",
        "pilha": "GTAArtDeco, 'Avenir Next', Avenir, Futura, 'Century Gothic', Montserrat, 'Segoe UI', system-ui, -apple-system, Roboto, sans-serif",
    },
    "rotulo": {
        "conceito": "Serifa egípcia (observada na prancha, não declarada)",
        "uso": "Rótulos das seções de especificação",
        "pilha": "'Zilla Slab', Rockwell, 'Roboto Slab', Georgia, serif",
    },
    "script": {
        "conceito": "Script de pincel (observado, não declarado)",
        "uso": "Título de cada prancha, em branco sobre onda roxa",
        "pilha": "SignPainter, 'Brush Script MT', 'Segoe Script', cursive",
    },
    "mono": {
        "conceito": "Não existe no conceito; herdado do forge-sdd",
        "uso": "Comandos de terminal",
        "pilha": "'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
    },
}

def c(group, name):
    return TOKENS[group][name][0]

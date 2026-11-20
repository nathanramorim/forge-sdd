"""Robôs da Edição Missões em arte 2D de tela de carregamento. Uso: python3 personagens.py <pasta-de-saida>"""
import os
import sys

from parts import grad
from tokens import c

CONTORNO = "#151515"
OURO = "#f2b84b"
ROSA = c("acento", "rosa-choque")
VISOR = "#0b0b10"

ROBOS = {
    "lider": {"nome": "Líder", "cor": "#f2a65a", "escuro": "#b8651f", "claro": "#ffd9a8", "marca": "ring", "olhos": "barras", "roupa": "terno", "cabeca": ["panama", "charuto"], "peito": ["corrente"], "mao": "arminha"},
    "pm": {"nome": "PM & Arquiteto", "cor": "#8aa4ff", "escuro": "#4560c9", "claro": "#d3deff", "marca": "diamond", "olhos": "arcos", "roupa": "florida", "cabeca": ["oculos", "brinco"], "peito": [], "mao": "arminha", "quadril": True},
    "planejador": {"nome": "Planejador", "cor": "#c9a7ff", "escuro": "#8a5fd1", "claro": "#ecdcff", "marca": "bar", "olhos": "barras", "roupa": "regata", "cabeca": ["bonereto"], "peito": ["corrente"], "mao": "porrete"},
    "builders": {"nome": "Builders", "cor": "#4fd1b5", "escuro": "#1f8f7b", "claro": "#b6f2e3", "marca": "wave", "olhos": "barras", "roupa": "macacao", "cabeca": ["bandana", "oculos", "brinco"], "peito": [], "mao": "arminha", "quadril": True},
    "governanca": {"nome": "Governança", "cor": "#ff8a5c", "escuro": "#c24a1e", "claro": "#ffcdb5", "marca": "tri", "olhos": "crescentes", "roupa": "colete", "cabeca": ["bone"], "peito": ["corrente"], "mao": "porrete"},
}

ORDEM_EQUIPE = ["planejador", "pm", "lider", "builders", "governanca"]


def p(d, fill="none", stroke=CONTORNO, sw=5, extra=""):
    return f'<path d="{d}" fill="{fill}" stroke="{stroke}" stroke-width="{sw}" stroke-linejoin="round" stroke-linecap="round" {extra}/>'


def membro(pontos, cor, largura=24):
    d = "M" + " L".join(f"{x},{y}" for x, y in pontos)
    return (f'<path d="{d}" fill="none" stroke="{CONTORNO}" stroke-width="{largura + 10}" stroke-linejoin="round" stroke-linecap="round"/>'
            f'<path d="{d}" fill="none" stroke="{cor}" stroke-width="{largura}" stroke-linejoin="round" stroke-linecap="round"/>')


def mao(x, y, cor):
    return f'<circle cx="{x}" cy="{y}" r="15" fill="{cor}" stroke="{CONTORNO}" stroke-width="5"/>'


def arminha(x, y, escuro):
    return (f'<g transform="translate({x} {y}) rotate(-62)">'
            f'<rect x="-10" y="-14" width="52" height="24" rx="10" fill="{c("acento", "rosa-choque")}" stroke="{CONTORNO}" stroke-width="5"/>'
            f'<rect x="38" y="-10" width="14" height="16" rx="6" fill="#ffe27a" stroke="{CONTORNO}" stroke-width="5"/>'
            f'<rect x="2" y="-30" width="20" height="18" rx="8" fill="#4053c6" stroke="{CONTORNO}" stroke-width="5"/>'
            f'<rect x="-8" y="6" width="16" height="26" rx="6" fill="{escuro}" stroke="{CONTORNO}" stroke-width="5"/></g>')


def porrete(x, y):
    return (f'<g transform="translate({x} {y}) rotate(18)">'
            f'<rect x="-9" y="-92" width="34" height="84" rx="16" fill="{c("acento", "ambar")}" stroke="{CONTORNO}" stroke-width="5"/>'
            f'<rect x="-9" y="-62" width="34" height="9" fill="#ffe27a" stroke="{CONTORNO}" stroke-width="4"/>'
            f'<rect x="-2" y="-12" width="20" height="40" rx="8" fill="{c("superficie", "botao-escuro")}" stroke="{CONTORNO}" stroke-width="5"/></g>')


def roupa(tipo, r):
    cor, esc, cla = r["cor"], r["escuro"], r["claro"]
    corpo = "M34,300 L42,180 Q100,146 158,180 L166,300 Z"
    o = []
    if tipo == "terno":
        o.append(p(corpo, "#f6dfe6"))
        o.append(p("M100,168 L78,214 L100,300 L122,214 Z", ROSA, sw=4))
        o.append(p("M60,176 L100,168 L86,238 L50,300 L42,180 Z", "#fff3f7", sw=4))
        o.append(p("M140,176 L100,168 L114,238 L150,300 L158,180 Z", "#fff3f7", sw=4))
    elif tipo == "florida":
        o.append(p(corpo, c("acento", "coral")))
        for i, (x, y, k) in enumerate([(60, 214, 0), (84, 250, 1), (122, 226, 2), (142, 266, 3), (66, 278, 2), (108, 276, 0), (130, 196, 1), (88, 200, 3)]):
            col = ["#ffe27a", "#ffffff", "#824cc8", "#ffaa6e"][k]
            o.append(f'<circle cx="{x}" cy="{y}" r="8" fill="{col}" stroke="{CONTORNO}" stroke-width="3"/>')
        o.append(p("M100,168 L70,176 L90,214 Z", "#d94d63", sw=4))
        o.append(p("M100,168 L130,176 L110,214 Z", "#d94d63", sw=4))
    elif tipo == "regata":
        o.append(p("M34,300 L48,186 Q100,156 152,186 L166,300 Z", "#f4f4f4"))
        o.append(p("M70,176 Q100,214 130,176", cor, sw=4))
        o.append(p("M58,184 L74,176 M142,184 L126,176", CONTORNO, sw=4))
    elif tipo == "macacao":
        o.append(p(corpo, "#ffaa6e"))
        o.append(p("M50,300 L56,214 L144,214 L150,300 Z", esc, sw=5))
        o.append(p("M62,214 L58,172 M138,214 L142,172", esc, sw=14))
        o.append(f'<rect x="84" y="236" width="32" height="26" rx="5" fill="{cor}" stroke="{CONTORNO}" stroke-width="4"/>')
    elif tipo == "colete":
        o.append(p(corpo, "#2a2a35"))
        o.append(p("M100,168 L100,300", "#151515", sw=4))
        o.append(p("M70,180 L88,300 M130,180 L112,300", "#3b3b4a", sw=4))
        o.append(p("M62,206 L90,206 L90,246 Q76,262 62,246 Z", OURO, sw=4))
    return "".join(o)


def cabeca_acessorios(r):
    o = []
    for item in r["cabeca"]:
        if item == "panama":
            o.append(f'<ellipse cx="100" cy="42" rx="84" ry="15" fill="#ffaa6e" stroke="{CONTORNO}" stroke-width="5"/>')
            o.append(p("M58,40 L62,6 Q100,-8 138,6 L142,40 Z", "#ffbf8a"))
            o.append(f'<rect x="60" y="24" width="80" height="14" fill="{ROSA}" stroke="{CONTORNO}" stroke-width="4"/>')
        elif item == "bonereto":
            o.append(p("M40,62 Q40,6 100,4 Q160,6 160,62 Z", "#232123"))
            o.append(f'<ellipse cx="100" cy="64" rx="70" ry="13" fill="#ffaa6e" stroke="{CONTORNO}" stroke-width="5"/>')
            o.append(p("M44,52 Q100,40 156,52", "#824cc8", sw=7))
        elif item == "bone":
            o.append(p("M42,60 Q44,8 100,6 Q156,8 158,60 Z", "#3b3b4a"))
            o.append(f'<ellipse cx="100" cy="62" rx="62" ry="10" fill="{ROSA}" stroke="{CONTORNO}" stroke-width="5"/>')
            o.append(p("M100,24 L110,36 L100,48 L90,36 Z", OURO, sw=3))
        elif item == "bandana":
            o.append(p("M40,70 Q100,40 160,70", "none", CONTORNO, 17))
            o.append(p("M40,70 Q100,40 160,70", "none", "#824cc8", 10))
            o.append(f'<circle cx="38" cy="72" r="9" fill="#824cc8" stroke="{CONTORNO}" stroke-width="4"/>')
            o.append(p("M34,78 L22,100", "#ffaa6e", sw=8))
            o.append(p("M40,80 L36,106", "#824cc8", sw=8))
        elif item == "oculos":
            cor = "#ffaa6e" if r["marca"] == "diamond" else ROSA
            for x in (76, 124):
                o.append(f'<ellipse cx="{x}" cy="99" rx="23" ry="19" fill="#ffffff" fill-opacity="0.08" stroke="{cor}" stroke-width="5"/>')
            o.append(p("M99,97 L101,97", cor, sw=6))
        elif item == "brinco":
            o.append(f'<circle cx="34" cy="124" r="8" fill="none" stroke="{OURO}" stroke-width="5"/>')
        elif item == "charuto":
            o.append(f'<g transform="translate(86 140) rotate(24)"><rect x="-44" y="-5" width="46" height="11" rx="5" fill="#7a4a2b" stroke="{CONTORNO}" stroke-width="4"/>'
                     f'<rect x="-12" y="-6" width="7" height="13" fill="{OURO}"/><circle cx="-46" cy="0" r="7" fill="#ff6a3d" stroke="{CONTORNO}" stroke-width="3"/></g>')
    return "".join(o)


def marca(tipo, x, y):
    f = f'fill="{CONTORNO}" fill-opacity="0.55"'
    if tipo == "ring":
        return f'<circle cx="{x}" cy="{y}" r="9" fill="none" stroke="{CONTORNO}" stroke-opacity="0.55" stroke-width="5"/>'
    if tipo == "diamond":
        return f'<path d="M{x},{y - 11} L{x + 11},{y} L{x},{y + 11} L{x - 11},{y} Z" {f}/>'
    if tipo == "bar":
        return f'<rect x="{x - 12}" y="{y - 4}" width="24" height="8" rx="3" {f}/>'
    if tipo == "wave":
        return f'<circle cx="{x}" cy="{y}" r="8" {f}/>'
    return f'<path d="M{x},{y - 11} L{x + 12},{y + 9} L{x - 12},{y + 9} Z" {f}/>'


def olhos(tipo):
    b = '#d9e6ff'
    if tipo == "barras":
        return f'<rect x="72" y="86" width="9" height="26" rx="4" fill="{b}"/><rect x="119" y="86" width="9" height="26" rx="4" fill="{b}"/>'
    if tipo == "arcos":
        return f'<path d="M64,106 Q76,88 88,106 M112,106 Q124,88 136,106" fill="none" stroke="{b}" stroke-width="7" stroke-linecap="round"/>'
    return f'<path d="M84,88 A14 14 0 1 0 84,112 A10 10 0 1 1 84,88 Z" fill="{b}"/><path d="M116,88 A14 14 0 1 1 116,112 A10 10 0 1 0 116,88 Z" fill="{b}"/>'


def figura(chave, uid):
    r = ROBOS[chave]
    cor, esc, cla = r["cor"], r["escuro"], r["claro"]
    o = []
    o.append(f'<radialGradient id="{uid}-g" cx="0.34" cy="0.28" r="0.85"><stop offset="0" stop-color="{cla}"/><stop offset="0.45" stop-color="{cor}"/><stop offset="1" stop-color="{esc}"/></radialGradient>')
    defs = "".join(o)
    o = []
    # braço esquerdo (do espectador) e roupa
    if r.get("quadril"):
        o.append(membro([(48, 188), (18, 224), (52, 250)], esc))
        o.append(mao(52, 250, cor))
    else:
        o.append(membro([(46, 190), (32, 240)], esc))
        o.append(mao(32, 254, cor))
    o.append(roupa(r["roupa"], r))
    o.append(f'<rect x="80" y="138" width="40" height="30" rx="8" fill="{esc}" stroke="{CONTORNO}" stroke-width="5"/>')
    for item in r["peito"]:
        if item == "corrente":
            o.append(p("M72,166 Q100,214 128,166", "none", OURO, 6))
            o.append(p("M100,204 L108,218 L100,232 L92,218 Z", OURO, sw=4))
    # cabeça
    o.append(f'<ellipse cx="36" cy="100" rx="11" ry="19" fill="{esc}" stroke="{CONTORNO}" stroke-width="5"/>')
    o.append(f'<ellipse cx="164" cy="100" rx="11" ry="19" fill="{esc}" stroke="{CONTORNO}" stroke-width="5"/>')
    o.append(f'<circle cx="100" cy="94" r="64" fill="url(#{uid}-g)" stroke="{CONTORNO}" stroke-width="5"/>')
    o.append(f'<ellipse cx="76" cy="58" rx="22" ry="11" transform="rotate(-24 76 58)" fill="#ffffff" fill-opacity="0.38"/>')
    o.append(f'<path d="M146,48 A64 64 0 0 1 154,128" fill="none" stroke="{ROSA}" stroke-width="7" stroke-linecap="round" stroke-opacity="0.95"/>')
    o.append(marca(r["marca"], 100, 50))
    o.append(f'<line x1="100" y1="30" x2="100" y2="12" stroke="{CONTORNO}" stroke-width="5" stroke-linecap="round"/>')
    o.append(f'<circle cx="100" cy="8" r="9" fill="{cla}" stroke="{CONTORNO}" stroke-width="4"/>')
    o.append(f'<rect x="48" y="74" width="104" height="50" rx="25" fill="{VISOR}" stroke="{CONTORNO}" stroke-width="5"/>')
    o.append(olhos(r["olhos"]))
    o.append(f'<path d="M62,82 Q100,76 138,82" fill="none" stroke="#ffffff" stroke-opacity="0.18" stroke-width="3" stroke-linecap="round"/>')
    o.append(cabeca_acessorios(r))
    # braço direito com o item
    if r["mao"] == "arminha":
        o.append(membro([(152, 190), (178, 222)], esc))
        o.append(mao(178, 232, cor))
        o.append(arminha(178, 232, esc))
    else:
        o.append(membro([(154, 190), (176, 226)], esc))
        o.append(porrete(176, 238))
        o.append(mao(176, 238, cor))
    return defs, "".join(o)


def retrato(chave):
    defs, corpo = figura(chave, chave)
    fundo = grad("fundo", "ceu")
    return (f'<svg xmlns="http://www.w3.org/2000/svg" width="200" height="220" viewBox="0 -4 200 220" aria-hidden="true">'
            f'<defs>{defs}{fundo}<radialGradient id="halo" cx="0.5" cy="0.45" r="0.6"><stop offset="0" stop-color="#ffb06d" stop-opacity="0.7"/><stop offset="1" stop-color="#ffb06d" stop-opacity="0"/></radialGradient></defs>'
            f'<rect x="0" y="-4" width="200" height="220" fill="url(#fundo)"/><rect x="0" y="-4" width="200" height="220" fill="url(#halo)"/>{corpo}</svg>\n')


def equipe():
    largura, altura = 1100, 330
    escalas = {"planejador": 0.96, "pm": 1.0, "lider": 1.1, "builders": 1.0, "governanca": 0.96}
    xs = {"planejador": 40, "pm": 250, "lider": 450, "builders": 690, "governanca": 880}
    todos_defs, todos = [], []
    for chave in ORDEM_EQUIPE:
        defs, corpo = figura(chave, "e-" + chave)
        todos_defs.append(defs)
        k = escalas[chave]
        y = altura - 300 * k
        todos.append(f'<ellipse cx="{xs[chave] + 100 * k:.0f}" cy="{altura - 6}" rx="{80 * k:.0f}" ry="9" fill="#151515" fill-opacity="0.45"/>'
                     f'<g transform="translate({xs[chave]} {y:.0f}) scale({k})">{corpo}</g>')
    return (f'<svg xmlns="http://www.w3.org/2000/svg" width="{largura}" height="{altura}" viewBox="0 0 {largura} {altura}" aria-hidden="true">'
            f'<defs>{"".join(todos_defs)}</defs>{"".join(todos)}</svg>\n')


def gerar(pasta):
    os.makedirs(os.path.join(pasta, "personagens"), exist_ok=True)
    for chave in ROBOS:
        with open(os.path.join(pasta, "personagens", f"{chave}.svg"), "w", encoding="utf-8") as f:
            f.write(retrato(chave))
    with open(os.path.join(pasta, "equipe.svg"), "w", encoding="utf-8") as f:
        f.write(equipe())


if __name__ == "__main__":
    gerar(sys.argv[1])
    print("ok", sys.argv[1])

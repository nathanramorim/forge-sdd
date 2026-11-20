"""Folha de espécime do design system da Edição Missões. Uso: python3 specimen.py <saida.svg>"""
import sys

from parts import CONTORNO, FONT, MONO, grad, neon, palm, striped_sun
from tokens import GRADIENTS, TOKENS, c, missoes, titulos_missoes

W, H = 1280, 1330
WINE = c("superficie", "vinho")
SOFT = "#e7c9d5"


def label(x, y, text, size=20):
    return f'<text x="{x}" y="{y}" font-family="{FONT}" font-weight="700" font-size="{size}" fill="#ffffff">{text}</text>'


def note(x, y, text, anchor="start", size=12):
    return f'<text x="{x}" y="{y}" text-anchor="{anchor}" font-family="{FONT}" font-weight="500" font-size="{size}" fill="{SOFT}">{text}</text>'


def build():
    m = missoes()
    titulos = titulos_missoes()
    o = []
    o.append(f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" role="img" '
             f'aria-label="Espécime do design system da {m["edicao"]} do forge-sdd: tipografia, paleta, degradês, cores e componentes.">')
    o.append(f'<title>Design system da {m["edicao"]} — espécime</title>')
    o.append('<defs>')
    for gid, name in [("gSig", "assinatura"), ("gChip", "chip"), ("gPr", "prancha"), ("gSky", "ceu"), ("gS1", "sol-crepusculo"), ("gS2", "sol-orquidea"), ("gS3", "sol-ambar")]:
        o.append(grad(gid, name))
    o.append(grad("gSkyH", "ceu", horizontal=True))
    o.append('<filter id="fGlow" x="-5%" y="-400%" width="110%" height="900%"><feGaussianBlur stdDeviation="3"/></filter>')
    o.append('<clipPath id="cHead"><rect x="0" y="0" width="1280" height="150"/></clipPath>')
    o.append('<clipPath id="cBody"><rect x="40" y="996" width="1200" height="222"/></clipPath>')
    o.append('</defs>')
    o.append(f'<rect width="{W}" height="{H}" fill="{WINE}"/>')

    # cabeçalho
    o.append('<g clip-path="url(#cHead)">')
    o.append('<rect width="1280" height="150" fill="url(#gPr)"/>')
    o.append(palm(1088, 156, 118, -16, 0.86, c("superficie", "ameixa"), 0.42))
    o.append(palm(1190, 156, 150, 14, 1.05, c("superficie", "ameixa"), 0.5))
    o.append('</g>')
    o.append(f'<text x="40" y="84" font-family="{FONT}" font-weight="800" font-size="40" letter-spacing="0.5" fill="#ffffff">DESIGN SYSTEM · {m["edicao"].upper()}</text>')
    o.append(f'<text x="40" y="116" font-family="{FONT}" font-weight="600" font-size="15" fill="#ffffff">forge-sdd 5.0 · cores e degradês medidos em um conceito de interface publicado no Behance (Matías Rodríguez, 2025)</text>')

    # tipografia
    o.append(label(40, 200, "Tipografia"))
    o.append(f'<text x="40" y="246" font-family="{FONT}" font-weight="700" font-size="22" letter-spacing="1.8" fill="#ffffff">SEU TIME DE IA, COM MÉTODO.</text>')
    o.append(note(40, 268, "Títulos · Montserrat 700 a 800, em caixa alta espaçada"))
    o.append(f'<text x="40" y="312" font-family="{FONT}" font-weight="600" font-size="16" fill="#ffffff">{m["promessa"]}</text>')
    o.append(f'<text x="40" y="336" font-family="{FONT}" font-weight="500" font-size="15" fill="#ffffff">Do plano ao PR, sem repetir instruções.</text>')
    o.append(note(40, 358, "Corpo · Montserrat 500 a 600"))
    o.append(f'<text x="40" y="400" font-family="{MONO}" font-size="14"><tspan fill="{c("texto", "ativo")}" font-weight="700">$</tspan><tspan dx="8" fill="#ffffff">npm i -g @nathanramorim/forge-sdd</tspan></text>')
    o.append(note(40, 422, "Comandos · JetBrains Mono"))
    o.append(f'<text x="38" y="494" font-family="{FONT}" font-weight="800" font-size="64" letter-spacing="-2" fill="url(#gSig)">5.0</text>')
    o.append(note(160, 474, "Destaque da versão · o 5.0 em degradê assinatura,"))
    o.append(note(160, 492, "na tipografia da marca"))

    # paleta e degradês
    o.append(label(700, 200, "Paleta"))
    sois = [("gS1", "crepúsculo", "#4053c6 → #f161ae → #cc7743"), ("gS2", "orquídea", "#e446c4 → #f35caf → #e8698e"), ("gS3", "âmbar", "#de8640 → #e28642 → #cd7745")]
    for i, (gid, nome, faixa) in enumerate(sois):
        cx = 770 + i * 186
        o.append(striped_sun(f"mS{i}", cx, 270, 46, f"url(#{gid})"))
        o.append(f'<text x="{cx}" y="340" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="13" fill="#ffffff">sol {nome}</text>')
        o.append(f'<text x="{cx}" y="357" text-anchor="middle" font-family="{MONO}" font-size="10" fill="{SOFT}">{faixa}</text>')
    barras = [("gSig", "assinatura", "linha neon e destaque do 5.0"), ("gChip", "chip", "aba ou seção ativa"), ("gPr", "prancha", "fundo atrás das molduras"), ("gSkyH", "ceu", "vertical na arte; aqui deitado")]
    for i, (gid, nome, uso) in enumerate(barras):
        y = 384 + i * 36
        paradas = GRADIENTS[nome][2]
        faixa = " → ".join(col for _, col in (paradas[0], paradas[len(paradas) // 2], paradas[-1]))
        o.append(f'<rect x="700" y="{y}" width="230" height="18" rx="3" fill="url(#{gid})"/>')
        o.append(f'<text x="942" y="{y + 13}" font-family="{FONT}" font-size="12" fill="#ffffff"><tspan font-weight="700">{nome}</tspan><tspan fill="{SOFT}"> · {uso}</tspan></text>')
        o.append(f'<text x="1240" y="{y + 30}" text-anchor="end" font-family="{MONO}" font-size="10" fill="{SOFT}">{faixa}</text>')

    # cores
    o.append(label(40, 566, "Cores"))
    for r, (titulo, chave) in enumerate([("Superfícies", "superficie"), ("Texto", "texto"), ("Acentos", "acento"), ("Apoio", "apoio")]):
        y = 590 + r * 86
        o.append(f'<text x="40" y="{y + 26}" font-family="{FONT}" font-weight="600" font-size="14" fill="{SOFT}">{titulo}</text>')
        for i, (nome, (hexv, _)) in enumerate(TOKENS[chave].items()):
            x = 150 + i * 91
            o.append(f'<rect x="{x}" y="{y}" width="83" height="40" rx="4" fill="{hexv}" stroke="#ffffff" stroke-opacity="0.22"/>')
            o.append(f'<text x="{x}" y="{y + 55}" font-family="{FONT}" font-weight="600" font-size="10.5" fill="#ffffff">{nome}</text>')
            o.append(f'<text x="{x}" y="{y + 69}" font-family="{MONO}" font-size="10" fill="{SOFT}">{hexv}</text>')

    # componentes: uma moldura com abas em chip, lista de opções e painel de descrição
    o.append(label(40, 950, "Componentes"))
    fx, fy, fw = 40, 962, 1200
    o.append(f'<rect x="{fx}" y="{fy}" width="{fw}" height="290" fill="{c("superficie", "preto-fundo")}"/>')
    o.append('<g clip-path="url(#cBody)">')
    o.append('<rect x="40" y="996" width="1200" height="222" fill="url(#gSky)"/>')
    o.append(palm(96, 1226, 190, 22, 1.1, c("superficie", "ameixa"), 0.38))
    o.append(palm(1196, 1226, 150, -18, 0.9, c("superficie", "ameixa"), 0.3))
    o.append('</g>')
    o.append(f'<rect x="{fx}" y="{fy}" width="{fw}" height="34" fill="{c("superficie", "preto-moldura")}"/>')
    o.append(f'<rect x="{fx}" y="1218" width="{fw}" height="34" fill="{c("superficie", "preto-moldura")}"/>')
    o.append(neon(fx, 996, fw))
    o.append(neon(fx, 1216, fw))
    # abas em chip: as 5 missões, com a terceira ativa
    passo = fw / len(titulos)
    for i, t in enumerate(titulos):
        cx = fx + passo * (i + 0.5)
        ativa = i == 2
        if ativa:
            o.append(f'<rect x="{cx - 82:.0f}" y="968" width="164" height="22" rx="3" fill="url(#gChip)"/>')
        extra = f' {CONTORNO}' if ativa else ''
        o.append(f'<text x="{cx:.0f}" y="983.5" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="12" fill="#ffffff"{extra}>{t}</text>')
    # lista de opções
    for i, (nome, valor, ativo) in enumerate([("Agente da Discovery", "CLAUDE", False), ("Agente do Builder", "CODEX", True), ("Agente do Reviewer", "GEMINI", False), ("Permissões", "FULL", False), ("Teto de tokens", "OFF", False)]):
        y = 1016 + i * 34
        if ativo:
            o.append(f'<rect x="70" y="{y}" width="510" height="30" fill="{c("acento", "ambar")}"/>')
        else:
            o.append(f'<rect x="70" y="{y}" width="510" height="30" fill="{c("superficie", "preto-fundo")}" fill-opacity="0.42"/>')
        o.append(f'<text x="86" y="{y + 19.5}" font-family="{FONT}" font-weight="700" font-size="12.5" fill="#ffffff"{" " + CONTORNO if ativo else ""}>{nome}</text>')
        o.append(f'<text x="564" y="{y + 19.5}" text-anchor="end" font-family="{FONT}" font-weight="700" font-size="11.5" letter-spacing="0.8" fill="#ffffff"{" " + CONTORNO if ativo else ""}>{valor}</text>')
    # painel de descrição
    o.append(f'<rect x="606" y="1016" width="330" height="26" fill="{c("superficie", "preto-fundo")}" fill-opacity="0.6"/>')
    o.append(f'<text x="771" y="1033.5" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="11" letter-spacing="1.4" fill="#ffffff">DESCRIÇÃO</text>')
    o.append(f'<rect x="606" y="1046" width="330" height="136" fill="{c("superficie", "painel")}" fill-opacity="0.9"/>')
    for i, linha in enumerate(["Escolhe o agente que roda em cada sessão do", "pipeline. O Reviewer só lê: quem confere não é", "quem construiu."]):
        o.append(f'<text x="622" y="{1072 + i * 18}" font-family="{FONT}" font-weight="500" font-size="12" fill="#ffffff">{linha}</text>')
    for i, linha in enumerate(["Trocar o agente no meio de um job reinicia a sessão."]):
        o.append(f'<text x="622" y="{1150 + i * 16}" font-family="{FONT}" font-weight="600" font-size="12" fill="{c("texto", "aviso")}">{linha}</text>')
    # destaque da versão e estados de aba
    o.append(f'<text x="988" y="1096" font-family="{FONT}" font-weight="800" font-size="78" letter-spacing="-3" fill="url(#gSig)" stroke="#151515" stroke-opacity="0.5" stroke-width="2" paint-order="stroke fill">5.0</text>')
    o.append('<rect x="990" y="1128" width="74" height="22" rx="3" fill="url(#gChip)"/>')
    o.append(f'<text x="1027" y="1143.5" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="11.5" fill="#ffffff" {CONTORNO}>Ativa</text>')
    o.append(f'<text x="1084" y="1143.5" font-family="{FONT}" font-weight="700" font-size="11.5" fill="#ffffff">Inativa</text>')
    # barra inferior: promessa
    o.append(f'<text x="70" y="1240" font-family="{FONT}" font-weight="600" font-size="12.5" fill="#ffffff">{m["promessa"]}</text>')

    # rodapé
    o.append(note(40, 1288, "Conteúdo original, sem ligação com a Rockstar/Take-Two. Nenhuma arte, logotipo, letreiro ou fonte de terceiros é distribuída aqui."))
    o.append(note(40, 1308, "Cores amostradas das pranchas publicadas, com tolerância de ±2 por canal. Fonte: behance.net/gallery/224512205"))
    o.append('</svg>')
    return '\n'.join(o) + '\n'


if __name__ == "__main__":
    with open(sys.argv[1], "w", encoding="utf-8") as f:
        f.write(build())
    print("ok", sys.argv[1])

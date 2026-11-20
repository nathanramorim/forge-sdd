"""Banner do README na Edição Missões. Uso: python3 banner.py <saida.svg>"""
import sys

from parts import CONTORNO, FONT, MONO, grad, logo, neon, palm, skyline, striped_sun
from tokens import c, missoes, titulos_missoes

W, H = 1280, 420
TOP, BOT = 40, 372
ART0, ART1 = 42, 370


def build():
    m = missoes()
    titulos = titulos_missoes()
    o = []
    rotulo = f'forge-sdd 5.0, {m["edicao"]}: seu time de IA, com método. ' + '; '.join(titulos) + '.'
    o.append(f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" role="img" aria-label="{rotulo}">')
    o.append(f'<title>forge-sdd 5.0 — {m["edicao"]}</title>')
    o.append('<defs>')
    o.append(grad("gSig", "assinatura"))
    o.append(grad("gChip", "chip"))
    o.append(grad("gSky", "ceu"))
    o.append('<linearGradient id="gSun" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#ffaa6e"/><stop offset="0.5" stop-color="#fe818d"/><stop offset="1" stop-color="#f061af"/></linearGradient>')
    o.append('<linearGradient id="gScrim" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#060506" stop-opacity="0.9"/><stop offset="0.36" stop-color="#060506" stop-opacity="0.8"/><stop offset="0.56" stop-color="#060506" stop-opacity="0.46"/><stop offset="0.72" stop-color="#060506" stop-opacity="0"/></linearGradient>')
    o.append('<linearGradient id="gFloor" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#060506" stop-opacity="0"/><stop offset="1" stop-color="#060506" stop-opacity="0.62"/></linearGradient>')
    o.append('<radialGradient id="gHalo" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="#ffb06d" stop-opacity="0.55"/><stop offset="1" stop-color="#ffb06d" stop-opacity="0"/></radialGradient>')
    o.append('<filter id="fGlow" x="-5%" y="-400%" width="110%" height="900%"><feGaussianBlur stdDeviation="3.2"/></filter>')
    o.append('<filter id="fDrop" x="-10%" y="-10%" width="120%" height="130%"><feDropShadow dx="0" dy="6" stdDeviation="7" flood-color="#060506" flood-opacity="0.55"/></filter>')
    o.append(f'<clipPath id="cArt"><rect x="0" y="{ART0}" width="{W}" height="{ART1 - ART0}"/></clipPath>')
    o.append(f'<clipPath id="cAll"><rect width="{W}" height="{H}" rx="10"/></clipPath>')
    o.append('</defs>')
    o.append('<g clip-path="url(#cAll)">')

    # arte: céu, sol, cidade e palmeiras
    o.append(f'<rect width="{W}" height="{H}" fill="{c("superficie", "preto-fundo")}"/>')
    o.append('<g clip-path="url(#cArt)">')
    o.append(f'<rect x="0" y="{ART0}" width="{W}" height="{ART1 - ART0}" fill="url(#gSky)"/>')
    scx, scy, sr = 1010, 222, 112
    o.append(f'<circle cx="{scx}" cy="{scy}" r="{sr * 1.9:.0f}" fill="url(#gHalo)"/>')
    o.append(striped_sun("mSun", scx, scy, sr, "url(#gSun)"))
    for bx, by, k in [(892, 92, 1.0), (924, 106, 0.8), (866, 116, 0.7)]:
        o.append(f'<path d="M{bx},{by} q{5 * k:.1f},{-6 * k:.1f} {10 * k:.1f},0 q{5 * k:.1f},{-6 * k:.1f} {10 * k:.1f},0" fill="none" stroke="{c("superficie", "ameixa")}" stroke-width="1.6" stroke-linecap="round" opacity="0.8"/>')
    longe = "#684768"
    o.append(skyline(340, [(640, 46, 26), (690, 34, 44), (728, 52, 30), (786, 30, 58), (820, 44, 36), (872, 38, 50), (914, 60, 28), (980, 34, 62), (1018, 48, 40), (1072, 40, 54), (1116, 56, 32), (1178, 36, 48), (1218, 62, 30)], longe))
    escuro = "#3a1430"
    o.append(palm(800, 372, 118, 10, 0.55, longe, 0.95))
    o.append(palm(858, 372, 150, 16, 0.72, c("superficie", "ameixa")))
    o.append(palm(1142, 372, 232, -40, 1.28, escuro))
    o.append(palm(1236, 372, 168, 16, 1.0, escuro))
    o.append(f'<rect x="0" y="{ART0}" width="{W}" height="{ART1 - ART0}" fill="url(#gScrim)"/>')
    o.append(f'<rect x="0" y="262" width="{W}" height="{ART1 - 262}" fill="url(#gFloor)"/>')
    o.append('</g>')

    # o 5.0 como destaque gráfico, na tipografia da marca
    o.append(f'<text x="40" y="290" font-family="{FONT}" font-weight="800" font-size="190" letter-spacing="-6" fill="url(#gSig)" '
             f'stroke="#151515" stroke-opacity="0.55" stroke-width="3" stroke-linejoin="round" paint-order="stroke fill" filter="url(#fDrop)">5.0</text>')

    # bloco de título
    tx = 376
    o.append(f'<rect x="{tx}" y="150" width="156" height="24" rx="3" fill="url(#gChip)"/>')
    o.append(f'<text x="{tx + 78}" y="167" text-anchor="middle" font-family="{FONT}" font-weight="800" font-size="12" letter-spacing="0.9" fill="#ffffff" {CONTORNO}>{m["edicao"].upper()}</text>')
    o.append(f'<text x="{tx + 170}" y="167" font-family="{FONT}" font-weight="700" font-size="12" letter-spacing="2.2" fill="#f6dfe6">NOVA VERSÃO ESTÁVEL</text>')
    o.append(f'<text x="{tx}" y="208" font-family="{FONT}" font-weight="700" font-size="20" letter-spacing="1.6" fill="#ffffff">SEU TIME DE IA, COM MÉTODO.</text>')
    o.append(f'<text x="{tx}" y="238" font-family="{FONT}" font-weight="700" font-size="15.5" fill="#ffffff">{m["promessa"]}</text>')
    for i, linha in enumerate(["Do plano ao PR, sem repetir instruções. A Discovery", "planeja, o Builder constrói e o Reviewer confere."]):
        o.append(f'<text x="{tx}" y="{264 + i * 20}" font-family="{FONT}" font-weight="500" font-size="14.5" fill="#ffffff" fill-opacity="0.93">{linha}</text>')

    # barra superior: marca e comando
    o.append(f'<rect x="0" y="0" width="{W}" height="{TOP}" fill="{c("superficie", "preto-moldura")}"/>')
    o.append(logo(24, 8, 0.75))
    o.append(f'<text x="56" y="25" font-family="{FONT}" font-size="15"><tspan font-weight="800" fill="#ffffff">forge-sdd</tspan>'
             f'<tspan dx="10" font-weight="600" font-size="12" fill="{c("texto", "inativo")}">{m["edicao"]}</tspan></text>')
    o.append(f'<text x="1250" y="25" text-anchor="end" font-family="{MONO}" font-size="13.5"><tspan fill="{c("texto", "ativo")}" font-weight="700">$</tspan>'
             f'<tspan dx="8" fill="#ffffff" fill-opacity="0.94">npm i -g @nathanramorim/forge-sdd</tspan></text>')

    # barra inferior: as missões e o fecho, da fonte única
    o.append(f'<rect x="0" y="{BOT}" width="{W}" height="{H - BOT}" fill="{c("superficie", "preto-moldura")}"/>')
    passo = W / len(titulos)
    for i, t in enumerate(titulos):
        cx = passo * (i + 0.5)
        ultima = i == len(titulos) - 1
        if ultima:
            o.append(f'<rect x="{cx - 112:.0f}" y="383" width="224" height="26" rx="3" fill="url(#gChip)"/>')
        extra = f' {CONTORNO}' if ultima else ''
        o.append(f'<text x="{cx:.0f}" y="401" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="13" fill="#ffffff"{extra}>{t}</text>')
        if i:
            o.append(f'<path d="M{passo * i - 3:.0f},391 l6,5.5 l-6,5.5" fill="none" stroke="{c("texto", "inativo")}" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>')

    o.append(neon(0, TOP, W))
    o.append(neon(0, ART1, W))
    o.append('</g>')
    o.append('</svg>')
    return '\n'.join(o) + '\n'


if __name__ == "__main__":
    with open(sys.argv[1], "w", encoding="utf-8") as f:
        f.write(build())
    print("ok", sys.argv[1])

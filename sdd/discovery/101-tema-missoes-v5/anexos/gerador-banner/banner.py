import sys
from parts import *

W, H = 1280, 420
TOP, BOT = 40, 372          # barra superior até 40; barra inferior a partir de 372
ART0, ART1 = 42, 370

def build():
    o = []
    o.append(f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" role="img" '
             f'aria-label="forge-sdd 5.0, edição GTA VI: seu time de IA, com método. Discovery planeja, Builder constrói e Reviewer confere.">')
    o.append('<title>forge-sdd 5.0 — edição GTA VI</title>')
    o.append('<defs>')
    o.append(grad("gSig", "assinatura"))
    o.append(grad("gChip", "chip"))
    o.append(grad("gSky", "ceu"))
    o.append('<linearGradient id="gSun" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#ffaa6e"/><stop offset="0.5" stop-color="#fe818d"/><stop offset="1" stop-color="#f061af"/></linearGradient>')
    o.append('<linearGradient id="gV" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#4053c6"/><stop offset="0.3" stop-color="#a849c7"/><stop offset="0.52" stop-color="#e747c9"/><stop offset="0.76" stop-color="#f66e9f"/><stop offset="1" stop-color="#ffaa6e"/></linearGradient>')
    o.append('<linearGradient id="gScrim" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#060506" stop-opacity="0.9"/><stop offset="0.36" stop-color="#060506" stop-opacity="0.8"/><stop offset="0.56" stop-color="#060506" stop-opacity="0.46"/><stop offset="0.72" stop-color="#060506" stop-opacity="0"/></linearGradient>')
    o.append('<linearGradient id="gFloor" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#060506" stop-opacity="0"/><stop offset="1" stop-color="#060506" stop-opacity="0.62"/></linearGradient>')
    o.append('<radialGradient id="gHalo" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="#ffb06d" stop-opacity="0.55"/><stop offset="1" stop-color="#ffb06d" stop-opacity="0"/></radialGradient>')
    o.append('<filter id="fGlow" x="-5%" y="-400%" width="110%" height="900%"><feGaussianBlur stdDeviation="3.2"/></filter>')
    o.append('<filter id="fDrop" x="-20%" y="-20%" width="140%" height="150%"><feDropShadow dx="0" dy="6" stdDeviation="7" flood-color="#060506" flood-opacity="0.55"/></filter>')
    o.append(f'<clipPath id="cArt"><rect x="0" y="{ART0}" width="{W}" height="{ART1 - ART0}"/></clipPath>')
    o.append(f'<clipPath id="cAll"><rect width="{W}" height="{H}" rx="10"/></clipPath>')
    o.append('</defs>')
    o.append('<g clip-path="url(#cAll)">')

    # --- arte: céu, sol, cidade, palmeiras
    o.append(f'<rect width="{W}" height="{H}" fill="{c("superficie", "preto-fundo")}"/>')
    o.append('<g clip-path="url(#cArt)">')
    o.append(f'<rect x="0" y="{ART0}" width="{W}" height="{ART1 - ART0}" fill="url(#gSky)"/>')
    scx, scy, sr = 1010, 222, 112
    o.append(f'<circle cx="{scx}" cy="{scy}" r="{sr * 1.9:.0f}" fill="url(#gHalo)"/>')
    o.append(striped_sun("mSun", scx, scy, sr, "url(#gSun)"))
    # aves
    for bx, by, k in [(892, 92, 1.0), (924, 106, 0.8), (866, 116, 0.7)]:
        o.append(f'<path d="M{bx},{by} q{5 * k:.1f},{-6 * k:.1f} {10 * k:.1f},0 q{5 * k:.1f},{-6 * k:.1f} {10 * k:.1f},0" fill="none" stroke="{c("superficie", "ameixa")}" stroke-width="1.6" stroke-linecap="round" opacity="0.8"/>')
    # skyline art déco (blocos com topo escalonado)
    far = "#684768"
    blds = [(640, 46, 26), (690, 34, 44), (728, 52, 30), (786, 30, 58), (820, 44, 36), (872, 38, 50), (914, 60, 28), (980, 34, 62), (1018, 48, 40), (1072, 40, 54), (1116, 56, 32), (1178, 36, 48), (1218, 62, 30)]
    g = []
    for i, (bx, bw, bh) in enumerate(blds):
        top = 340 - bh
        g.append(f'<rect x="{bx}" y="{top}" width="{bw}" height="{bh + 32}"/>')
        g.append(f'<rect x="{bx + bw * 0.22:.0f}" y="{top - 7}" width="{bw * 0.56:.0f}" height="8"/>')
        if i % 3 == 1:
            g.append(f'<rect x="{bx + bw / 2 - 1:.0f}" y="{top - 22}" width="2" height="16"/>')
    o.append(f'<g fill="{far}" opacity="0.92">{"".join(g)}</g>')
    win = []
    for i, (bx, bw, bh) in enumerate(blds):
        if i % 2:
            continue
        for row in range(int(bh // 11)):
            for col in range(int((bw - 10) // 9)):
                if (row * 3 + col * 5 + i) % 4 == 0:
                    continue
                win.append(f'<rect x="{bx + 6 + col * 9}" y="{340 - bh + 7 + row * 11}" width="4" height="5"/>')
    o.append(f'<g fill="#ffb06d" opacity="0.55">{"".join(win)}</g>')
    o.append(f'<rect x="0" y="340" width="{W}" height="32" fill="{c("superficie", "ameixa")}"/>')
    # palmeiras
    dark = "#3a1430"
    o.append(palm(800, 372, 118, 10, 0.55, far, 0.95))
    o.append(palm(858, 372, 150, 16, 0.72, c("superficie", "ameixa")))
    o.append(palm(1142, 372, 232, -40, 1.28, dark))
    o.append(palm(1236, 372, 168, 16, 1.0, dark))
    # véus para o bloco de título
    o.append(f'<rect x="0" y="{ART0}" width="{W}" height="{ART1 - ART0}" fill="url(#gScrim)"/>')
    o.append(f'<rect x="0" y="262" width="{W}" height="{ART1 - 262}" fill="url(#gFloor)"/>')
    o.append('</g>')

    # --- lockup: numeral V (versão 5) com o letreiro por cima
    V = [(50, 88), (128, 88), (175, 256), (222, 88), (300, 88), (212, 336), (138, 336)]
    pts = ' '.join(f'{a},{b}' for a, b in V)
    o.append('<g filter="url(#fDrop)">')
    o.append(f'<polygon points="{pts}" fill="{c("superficie", "ameixa")}" stroke="{c("superficie", "ameixa")}" stroke-width="14" stroke-linejoin="round"/>')
    o.append(f'<polygon points="{pts}" fill="url(#gV)"/>')
    o.append(f'<polygon points="{pts}" fill="none" stroke="#ffffff" stroke-opacity="0.28" stroke-width="1.5" stroke-linejoin="round"/>')
    wf = word_width("forge")
    ws = word_width("sdd")
    o.append(word("forge", 175 - wf / 2, 146))
    o.append(word("sdd", 175 - ws / 2, 224))
    o.append('</g>')

    # --- bloco de título
    tx = 348
    o.append(f'<rect x="{tx}" y="170" width="84" height="24" rx="3" fill="url(#gChip)"/>')
    o.append(f'<text x="{tx + 42}" y="187" text-anchor="middle" font-family="{HEAVY}" font-weight="900" font-size="12.5" letter-spacing="0.8" fill="#ffffff" stroke="#151515" stroke-opacity="0.5" stroke-width="2.4" stroke-linejoin="round" paint-order="stroke fill">V5.0.0</text>')
    o.append(f'<text x="{tx + 96}" y="187" font-family="{FONT}" font-weight="700" font-size="12" letter-spacing="2.4" fill="#f6dfe6">NOVA VERSÃO ESTÁVEL</text>')
    o.append(f'<text x="{tx}" y="232" font-family="{FONT}" font-weight="700" font-size="22" letter-spacing="1.8" fill="#ffffff">SEU TIME DE IA, COM MÉTODO.</text>')
    for i, line in enumerate(["Do plano ao PR, sem repetir instruções. A Discovery",
                              "planeja, o Builder constrói e o Reviewer confere.",
                              "Nada é entregue sem a sua aprovação final."]):
        o.append(f'<text x="{tx}" y="{261 + i * 21}" font-family="{FONT}" font-weight="500" font-size="15" fill="#ffffff" fill-opacity="0.93">{line}</text>')

    # --- HUD: nível de procurado = versão
    for i in range(5):
        o.append(star(1164 + i * 20, 66, 8.5, "#ffffff"))
    o.append(f'<text x="1252" y="99" text-anchor="end" font-family="{HEAVY}" font-weight="900" font-size="19" fill="#ffffff" stroke="{c("superficie", "preto-moldura")}" stroke-width="4" stroke-linejoin="round" paint-order="stroke fill">v5.0.0</text>')

    # --- barra superior com abas
    o.append(f'<rect x="0" y="0" width="{W}" height="{TOP}" fill="{c("superficie", "preto-moldura")}"/>')
    o.append('<g transform="translate(24 8) scale(0.75)">'
             '<path d="M16 2.2 L28.4 9.1 V22.9 L16 29.8 L3.6 22.9 V9.1 Z" fill="none" stroke="url(#gSig)" stroke-width="2" stroke-linejoin="round"/>'
             '<path d="M16 9 L22.5 12.6 V13.4 L16 17 L9.5 13.4 V12.6 Z" fill="#ffffff"/>'
             '<path d="M9.5 13.4 L16 17 V23.8 L9.5 20.2 Z" fill="#f6dfe6"/>'
             '<path d="M22.5 13.4 L16 17 V23.8 L22.5 20.2 Z" fill="#e2bccb"/></g>')
    o.append(f'<text x="56" y="25" font-family="{FONT}" font-size="15"><tspan font-weight="800" fill="#ffffff">forge-sdd</tspan>'
             f'<tspan dx="10" font-weight="600" font-size="12" fill="{c("texto", "inativo")}">edição GTA VI</tspan></text>')
    o.append(keycap(866, 10, "L1"))
    o.append('<rect x="900" y="9" width="108" height="22" rx="3" fill="url(#gChip)"/>')
    tab = f'font-family="{HEAVY}" font-weight="900" font-size="11.5" letter-spacing="0.8" fill="#ffffff" text-anchor="middle"'
    o.append(f'<text x="954" y="24" {tab} stroke="#151515" stroke-opacity="0.5" stroke-width="2.4" stroke-linejoin="round" paint-order="stroke fill">DISCOVERY</text>')
    o.append(f'<text x="1062" y="24" {tab}>BUILDER</text>')
    o.append(f'<text x="1158" y="24" {tab}>REVIEWER</text>')
    o.append(keycap(1214, 10, "R1"))

    # --- barra inferior com comando e dicas
    o.append(f'<rect x="0" y="{BOT}" width="{W}" height="{H - BOT}" fill="{c("superficie", "preto-moldura")}"/>')
    o.append(f'<text x="44" y="401" font-family="{MONO}" font-size="14"><tspan fill="{c("texto", "ativo")}" font-weight="700">$</tspan>'
             f'<tspan dx="9" fill="#ffffff" fill-opacity="0.94">npm i -g @nathanramorim/forge-sdd</tspan></text>')
    hint = f'font-family="{FONT}" font-weight="600" font-size="13" fill="#ffffff" text-anchor="end"'
    o.append(f'<text x="868" y="401" {hint}>Sessões</text>')
    o.append(keycap(876, 386, "L1", 24, 20) + keycap(904, 386, "R1", 24, 20))
    o.append(f'<text x="1002" y="401" {hint}>Planejar</text>' + face_button(1020, 396, "t"))
    o.append(f'<text x="1108" y="401" {hint}>Construir</text>' + face_button(1126, 396, "o"))
    o.append(f'<text x="1210" y="401" {hint}>Aprovar</text>' + face_button(1228, 396, "x"))

    # --- linhas neon
    for y in (TOP, ART1):
        o.append(f'<rect x="0" y="{y - 1}" width="{W}" height="4" fill="url(#gSig)" opacity="0.75" filter="url(#fGlow)"/>')
        o.append(f'<rect x="0" y="{y}" width="{W}" height="2" fill="url(#gSig)"/>')
    o.append('</g>')
    o.append('</svg>')
    return '\n'.join(o) + '\n'

if __name__ == "__main__":
    open(sys.argv[1], "w", encoding="utf-8").write(build())
    print("ok", sys.argv[1])

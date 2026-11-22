"""Arte do topo da LP na Edição Missões (sem texto). Uso: python3 hero.py <saida.svg>"""
import sys

from parts import grad, palm, skyline, striped_sun
from tokens import c

W, H = 1600, 620


def build():
    longe = "#684768"
    escuro = "#3a1430"
    fundo = c("superficie", "ameixa")
    o = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" preserveAspectRatio="xMidYMax slice" aria-hidden="true">']
    o.append('<defs>')
    o.append(grad("gSky", "ceu"))
    o.append('<radialGradient id="gHalo" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="#ffb06d" stop-opacity="0.6"/><stop offset="1" stop-color="#ffb06d" stop-opacity="0"/></radialGradient>')
    o.append('<linearGradient id="gSun" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#ffaa6e"/><stop offset="0.5" stop-color="#fe818d"/><stop offset="1" stop-color="#f061af"/></linearGradient>')
    o.append(f'<linearGradient id="gFade" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="{fundo}" stop-opacity="0"/><stop offset="0.72" stop-color="{fundo}" stop-opacity="0"/><stop offset="1" stop-color="{fundo}"/></linearGradient>')
    o.append('<linearGradient id="gTopo" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#151515" stop-opacity="0.85"/><stop offset="0.35" stop-color="#151515" stop-opacity="0"/></linearGradient>')
    o.append('</defs>')
    o.append(f'<rect width="{W}" height="{H}" fill="url(#gSky)"/>')
    o.append(f'<circle cx="800" cy="330" r="360" fill="url(#gHalo)"/>')
    o.append(striped_sun("mSun", 800, 330, 190, "url(#gSun)"))
    for bx, by, k in [(520, 150, 1.2), (566, 172, 0.9), (1010, 130, 1.0), (1052, 150, 0.8)]:
        o.append(f'<path d="M{bx},{by} q{6 * k:.1f},{-7 * k:.1f} {12 * k:.1f},0 q{6 * k:.1f},{-7 * k:.1f} {12 * k:.1f},0" fill="none" stroke="{fundo}" stroke-width="1.8" stroke-linecap="round" opacity="0.8"/>')
    o.append(skyline(540, [(0, 70, 60), (80, 50, 96), (140, 80, 70), (230, 44, 120), (284, 70, 84), (364, 56, 104), (430, 90, 66), (1070, 60, 110), (1140, 52, 76), (1202, 84, 98), (1296, 48, 128), (1354, 74, 80), (1438, 60, 106), (1508, 92, 70)], longe, largura=W, chao=H - 540))
    o.append(palm(150, 590, 300, 24, 1.5, escuro))
    o.append(palm(250, 590, 210, -16, 1.0, fundo))
    o.append(palm(1380, 590, 330, -34, 1.6, escuro))
    o.append(palm(1480, 590, 220, 18, 1.1, fundo))
    o.append(f'<rect width="{W}" height="{H}" fill="url(#gTopo)"/>')
    o.append(f'<rect width="{W}" height="{H}" fill="url(#gFade)"/>')
    o.append('</svg>')
    return '\n'.join(o) + '\n'


if __name__ == "__main__":
    with open(sys.argv[1], "w", encoding="utf-8") as f:
        f.write(build())
    print("ok", sys.argv[1])

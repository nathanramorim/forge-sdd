"""Peças compartilhadas pelo banner e pelo espécime da Edição Missões."""
import math

from tokens import FONTS, GRADIENTS, c

FONT = FONTS["marca"]["pilha"]
MONO = FONTS["mono"]["pilha"]
CONTORNO = 'stroke="#151515" stroke-opacity="0.5" stroke-width="2.4" stroke-linejoin="round" paint-order="stroke fill"'


def grad(gid, name, horizontal=None):
    direction, _, stops = GRADIENTS[name]
    vertical = direction == "vertical" if horizontal is None else not horizontal
    x2, y2 = (0, 1) if vertical else (1, 0)
    s = ''.join(f'<stop offset="{o / 100:g}" stop-color="{col}"/>' for o, col in stops)
    return f'<linearGradient id="{gid}" x1="0" y1="0" x2="{x2}" y2="{y2}">{s}</linearGradient>'


def logo(x, y, scale):
    return (f'<g transform="translate({x} {y}) scale({scale})">'
            '<path d="M16 2.2 L28.4 9.1 V22.9 L16 29.8 L3.6 22.9 V9.1 Z" fill="none" stroke="url(#gSig)" stroke-width="2" stroke-linejoin="round"/>'
            '<path d="M16 9 L22.5 12.6 V13.4 L16 17 L9.5 13.4 V12.6 Z" fill="#ffffff"/>'
            '<path d="M9.5 13.4 L16 17 V23.8 L9.5 20.2 Z" fill="#f6dfe6"/>'
            '<path d="M22.5 13.4 L16 17 V23.8 L22.5 20.2 Z" fill="#e2bccb"/></g>')


def neon(x, y, w):
    return (f'<rect x="{x}" y="{y - 1}" width="{w}" height="4" fill="url(#gSig)" opacity="0.72" filter="url(#fGlow)"/>'
            f'<rect x="{x}" y="{y}" width="{w}" height="2" fill="url(#gSig)"/>')


def palm(x, by, h, lean, s, fill, op=1.0):
    tx, ty = x + lean, by - h
    cx, cy = x + lean * 0.15, by - h * 0.55
    wb, wt = 10 * s, 5 * s
    trunk = (f'M{x - wb / 2:.1f},{by} Q{cx - wb * 0.35:.1f},{cy:.1f} {tx - wt / 2:.1f},{ty:.1f} '
             f'L{tx + wt / 2:.1f},{ty:.1f} Q{cx + wb * 0.35:.1f},{cy:.1f} {x + wb / 2:.1f},{by} Z')
    L = 78 * s
    fr = []
    for a, k in [(205, .70), (172, .95), (146, 1.0), (120, .88), (94, .74), (66, .88), (38, 1.0), (10, .95), (-22, .70)]:
        r = math.radians(a)
        ux, uy = math.cos(r), -math.sin(r)
        Lk = L * k
        droop = Lk * (0.44 - 0.24 * abs(math.sin(r)))
        ex, ey = tx + ux * Lk, ty + uy * Lk * 0.9 + droop
        c1x, c1y = tx + ux * Lk * 0.55, ty + uy * Lk * 0.75 - Lk * 0.18
        nx, ny = -uy, ux
        if ny < 0:
            nx, ny = -nx, -ny
        th = Lk * 0.22
        c2x, c2y = c1x + nx * th, c1y + ny * th
        fr.append(f'M{tx:.1f},{ty:.1f} Q{c1x:.1f},{c1y:.1f} {ex:.1f},{ey:.1f} Q{c2x:.1f},{c2y:.1f} {tx:.1f},{ty:.1f} Z')
    return (f'<g fill="{fill}" opacity="{op}"><path d="{trunk}"/><path d="{" ".join(fr)}"/>'
            f'<circle cx="{tx:.1f}" cy="{ty:.1f}" r="{5 * s:.1f}"/></g>')


def striped_sun(mask_id, cx, cy, r, fill):
    bars = []
    for i in range(7):
        y = cy + r * (0.06 + 0.135 * i)
        bars.append(f'<rect x="{cx - r - 2}" y="{y:.1f}" width="{2 * r + 4}" height="{r * (0.02 + 0.0135 * i):.1f}" fill="#000"/>')
    return (f'<mask id="{mask_id}"><circle cx="{cx}" cy="{cy}" r="{r}" fill="#fff"/>{"".join(bars)}</mask>'
            f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{fill}" mask="url(#{mask_id})"/>')


def skyline(base_y, blds, cor, largura=1280, chao=32):
    g, win = [], []
    for i, (bx, bw, bh) in enumerate(blds):
        top = base_y - bh
        g.append(f'<rect x="{bx}" y="{top}" width="{bw}" height="{bh + 32}"/>')
        g.append(f'<rect x="{bx + bw * 0.22:.0f}" y="{top - 7}" width="{bw * 0.56:.0f}" height="8"/>')
        if i % 3 == 1:
            g.append(f'<rect x="{bx + bw / 2 - 1:.0f}" y="{top - 22}" width="2" height="16"/>')
        if i % 2:
            continue
        for row in range(int(bh // 11)):
            for col in range(int((bw - 10) // 9)):
                if (row * 3 + col * 5 + i) % 4:
                    win.append(f'<rect x="{bx + 6 + col * 9}" y="{top + 7 + row * 11}" width="4" height="5"/>')
    return (f'<g fill="{cor}" opacity="0.92">{"".join(g)}</g>'
            f'<g fill="#ffb06d" opacity="0.55">{"".join(win)}</g>'
            f'<rect x="0" y="{base_y}" width="{largura}" height="{chao}" fill="{c("superficie", "ameixa")}"/>')

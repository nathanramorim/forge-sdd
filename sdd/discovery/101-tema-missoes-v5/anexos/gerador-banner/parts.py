import math
from tokens import TOKENS, GRADIENTS, FONTS, c

FONT = "'Avenir Next', Avenir, Futura, 'Century Gothic', Montserrat, 'Segoe UI', system-ui, -apple-system, Roboto, sans-serif"
HEAVY = "'Arial Black', 'Avenir Next', 'Helvetica Neue', 'Segoe UI', system-ui, sans-serif"
MONO = FONTS["mono"]["pilha"]
SLAB = FONTS["rotulo"]["pilha"]

def grad(gid, name, x1=0, y1=0, x2=None, y2=None, units=None):
    direction, _, stops = GRADIENTS[name]
    if x2 is None:
        x2, y2 = (0, 1) if direction == "vertical" else (1, 0)
    u = f' gradientUnits="{units}"' if units else ''
    s = ''.join(f'<stop offset="{o / 100:g}" stop-color="{col}"/>' for o, col in stops)
    return f'<linearGradient id="{gid}" x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}"{u}>{s}</linearGradient>'

# Letreiro próprio (unicase em bloco): traços de linha central numa caixa 48x64, peso 14.
GLYPHS = {
    'f': (48, ["M17,64 V7 H48", "M0,30 H38"], []),
    'o': (48, ["M7,7 H41 V57 H7 Z"], []),
    'r': (48, ["M7,64 V7 H41 V35 H7"], [[(14, 35), (30, 35), (48, 64), (31, 64)]]),
    'g': (48, ["M48,7 H7 V57 H41 V33 H23"], []),
    'e': (48, ["M7,32 H41 V7 H7 V57 H48"], []),
    's': (48, ["M48,7 H7 V32 H41 V57 H0"], []),
    'd': (48, ["M41,0 V57 H7 V25 H41"], []),
    '-': (30, ["M2,32 H28"], []),
}

def word_width(text, gap=8):
    return sum(GLYPHS[ch][0] for ch in text) + gap * (len(text) - 1)

def word(text, x, y, scale=1.0, gap=8, weight=14, outline=5, fill="#ffffff", edge="#151515"):
    strokes, polys = [], []
    cx = 0
    for ch in text:
        w, paths, pg = GLYPHS[ch]
        for d in paths:
            strokes.append(f'<path transform="translate({cx} 0)" d="{d}"/>')
        for pts in pg:
            polys.append('<polygon transform="translate(%d 0)" points="%s"/>' % (cx, ' '.join(f'{a},{b}' for a, b in pts)))
        cx += w + gap
    s, p = ''.join(strokes), ''.join(polys)
    return (
        f'<g transform="translate({x} {y}) scale({scale})">'
        f'<g fill="none" stroke="{edge}" stroke-width="{weight + 2 * outline}" stroke-linejoin="round" stroke-linecap="round">{s}</g>'
        f'<g fill="{edge}" stroke="{edge}" stroke-width="{2 * outline}" stroke-linejoin="round">{p}</g>'
        f'<g fill="none" stroke="{fill}" stroke-width="{weight}" stroke-linejoin="miter" stroke-linecap="butt">{s}</g>'
        f'<g fill="{fill}">{p}</g>'
        f'</g>'
    )

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

def star(cx, cy, R, fill):
    pts = []
    for i in range(10):
        r = R if i % 2 == 0 else R * 0.42
        a = math.radians(-90 + i * 36)
        pts.append(f'{cx + r * math.cos(a):.1f},{cy + r * math.sin(a):.1f}')
    return f'<polygon points="{" ".join(pts)}" fill="{fill}"/>'

def keycap(x, y, label, w=26, h=20, stroke="url(#gSig)"):
    return (f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="5" fill="{c("superficie", "tecla")}" stroke="{stroke}" stroke-width="1.2"/>'
            f'<text x="{x + w / 2}" y="{y + h / 2 + 3.6}" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="10.5" fill="#ffffff">{label}</text>')

def face_button(cx, cy, kind, r=10):
    light = "#f6dfe6"
    base = f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{c("superficie", "tecla")}" stroke="{light}" stroke-width="1.3"/>'
    k = r * 0.42
    if kind == "x":
        g = f'<path d="M{cx - k},{cy - k} L{cx + k},{cy + k} M{cx + k},{cy - k} L{cx - k},{cy + k}" stroke="{light}" stroke-width="1.6" stroke-linecap="round"/>'
    elif kind == "o":
        g = f'<circle cx="{cx}" cy="{cy}" r="{k}" fill="none" stroke="{light}" stroke-width="1.6"/>'
    else:
        g = f'<path d="M{cx},{cy - k - 0.6} L{cx + k + 0.6},{cy + k - 0.4} L{cx - k - 0.6},{cy + k - 0.4} Z" fill="none" stroke="{light}" stroke-width="1.5" stroke-linejoin="round"/>'
    return base + g

def striped_sun(mask_id, cx, cy, r, fill):
    bars = []
    for i in range(7):
        y = cy + r * (0.06 + 0.135 * i)
        bars.append(f'<rect x="{cx - r - 2}" y="{y:.1f}" width="{2 * r + 4}" height="{r * (0.02 + 0.0135 * i):.1f}" fill="#000"/>')
    return (f'<mask id="{mask_id}"><circle cx="{cx}" cy="{cy}" r="{r}" fill="#fff"/>{"".join(bars)}</mask>'
            f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{fill}" mask="url(#{mask_id})"/>')

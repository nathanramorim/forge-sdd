#!/usr/bin/env python3
"""Injeta os bots do Forge (mesmo desenho de site/src/components/MascotFallbackSvg.astro) num SVG do diagram-studio e regrava o PNG.
Uso: uv run --with resvg-py docs/diagramas/bots.py <svg> <largura> <cor:marca:cx:y[:escala]>..."""
import sys, re
from pathlib import Path
import resvg_py

MARCAS = {
 "bar": '<rect x="52" y="49.7" width="16" height="4.6" rx="2.3" fill="rgba(18,24,48,0.6)"/>',
 "ring": '<circle cx="60" cy="52" r="4.6" fill="none" stroke="rgba(18,24,48,0.6)" stroke-width="2.6"/>',
 "wave": '<path d="M51 53 Q55.5 47 60 53 Q64.5 59 69 53" fill="none" stroke="rgba(18,24,48,0.6)" stroke-width="2.6" stroke-linecap="round"/>',
 "tri": '<polygon points="60,46.5 66,57 54,57" fill="rgba(18,24,48,0.6)"/>',
}
def bot(cor, marca, cx, y, esc=0.6):
    orb = "#4fd1b5"
    return f'''<g transform="translate({cx-60*esc:.1f},{y}) scale({esc})">
<ellipse cx="60" cy="127" rx="32" ry="4.5" fill="rgba(0,0,0,0.38)"/>
<ellipse cx="13" cy="90" rx="7" ry="11" fill="{cor}"/><ellipse cx="107" cy="90" rx="7" ry="11" fill="{cor}"/>
<line x1="60" y1="30" x2="60" y2="17" stroke="{cor}" stroke-width="3.2" stroke-linecap="round"/>
<circle cx="60" cy="11" r="14" fill="{orb}" opacity="0.3"/><circle cx="60" cy="11" r="7.8" fill="{orb}"/>
<path d="M60 28 C86 28 102 52 105 80 C108 106 92 118 60 118 C28 118 12 106 15 80 C18 52 34 28 60 28 Z" fill="{cor}"/>
<path d="M15 80 C12 106 28 118 60 118 C92 118 108 106 105 80 C97 98 80 106 60 106 C40 106 23 98 15 80 Z" fill="rgba(0,0,0,0.22)"/>
{MARCAS[marca]}
<rect x="25" y="70" width="70" height="32" rx="16" fill="#10131a" stroke="rgba(255,255,255,0.1)"/>
<g fill="{orb}"><rect x="41.5" y="78" width="7" height="16" rx="3.5"/><rect x="71.5" y="78" width="7" height="16" rx="3.5"/></g>
</g>'''
svg = Path(sys.argv[1]); larg = int(sys.argv[2])
s = svg.read_text()
g = "".join(bot(*(lambda a: (a[0], a[1], float(a[2]), float(a[3]), float(a[4]) if len(a) > 4 else 0.6))(x.split(":"))) for x in sys.argv[3:])
s = s.replace("</svg>", f'<g aria-hidden="true">{g}</g></svg>')
svg.write_text(s)
svg.with_suffix(".png").write_bytes(bytes(resvg_py.svg_to_bytes(svg_string=s, width=larg * 2)))
print("ok", svg)

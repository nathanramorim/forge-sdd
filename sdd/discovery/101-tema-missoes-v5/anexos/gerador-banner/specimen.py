import sys
from parts import *

W, H = 1280, 1330
WINE = c("superficie", "vinho")
SOFT = "#e7c9d5"
SCRIPT = FONTS["script"]["pilha"]

def label(x, y, text, size=20):
    return f'<text x="{x}" y="{y}" font-family="{SLAB}" font-size="{size}" fill="#ffffff">{text}</text>'

def note(x, y, text, anchor="start", size=12):
    return f'<text x="{x}" y="{y}" text-anchor="{anchor}" font-family="{FONT}" font-weight="500" font-size="{size}" fill="{SOFT}">{text}</text>'

def build():
    o = []
    o.append(f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" role="img" '
             f'aria-label="Espécime do design system GTA VI do forge-sdd: tipografia, paleta, degradês, cores e componentes.">')
    o.append('<title>Design system GTA VI — espécime</title>')
    o.append('<defs>')
    for gid, name in [("gSig", "assinatura"), ("gChip", "chip"), ("gPr", "prancha"), ("gSky", "ceu"), ("gS1", "sol-crepusculo"), ("gS2", "sol-orquidea"), ("gS3", "sol-ambar")]:
        o.append(grad(gid, name))
    o.append('<linearGradient id="gSkyH" x1="0" y1="0" x2="1" y2="0">' + ''.join(f'<stop offset="{p / 100:g}" stop-color="{col}"/>' for p, col in GRADIENTS["ceu"][2]) + '</linearGradient>')
    o.append('<filter id="fGlow" x="-5%" y="-400%" width="110%" height="900%"><feGaussianBlur stdDeviation="3"/></filter>')
    o.append('<clipPath id="cHead"><rect x="0" y="0" width="1280" height="150"/></clipPath>')
    o.append('<clipPath id="cBody"><rect x="40" y="996" width="1200" height="222"/></clipPath>')
    o.append('</defs>')
    o.append(f'<rect width="{W}" height="{H}" fill="{WINE}"/>')

    # ---------- cabeçalho
    o.append('<g clip-path="url(#cHead)">')
    o.append('<rect width="1280" height="150" fill="url(#gPr)"/>')
    o.append(palm(1088, 156, 118, -16, 0.86, c("superficie", "ameixa"), 0.42))
    o.append(palm(1190, 156, 150, 14, 1.05, c("superficie", "ameixa"), 0.5))
    o.append('</g>')
    o.append(f'<text x="40" y="84" font-family="{HEAVY}" font-weight="900" font-size="44" letter-spacing="0.5" fill="#ffffff">DESIGN SYSTEM GTA VI</text>')
    o.append(f'<text x="40" y="116" font-family="{FONT}" font-weight="600" font-size="15" fill="#ffffff">forge-sdd 5.0 · tokens medidos nas pranchas do conceito de UI de Matías Rodríguez (Behance, 2025)</text>')

    # ---------- tipografia
    o.append(label(40, 200, "Tipografia"))
    o.append(f'<text x="40" y="244" font-family="{HEAVY}" font-weight="900" font-size="26" letter-spacing="0.6" fill="#ffffff" stroke="{c("superficie", "preto-moldura")}" stroke-width="5" stroke-linejoin="round" paint-order="stroke fill">DISCOVERY · BUILDER · REVIEWER</text>')
    o.append(note(40, 266, "Display · Pricedown no conceito · botões, abas, menus em jogo e dinheiro"))
    o.append(f'<text x="40" y="312" font-family="{FONT}" font-weight="700" font-size="22" letter-spacing="1.8" fill="#ffffff">SEU TIME DE IA, COM MÉTODO.</text>')
    o.append(f'<text x="40" y="336" font-family="{FONT}" font-weight="500" font-size="15" fill="#ffffff">Do plano ao PR, sem repetir instruções.</text>')
    o.append(note(40, 358, "Títulos e corpo · GTAArtDeco no conceito · título em caixa alta espaçada"))
    o.append(f'<text x="40" y="402" font-family="{SLAB}" font-size="19" fill="#ffffff">Rótulo de seção</text>')
    o.append(f'<text x="212" y="404" font-family="{SCRIPT}" font-size="30" fill="#ffffff">Tela de loading</text>')
    o.append(note(40, 424, "Rótulos em serifa egípcia e títulos de prancha em script (observados, não declarados)"))
    o.append(word("forge-sdd", 40, 448, scale=0.56))
    o.append(note(40, 506, "Letreiro do forge-sdd · desenho próprio em bloco, sem arquivo de fonte"))

    # ---------- paleta (sóis) e degradês
    o.append(label(700, 200, "Paleta"))
    suns = [("gS1", "crepúsculo", "#4053c6 → #f161ae → #cc7743"), ("gS2", "orquídea", "#e446c4 → #f35caf → #e8698e"), ("gS3", "âmbar", "#de8640 → #e28642 → #cd7745")]
    for i, (gid, name, rng) in enumerate(suns):
        cx = 770 + i * 186
        o.append(striped_sun(f"mS{i}", cx, 270, 46, f"url(#{gid})"))
        o.append(f'<text x="{cx}" y="340" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="13" fill="#ffffff">sol {name}</text>')
        o.append(f'<text x="{cx}" y="357" text-anchor="middle" font-family="{MONO}" font-size="10" fill="{SOFT}">{rng}</text>')
    bars = [("gSig", "assinatura", "linha neon das molduras", "#ffaa6e → #fe818d → #f061af"),
            ("gChip", "chip", "aba ou seção ativa", "#f7a164 → #f77191 → #e850af"),
            ("gPr", "prancha", "fundo atrás das molduras", "#fea96a → #f87795 → #a780a6"),
            ("gSkyH", "céu", "vertical na arte; aqui deitado", "#5764b0 → #ca84a9 → #ffb06d")]
    for i, (gid, name, use, rng) in enumerate(bars):
        y = 384 + i * 36
        o.append(f'<rect x="700" y="{y}" width="230" height="18" rx="3" fill="url(#{gid})"/>')
        o.append(f'<text x="942" y="{y + 13}" font-family="{FONT}" font-size="12" fill="#ffffff"><tspan font-weight="700">{name}</tspan><tspan fill="{SOFT}"> · {use}</tspan></text>')
        o.append(f'<text x="1240" y="{y + 30}" text-anchor="end" font-family="{MONO}" font-size="10" fill="{SOFT}">{rng}</text>')

    # ---------- cores
    o.append(label(40, 566, "Cores"))
    rows = [("Superfícies", "superficie"), ("Texto", "texto"), ("Acentos", "acento"), ("Apoio", "apoio")]
    for r, (title, key) in enumerate(rows):
        y = 590 + r * 86
        o.append(f'<text x="40" y="{y + 26}" font-family="{SLAB}" font-size="15" fill="{SOFT}">{title}</text>')
        for i, (name, (hexv, _)) in enumerate(TOKENS[key].items()):
            x = 150 + i * 91
            o.append(f'<rect x="{x}" y="{y}" width="83" height="40" rx="4" fill="{hexv}" stroke="#ffffff" stroke-opacity="0.22"/>')
            o.append(f'<text x="{x}" y="{y + 55}" font-family="{FONT}" font-weight="600" font-size="10.5" fill="#ffffff">{name}</text>')
            o.append(f'<text x="{x}" y="{y + 69}" font-family="{MONO}" font-size="10" fill="{SOFT}">{hexv}</text>')

    # ---------- componentes (mini tela de ajustes)
    o.append(label(40, 950, "Componentes"))
    fx, fy, fw = 40, 962, 1200
    o.append(f'<rect x="{fx}" y="{fy}" width="{fw}" height="290" fill="{c("superficie", "preto-fundo")}"/>')
    o.append('<g clip-path="url(#cBody)">')
    o.append('<rect x="40" y="996" width="1200" height="222" fill="url(#gSky)"/>')
    o.append(palm(96, 1226, 190, 22, 1.1, c("superficie", "ameixa"), 0.38))
    o.append('</g>')
    # barras e linhas
    o.append(f'<rect x="{fx}" y="{fy}" width="{fw}" height="34" fill="{c("superficie", "preto-moldura")}"/>')
    o.append(f'<rect x="{fx}" y="1218" width="{fw}" height="34" fill="{c("superficie", "preto-moldura")}"/>')
    for y in (996, 1216):
        o.append(f'<rect x="{fx}" y="{y - 1}" width="{fw}" height="4" fill="url(#gSig)" opacity="0.7" filter="url(#fGlow)"/>')
        o.append(f'<rect x="{fx}" y="{y}" width="{fw}" height="2" fill="url(#gSig)"/>')
    # abas
    tab = f'font-family="{HEAVY}" font-weight="900" font-size="11" letter-spacing="0.8" fill="#ffffff" text-anchor="middle"'
    o.append(keycap(894, 969, "L1"))
    o.append('<rect x="928" y="968" width="84" height="22" rx="3" fill="url(#gChip)"/>')
    o.append(f'<text x="970" y="983" {tab} stroke="#151515" stroke-opacity="0.5" stroke-width="2.4" stroke-linejoin="round" paint-order="stroke fill">AJUSTES</text><text x="1056" y="983" {tab}>FILA</text><text x="1134" y="983" {tab}>SESSÕES</text>')
    o.append(keycap(1196, 969, "R1"))
    # navegação
    o.append('<rect x="70" y="1016" width="176" height="28" rx="2" fill="url(#gChip)"/>')
    o.append(f'<text x="84" y="1035" font-family="{FONT}" font-weight="800" font-size="13" letter-spacing="1" fill="#ffffff" stroke="#151515" stroke-opacity="0.5" stroke-width="2.4" stroke-linejoin="round" paint-order="stroke fill">AJUSTES</text>')
    for i, (name, active) in enumerate([("Sessões", False), ("Agentes", True), ("Tetos", False), ("Telemetria", False)]):
        y = 1050 + i * 30
        o.append(f'<rect x="70" y="{y}" width="176" height="26" fill="{c("superficie", "faixa-inativa")}" fill-opacity="0.78"/>')
        col = c("texto", "ativo") if active else c("texto", "inativo")
        o.append(f'<text x="84" y="{y + 17.5}" font-family="{FONT}" font-weight="{700 if active else 600}" font-size="12.5" fill="{col}">{name}</text>')
    # lista de opções
    opts = [("Agente da Discovery", "CLAUDE", False), ("Agente do Builder", "CODEX", True), ("Agente do Reviewer", "GEMINI", False), ("Permissões", "FULL", False), ("Teto de tokens", "OFF", False)]
    for i, (name, val, active) in enumerate(opts):
        y = 1016 + i * 34
        if active:
            o.append(f'<rect x="266" y="{y}" width="420" height="30" fill="{c("acento", "ambar")}"/>')
        else:
            o.append(f'<rect x="266" y="{y}" width="420" height="30" fill="{c("superficie", "preto-fundo")}" fill-opacity="0.42"/>')
        o.append(f'<text x="280" y="{y + 19.5}" font-family="{FONT}" font-weight="700" font-size="12.5" fill="#ffffff">{name}</text>')
        o.append(f'<rect x="566" y="{y + 7}" width="16" height="16" fill="#ffffff" stroke="{c("superficie", "preto-moldura")}" stroke-width="1.2"/>')
        o.append(f'<path d="M{577},{y + 10.5} L{571},{y + 15} L{577},{y + 19.5} Z" fill="{c("superficie", "preto-moldura")}"/>')
        o.append(f'<text x="674" y="{y + 19.5}" text-anchor="end" font-family="{FONT}" font-weight="700" font-size="11.5" letter-spacing="0.8" fill="#ffffff">{val}</text>')
    # painel de descrição
    o.append(f'<rect x="706" y="1016" width="250" height="26" fill="{c("superficie", "preto-fundo")}" fill-opacity="0.6"/>')
    o.append(f'<text x="831" y="1033.5" text-anchor="middle" font-family="{FONT}" font-weight="700" font-size="11" letter-spacing="1.4" fill="#ffffff">DESCRIÇÃO</text>')
    o.append(f'<rect x="706" y="1046" width="250" height="136" fill="{c("superficie", "painel")}" fill-opacity="0.9"/>')
    for i, line in enumerate(["Escolhe o agente que roda em cada", "sessão do pipeline. O Reviewer só", "lê: quem confere não é quem", "construiu."]):
        o.append(f'<text x="720" y="{1068 + i * 17}" font-family="{FONT}" font-weight="500" font-size="12" fill="#ffffff">{line}</text>')
    for i, line in enumerate(["Trocar o agente no meio de um job", "reinicia a sessão."]):
        o.append(f'<text x="720" y="{1148 + i * 16}" font-family="{FONT}" font-weight="600" font-size="12" fill="{c("texto", "aviso")}">{line}</text>')
    # HUD
    hx = 996
    for i, (col, icon) in enumerate([(c("acento", "rosa-vida"), "heart"), (c("acento", "laranja-energia"), "bolt")]):
        cx, cy = hx + 18 + i * 52, 1036
        o.append(f'<circle cx="{cx}" cy="{cy}" r="16" fill="{c("superficie", "disco")}"/>')
        o.append(f'<circle cx="{cx}" cy="{cy}" r="19.5" fill="none" stroke="{col}" stroke-width="3" stroke-dasharray="38 200" transform="rotate(-80 {cx} {cy})" stroke-linecap="round"/>')
        if icon == "heart":
            o.append(f'<path transform="translate({cx} {cy - 5})" d="M0,3.2 C-5,-4.4 -11.5,1.6 0,11 C11.5,1.6 5,-4.4 0,3.2 Z" fill="{col}"/>')
        else:
            o.append(f'<path transform="translate({cx} {cy})" d="M2,-9 L-6,1.5 H-0.6 L-2.4,9 L6,-1.8 H0.6 Z" fill="{col}"/>')
    for i in range(5):
        o.append(star(hx + 124 + i * 20, 1030, 8.5, c("apoio", "estrela-apagada") if i < 3 else "#ffffff"))
    o.append(f'<text x="{hx + 214}" y="1064" text-anchor="end" font-family="{HEAVY}" font-weight="900" font-size="17" fill="#ffffff" stroke="{c("superficie", "preto-moldura")}" stroke-width="4" stroke-linejoin="round" paint-order="stroke fill">$5000</text>')
    # radar
    rcx, rcy = hx + 46, 1136
    o.append(f'<circle cx="{rcx}" cy="{rcy}" r="42" fill="#8d8b93"/>')
    o.append(f'<path d="M{rcx - 40},{rcy - 8} L{rcx + 40},{rcy + 12} M{rcx - 10},{rcy - 40} L{rcx + 4},{rcy + 41}" stroke="#bdbcc4" stroke-width="7" fill="none" clip-path="none" opacity="0.9"/>')
    o.append(f'<circle cx="{rcx}" cy="{rcy}" r="42" fill="none" stroke="{WINE}" stroke-width="10" opacity="0"/>')
    o.append(f'<circle cx="{rcx + 12}" cy="{rcy - 12}" r="13" fill="#b5403f" fill-opacity="0.55"/>')
    for dx, dy in [(8, -16), (15, -10), (11, -6), (18, -15)]:
        o.append(f'<circle cx="{rcx + dx}" cy="{rcy + dy}" r="2.4" fill="#a8272f" stroke="#5c1016" stroke-width="0.8"/>')
    o.append(f'<path d="M{rcx - 6},{rcy + 14} l5,-11 l5,11 l-5,-3 Z" fill="#ffffff" stroke="{c("superficie", "preto-moldura")}" stroke-width="0.8"/>')
    o.append(f'<circle cx="{rcx}" cy="{rcy}" r="42" fill="none" stroke="url(#gSig)" stroke-width="3"/>')
    # slots
    for i in range(3):
        x = hx + 104 + i * 46
        sel = i == 1
        o.append(f'<rect x="{x}" y="1112" width="38" height="38" rx="5" fill="{c("apoio", "salvia-escura")}" fill-opacity="0.92"'
                 + (f' stroke="{c("acento", "rosa-vida")}" stroke-width="2.2"' if sel else '') + '/>')
        o.append(f'<rect x="{x + 12}" y="1126" width="14" height="11" rx="2" fill="#ffffff" fill-opacity="{0.95 if sel else 0.55}"/>')
        o.append(f'<rect x="{x + 16}" y="1122" width="6" height="5" rx="1.5" fill="none" stroke="#ffffff" stroke-opacity="{0.95 if sel else 0.55}" stroke-width="1.4"/>')
    o.append(note(hx + 104, 1170, "slots de ação", size=10.5))
    # dicas
    hint = f'font-family="{FONT}" font-weight="600" font-size="12.5" fill="#ffffff" text-anchor="end"'
    o.append(f'<text x="924" y="1240" {hint}>Restaurar padrão</text>' + face_button(942, 1235, "t", 9))
    o.append(f'<text x="1040" y="1240" {hint}>Voltar</text>' + face_button(1058, 1235, "o", 9))
    o.append(f'<text x="1186" y="1240" {hint}>Selecionar</text>' + face_button(1204, 1235, "x", 9))

    # ---------- rodapé
    o.append(note(40, 1288, "Homenagem não oficial, sem vínculo com a Rockstar Games ou a Take-Two. Nenhuma arte, logotipo ou fonte do jogo ou do conceito é distribuída aqui."))
    o.append(note(40, 1308, "Cores amostradas das pranchas publicadas (tolerância de ±2 por canal). Fonte: behance.net/gallery/224512205"))
    o.append('</svg>')
    return '\n'.join(o) + '\n'

if __name__ == "__main__":
    open(sys.argv[1], "w", encoding="utf-8").write(build())

#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Renderiza todos os diagramas Mermaid de um arquivo markdown em PNG, com repetição automática.

Uso:
    uv run scripts/renderizar-mermaid.py ARQUIVO.md --saida PASTA [--escala 1.5] [--nomes nome1,nome2,...]

O que faz: extrai cada bloco ```mermaid, grava PASTA/<nome>.mmd e PASTA/<nome>.png, e imprime OK ou FALHOU por diagrama.
Instala o mermaid-cli uma vez em ~/.cache/mermaid-cli (ou em $MMD_DIR) usando o Chrome do sistema; não baixa outro navegador.
Depende de fora do Python: Node.js com npm e Google Chrome (ou Chromium). O uv não cobre esse pedaço.
Sai com código 1 se algum diagrama falhar. Depois de renderizar, ABRA os PNG e confira as setas."""
import argparse, json, os, pathlib, re, shutil, subprocess, sys

CACHE = pathlib.Path(os.environ.get("MMD_DIR", pathlib.Path.home() / ".cache" / "mermaid-cli"))
CHROMES = ["/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "/Applications/Chromium.app/Contents/MacOS/Chromium",
           shutil.which("google-chrome") or "", shutil.which("chromium") or "", shutil.which("chromium-browser") or ""]


def chrome():
    for c in CHROMES:
        if c and pathlib.Path(c).exists(): return c
    sys.exit("Google Chrome ou Chromium não encontrado. Instale um dos dois (o uv não cobre isso).")


def garantir_mmdc():
    exe = CACHE / "node_modules" / ".bin" / "mmdc"
    if exe.exists(): return exe
    if not shutil.which("npm"): sys.exit("npm não encontrado. Instale o Node.js (o uv não cobre isso).")
    CACHE.mkdir(parents=True, exist_ok=True)
    if not (CACHE / "package.json").exists(): subprocess.run(["npm", "init", "-y"], cwd=CACHE, capture_output=True)
    print("Instalando o mermaid-cli (uma vez, alguns minutos)...", flush=True)
    r = subprocess.run(["npm", "install", "--silent", "@mermaid-js/mermaid-cli"], cwd=CACHE, env={**os.environ, "PUPPETEER_SKIP_DOWNLOAD": "1"}, capture_output=True, text=True)
    if not exe.exists(): sys.exit(f"Falha ao instalar o mermaid-cli:\n{r.stderr[-600:]}")
    return exe


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("arquivo"); ap.add_argument("--saida", required=True); ap.add_argument("--escala", default="1.5"); ap.add_argument("--nomes", default="")
    a = ap.parse_args()
    blocos = re.findall(r"```mermaid\n(.*?)```", pathlib.Path(a.arquivo).read_text(encoding="utf8"), re.S)
    if not blocos: sys.exit("Nenhum bloco ```mermaid encontrado.")
    nomes = [n.strip() for n in a.nomes.split(",") if n.strip()]
    saida = pathlib.Path(a.saida); saida.mkdir(parents=True, exist_ok=True)
    mmdc = garantir_mmdc()
    cfg = saida / ".puppeteer.json"; cfg.write_text(json.dumps({"executablePath": chrome(), "args": ["--no-sandbox"]}))
    falhas = 0
    for i, b in enumerate(blocos, 1):
        nome = nomes[i - 1] if i <= len(nomes) else f"{i:02d}-diagrama"
        mmd, png = saida / f"{nome}.mmd", saida / f"{nome}.png"
        mmd.write_text(b, encoding="utf8"); png.unlink(missing_ok=True); ok = False
        for _ in range(3):
            subprocess.run([str(mmdc), "-p", str(cfg), "-i", str(mmd), "-o", str(png), "-s", a.escala, "-b", "white"], capture_output=True, text=True)
            if png.exists() and png.stat().st_size > 0: ok = True; break
        print(f"[{'OK   ' if ok else 'FALHOU'}] {nome:<45} {png.stat().st_size if ok else 0} bytes")
        falhas += 0 if ok else 1
    cfg.unlink(missing_ok=True)
    print(f"\n{len(blocos)} diagramas | {len(blocos) - falhas} OK | {falhas} falhas. Agora abra os PNG e confira setas, rótulos e distribuição.")
    sys.exit(1 if falhas else 0)


if __name__ == "__main__":
    main()

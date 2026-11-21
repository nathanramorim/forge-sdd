#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""Gerador das novidades da LP a partir do histórico de releases (feat-112-03).

Uso:
  uv run scripts/novidades_lp.py [--dir RAIZ] [--check] [--quiet]

Lê `sdd/releases/history.md` e grava `docs/data/releases.json` com as duas versões
que a LP mostra no topo: a estável mais recente e a beta mais recente.

  --check  não grava nada; sai com 1 se o JSON versionado divergir do histórico
  --quiet  não imprime o resumo

O histórico é dado não confiável (SEG-01, SEG-02): seções sem bullets são recusadas,
caracteres de controle e `..` em campos textuais são removidos, e nenhum campo
vira HTML — a LP insere tudo com textContent.
"""
import argparse
import json
import os
import re
import subprocess
import sys
from pathlib import Path

HISTORY_REL = Path("sdd") / "releases" / "history.md"
JSON_REL = Path("docs") / "data" / "releases.json"

INI = "### Versão "
RE_SECAO = re.compile(r"^###\s+Versão\s+(\d+\.\d+\.\d+(?:-[0-9A-Za-z.]+)?)\s*[—-]\s*(.+?)\s*$")
RE_BULLET = re.compile(r"^\*\s+(.+)$")
RE_BOLD = re.compile(r"\*\*(.+?)\*\*")
RE_CONTROLE = re.compile(r"[\x00-\x08\x0b-\x1f\x7f-\x9f]")
RE_ESTAVEL = re.compile(r"^\d+\.\d+\.\d+$")
RE_BETA = re.compile(r"^\d+\.\d+\.\d+-(?:beta|rc)(?:\.\d+)?$")
MAX_TEXTO = 600


class ErroHistorico(Exception):
    pass


def limpar(texto: str) -> str:
    """Tira caracteres de controle e espaços redundantes de um campo textual."""
    if not isinstance(texto, str):
        return ""
    sem_ctrl = RE_CONTROLE.sub("", texto)
    return " ".join(sem_ctrl.split())[:MAX_TEXTO]


def canal_de(versao: str, titulo: str = "") -> str:
    """Canal da seção: o sufixo da versão ou o marcador "(Beta)"/"(Estável)" do título.

    O histórico nomeia as seções sem o sufixo da versão — a 5.1.0-beta aparece como
    "### Versão 5.1.0 — ... (Beta)". O marcador do título é a fonte para esse caso.
    """
    if RE_ESTAVEL.match(versao):
        if re.search(r"\((?:beta|rc)[^)]*\)\s*$", titulo, re.IGNORECASE):
            return "beta"
        return "stable"
    if RE_BETA.match(versao):
        return "beta"
    return ""


def rotulo_de(versao: str, canal: str) -> str:
    # O rótulo é o que a LP mostra e o que o npm publica.
    if canal == "beta" and "-" not in versao:
        return versao + "-beta"
    return versao


def separar_item(bullet: str) -> dict:
    """Divide um bullet em destaque (negrito) e explicação."""
    texto = limpar(bullet)
    m = RE_BOLD.search(texto)
    if not m:
        return {"titulo": "", "texto": texto}
    titulo = limpar(m.group(1)).rstrip(":")
    resto = limpar((texto[: m.start()] + " " + texto[m.end() :]))
    return {"titulo": titulo, "texto": resto.lstrip(":—–- ").strip()}


def parse_historico(conteudo: str) -> list:
    # Recusa qualquer caminho com `..` no texto lido: é sinal de conteúdo hostil.
    if ".." in conteudo and "../" in conteudo:
        raise ErroHistorico("histórico contém caminho com '..'")

    versoes = []
    atual = None
    for linha in conteudo.splitlines():
        m = RE_SECAO.match(linha)
        if m:
            if atual is not None:
                versoes.append(atual)
            numero, titulo = m.group(1), limpar(m.group(2))
            canal = canal_de(numero, titulo)
            atual = {
                "versao": numero,
                "rotulo": rotulo_de(numero, canal),
                "canal": canal,
                "titulo": titulo,
                "itens": [],
            }
            continue
        if atual is None:
            continue
        if linha.startswith("### ") or linha.startswith("## "):
            versoes.append(atual)
            atual = None
            continue
        mb = RE_BULLET.match(linha.strip())
        if mb:
            item = separar_item(mb.group(1))
            if item["titulo"] or item["texto"]:
                atual["itens"].append(item)
    if atual is not None:
        versoes.append(atual)

    if not versoes:
        raise ErroHistorico("nenhuma seção '### Versão' encontrada")
    return versoes


def escolher_duas(versoes: list) -> list:
    """A estável mais recente e a beta mais recente, na ordem em que a LP mostra."""
    estavel = next((v for v in versoes if v["canal"] == "stable"), None)
    beta = next((v for v in versoes if v["canal"] == "beta"), None)
    escolhidas = [v for v in (estavel, beta) if v is not None]
    if not escolhidas:
        raise ErroHistorico("histórico não tem versão estável nem beta")
    # Só as duas publicadas precisam ter destaques reais; o histórico antigo tem
    # seções curtas sem bullets, que ficam fora do JSON.
    sem_bullets = [v["versao"] for v in escolhidas if not v["itens"]]
    if sem_bullets:
        raise ErroHistorico("versão publicada sem bullets: " + ", ".join(sem_bullets))
    return escolhidas


def sha_curto(raiz: Path) -> str:
    try:
        out = subprocess.run(
            ["git", "-C", str(raiz), "rev-parse", "--short", "HEAD"],
            capture_output=True, text=True, check=False, timeout=5,
        )
    except (OSError, subprocess.SubprocessError):
        return ""
    return out.stdout.strip() if out.returncode == 0 else ""


def gerar(raiz: Path) -> dict:
    hist = raiz / HISTORY_REL
    if not hist.is_file():
        raise ErroHistorico("arquivo %s não encontrado" % HISTORY_REL)
    try:
        conteudo = hist.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError) as exc:
        raise ErroHistorico("falha ao ler %s: %s" % (HISTORY_REL, exc))

    versoes = escolher_duas(parse_historico(conteudo))
    dados = {
        # gerado_em só muda quando o conteúdo muda (ver `chave_conteudo`)
        "chave_conteudo": chave_conteudo(versoes),
        "versao_historico": sha_curto(raiz),
        "versoes": versoes,
    }
    return dados


def chave_conteudo(versoes: list) -> str:
    """Impressão digital do conteúdo: mantém o arquivo idempotente entre execuções."""
    import hashlib

    payload = json.dumps(versoes, ensure_ascii=False, sort_keys=True)
    return hashlib.sha256(payload.encode("utf-8")).hexdigest()[:16]


def montar_saida(raiz: Path) -> str:
    dados = gerar(raiz)
    destino = raiz / JSON_REL
    anterior = None
    if destino.is_file():
        try:
            anterior = json.loads(destino.read_text(encoding="utf-8"))
        except (OSError, ValueError):
            anterior = None
    if anterior and anterior.get("chave_conteudo") == dados["chave_conteudo"]:
        # Conteúdo igual: preserva o arquivo como está (sem ruído no diff).
        return destino.read_text(encoding="utf-8")
    dados["gerado_em"] = os.environ.get("SOURCE_DATE_EPOCH_ISO", "")
    saida = {
        "schema": 1,
        "chave_conteudo": dados["chave_conteudo"],
        "versao_historico": dados["versao_historico"],
        "versoes": [
            {
                "versao": v["versao"],
                "rotulo": v["rotulo"],
                "canal": v["canal"],
                "titulo": v["titulo"],
                "itens": [
                    {"titulo": i["titulo"], "texto": i["texto"]} for i in v["itens"]
                ],
            }
            for v in dados["versoes"]
        ],
    }
    return json.dumps(saida, ensure_ascii=False, indent=2) + "\n"


def main(argv=None) -> int:
    ap = argparse.ArgumentParser(description="Gera docs/data/releases.json a partir do histórico.")
    ap.add_argument("--dir", default=".", help="raiz do projeto (padrão: diretório atual)")
    ap.add_argument("--check", action="store_true", help="não grava; sai 1 se divergir")
    ap.add_argument("--quiet", action="store_true", help="não imprime o resumo")
    args = ap.parse_args(argv)

    raiz = Path(args.dir).resolve()
    try:
        saida = montar_saida(raiz)
    except ErroHistorico as exc:
        print("novidades_lp: %s" % exc, file=sys.stderr)
        return 1

    destino = raiz / JSON_REL
    atual = destino.read_text(encoding="utf-8") if destino.is_file() else ""

    if args.check:
        if saida != atual:
            print("novidades_lp: %s diverge de %s; rode sem --check" % (JSON_REL, HISTORY_REL), file=sys.stderr)
            return 1
        if not args.quiet:
            print("novidades_lp: %s em dia com %s" % (JSON_REL, HISTORY_REL))
        return 0

    if saida == atual:
        if not args.quiet:
            print("novidades_lp: %s sem mudança" % JSON_REL)
        return 0

    destino.parent.mkdir(parents=True, exist_ok=True)
    temporario = destino.with_suffix(".json.tmp")
    temporario.write_text(saida, encoding="utf-8")
    os.replace(temporario, destino)
    if not args.quiet:
        n = sum(len(v.get("itens", [])) for v in json.loads(saida)["versoes"])
        print("novidades_lp: %s gravado (%d versões, %d itens)" % (JSON_REL, len(json.loads(saida)["versoes"]), n))
    return 0


if __name__ == "__main__":
    sys.exit(main())
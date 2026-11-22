#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""Verifica o atestado de um PR no CI (feat-106-32).

Este script roda no job do CI, **só no PR de feature** e, no modo `local`, nos
segundos. Ele lê o parecer do Reviewer gravado no repositório e confere:

  - o modo do atestado lido da Constituição;
  - o `head_sha` do parecer é o commit do PR (o código não mudou depois);
  - o veredito é `conforme`;
  - o PR não é o da spec, se o ponto for `spec` (o PR da spec é sempre `local`
    e dispensa a prova do VCS, mas não dispensa a aprovação do humano).

Uso:
  uv run scripts/atestado.py --verificar --sha <sha> --modo local
  uv run scripts/atestado.py --verificar --sha HEAD --modo desligado
  uv run scripts/atestado.py --verificar --sha <sha> --modo vcs   # recusado: adiado

Garantias:
  - Somente leitura (SEG-04): nenhum arquivo é gravado.
  - Sem rede: lê só o que está no checkout (SEG-05).
  - Modo desconhecido e SHA malformado são recusados na borda.
"""
import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

CONSTITUICAO = Path("sdd") / "memory" / "constitution.md"
PARECER_PADRAO = Path("sdd") / "memory" / "parecer.json"

RE_SHA = re.compile(r"^[0-9a-f]{7,64}$")
RE_CAMPO = re.compile(r"^\|\s*([^|]+?)\s*\|\s*([^|]*?)\s*\|\s*$")

MODOS = {"desligado", "local", "vcs"}
VEREDITOS = {"conforme", "violacao", "duvida"}
PONTOS = {"spec", "entrega"}


class Recusa(Exception):
    """Recusa do atestado; a mensagem explica o que fazer."""


def ler_modo_constituicao(caminho=CONSTITUICAO):
    """Lê `Modo do atestado` da tabela da Constituição; ausente vale `local`."""
    p = Path(caminho)
    if not p.is_file():
        return "local"
    try:
        texto = p.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError) as exc:
        raise Recusa("não consegui ler %s: %s" % (caminho, exc))
    for linha in texto.splitlines():
        m = RE_CAMPO.match(linha)
        if not m:
            continue
        if m.group(1).strip() == "Modo do atestado":
            valor = m.group(2).strip()
            return valor or "local"
    return "local"


def ler_parecer(caminho):
    p = Path(caminho)
    if not p.is_file():
        raise Recusa("parecer ausente em %s: rode o Reviewer antes de marcar o PR como pronto" % caminho)
    try:
        dados = json.loads(p.read_text(encoding="utf-8"))
    except (OSError, ValueError) as exc:
        raise Recusa("parecer ilegível em %s: %s" % (caminho, exc))
    if not isinstance(dados, dict):
        raise Recusa("parecer não é um objeto JSON")
    return dados


def sha_do_head():
    try:
        out = subprocess.run(
            ["git", "rev-parse", "HEAD"], capture_output=True, text=True, check=False, timeout=10
        )
    except (OSError, subprocess.SubprocessError) as exc:
        raise Recusa("não consegui ler o HEAD: %s" % exc)
    sha = out.stdout.strip().lower()
    if out.returncode != 0 or not RE_SHA.match(sha):
        raise Recusa("não consegui resolver o SHA do HEAD")
    return sha


def conferir(parecer, sha, modo):
    """Aplica o atestado conhecido; devolve uma linha de decisão."""
    veredito = str(parecer.get("veredito") or "").strip().lower()
    head = str(parecer.get("head_sha") or "").strip().lower()
    ponto = str(parecer.get("ponto") or "").strip().lower()

    if veredito not in VEREDITOS:
        raise Recusa("veredito %r desconhecido no parecer" % veredito)
    if ponto and ponto not in PONTOS:
        raise Recusa("ponto %r desconhecido no parecer (esperado spec ou entrega)" % ponto)

    # Desligado: nada é conferido, e isso é dito em voz alta.
    if modo == "desligado":
        return "modo desligado: nada foi conferido; o PR pode sair sem conformidade"

    # O PR da spec é sempre local e não depende da prova do VCS (Decisão xxi).
    if ponto == "spec":
        if veredito != "conforme":
            raise Recusa("parecer da spec não é conforme (veredito %r)" % veredito)
        return "PR da spec: conferência 1 conforme; atestado local satisfeito"

    if veredito != "conforme":
        raise Recusa("parecer não é conforme (veredito %r): o PR não pode sair do draft" % veredito)
    if not RE_SHA.match(head):
        raise Recusa("parecer sem head_sha válido (%r): recusado" % head)
    if head != sha:
        raise Recusa(
            "o parecer é de outro commit (parecer em %s, HEAD em %s): reexecute o Reviewer"
            % (head[:7], sha[:7])
        )
    return "atestado local válido: parecer conforme para o commit %s" % sha[:7]


def main(argv=None):
    ap = argparse.ArgumentParser(description="Verifica o atestado do PR no CI.")
    ap.add_argument("--verificar", action="store_true", help="verifica o atestado e sai")
    ap.add_argument("--sha", default="HEAD", help="SHA a conferir (padrão: HEAD)")
    ap.add_argument("--modo", default=None, help="modo do atestado; padrão: lido da Constituição")
    ap.add_argument("--parecer", default=str(PARECER_PADRAO), help="caminho do parecer")
    ap.add_argument("--constituicao", default=str(CONSTITUICAO), help="caminho da Constituição")
    args = ap.parse_args(argv)

    if not args.verificar:
        print("atestado: use --verificar", file=sys.stderr)
        return 2

    try:
        modo = (args.modo or ler_modo_constituicao(args.constituicao) or "local").strip().lower()
        if modo not in MODOS:
            raise Recusa("modo do atestado %r desconhecido (use desligado, local ou vcs)" % modo)
        if modo == "vcs":
            raise Recusa("modo adiado: a prova pelo VCS não está disponível nesta entrega")

        # O SHA informado é resolvido e validado antes de qualquer leitura.
        if args.sha.strip().upper() == "HEAD":
            sha = sha_do_head()
        else:
            sha = args.sha.strip().lower()
            if not RE_SHA.match(sha):
                raise Recusa("SHA %r inválido: precisa ser hexadecimal com 7 a 64 caracteres" % args.sha)
            head = sha_do_head()
            if not head.startswith(sha):
                raise Recusa("o SHA %s não corresponde ao HEAD %s" % (sha[:7], head[:7]))

        parecer = ler_parecer(args.parecer)
        linha = conferir(parecer, sha, modo)
        print(linha)
        return 0
    except Recusa as exc:
        print("atestado: %s" % exc, file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
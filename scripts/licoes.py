#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""Gerenciador de lições por rule da Constituição (106, task 1.2).

Uso:
  uv run scripts/licoes.py --rule COD-02 [--dir RAIZ]
  uv run scripts/licoes.py --rule COD-02,DUP-10 [--dir RAIZ]
  uv run scripts/licoes.py --add --rule COD-02 --pattern "..." --fix "..." [--ref "..."] [--dir RAIZ]
  uv run scripts/licoes.py --add --rule COD-02 "padrão → correção (ref)" [--dir RAIZ]

Lê os IDs canônicos de `sdd/memory/constitution.md`. Recusa rule inexistente na
Constituição, path traversal (`..`) e caracteres de controle (SEG-01, SEG-02, SEG-08).

Códigos de saída: 0 sucesso; 1 ou 2 erro de validação/entrada.
"""
import argparse
import re
import sys
from pathlib import Path

RE_RULE = re.compile(r"^[A-Z]{3}-\d{2}$")
RE_RULE_FIND = re.compile(r"\b[A-Z]{3}-\d{2}\b")
RE_CTRL = re.compile(r"[\x00-\x1f\x7f-\x9f]")
ORCAMENTO_BYTES = 2 * 1024


class EntradaInvalida(Exception):
    """Erro de validação de entrada na borda."""


def limpo(valor):
    """Remove caracteres de controle de texto destinado à saída."""
    return RE_CTRL.sub("", str(valor))[:400]


def validar_sem_controle(valor, nome):
    if RE_CTRL.search(str(valor)):
        raise EntradaInvalida(f"{nome}: caracteres de controle recusados")
    return str(valor).strip()


def validar_caminho(valor, nome):
    validar_sem_controle(valor, nome)
    p = Path(valor)
    if ".." in p.parts:
        raise EntradaInvalida(f"{nome}: caminho com '..' recusado")
    return p


def regras_da_constituicao(raiz):
    arq = raiz / "sdd" / "memory" / "constitution.md"
    if not arq.is_file():
        raise EntradaInvalida(f"constituição ausente em {limpo(arq)}")
    return set(RE_RULE_FIND.findall(arq.read_text(encoding="utf-8")))


def parse_rules(valores, regras_validas):
    if not valores:
        raise EntradaInvalida("informe ao menos uma rule com --rule <ID>")
    ids = []
    for bruto in valores:
        validar_sem_controle(bruto, "--rule")
        for parte in bruto.split(","):
            rid = parte.strip()
            if not rid:
                continue
            if ".." in rid or "/" in rid or "\\" in rid or not RE_RULE.match(rid):
                raise EntradaInvalida(f"rule inválida: {limpo(rid)}")
            if rid not in regras_validas:
                raise EntradaInvalida(f"rule inexistente na Constituição: {limpo(rid)}")
            if rid not in ids:
                ids.append(rid)
    if not ids:
        raise EntradaInvalida("informe ao menos uma rule com --rule <ID>")
    return ids


def cabecalho_rule(rule_id):
    return (
        f"# Lições Aprendidas — {rule_id}\n\n"
        f"Lições associadas à rule `{rule_id}` da Constituição (mais recentes primeiro).\n\n"
    )


def extrair_entradas(conteudo):
    entradas = []
    for linha in conteudo.splitlines():
        linha = linha.rstrip("\r")
        if linha.startswith("- "):
            entradas.append(limpo(linha))
    return entradas


def renderizar_rule(rule_id, entradas):
    cab = cabecalho_rule(rule_id)
    corpo = "".join(f"{e}\n" for e in entradas)
    return cab + corpo


def formatar_entrada(texto=None, pattern=None, fix=None, ref=None):
    if pattern or fix:
        if not pattern or not fix:
            raise EntradaInvalida("--pattern e --fix devem ser informados juntos")
        p = validar_sem_controle(pattern, "--pattern")
        f = validar_sem_controle(fix, "--fix")
        if not p or not f:
            raise EntradaInvalida("--pattern e --fix não podem ser vazios")
        if ref:
            r = validar_sem_controle(ref, "--ref")
            return f"- {p} → {f} ({r})"
        return f"- {p} → {f}"
    if texto:
        t = validar_sem_controle(texto, "texto")
        if not t:
            raise EntradaInvalida("texto da lição não pode ser vazio")
        return t if t.startswith("- ") else f"- {t}"
    raise EntradaInvalida("informe o texto da lição ou --pattern e --fix")


def gravar_licao(raiz, rule_id, entrada):
    pasta = raiz / "sdd" / "memory" / "licoes"
    pasta.mkdir(parents=True, exist_ok=True)
    arq = pasta / f"{rule_id}.md"
    atuais = []
    if arq.is_file():
        atuais = extrair_entradas(arq.read_text(encoding="utf-8"))
    entradas = [entrada] + [e for e in atuais if e != entrada]
    conteudo = renderizar_rule(rule_id, entradas)
    while len(conteudo.encode("utf-8")) > ORCAMENTO_BYTES and len(entradas) > 1:
        entradas.pop()
        conteudo = renderizar_rule(rule_id, entradas)
    if len(conteudo.encode("utf-8")) > ORCAMENTO_BYTES:
        cab_len = len(cabecalho_rule(rule_id).encode("utf-8"))
        max_len = max(10, ORCAMENTO_BYTES - cab_len - 4)
        entradas[0] = entradas[0][: max_len - 1] + "…"
        conteudo = renderizar_rule(rule_id, entradas)
    arq.write_text(conteudo, encoding="utf-8")
    return arq


def ler_licao(raiz, rule_id):
    arq = raiz / "sdd" / "memory" / "licoes" / f"{rule_id}.md"
    if arq.is_file():
        texto = arq.read_text(encoding="utf-8").strip()
        if rule_id in texto:
            return texto
        return f"# Lições Aprendidas — {rule_id}\n\n{texto}"
    return f"# Lições Aprendidas — {rule_id}\n\n(nenhuma lição registrada para {rule_id})"


def main(argv=None):
    if argv is None:
        argv = sys.argv[1:]
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--rule", action="append", help="ID da rule na Constituição (ex: COD-02)")
    ap.add_argument("--add", action="store_true", help="adiciona uma lição no arquivo da rule")
    ap.add_argument("--pattern", help="padrão de erro identificado")
    ap.add_argument("--fix", help="correção/como evitar")
    ap.add_argument("--ref", help="referência opcional da feature ou fix")
    ap.add_argument("--dir", default=".", help="diretório raiz do projeto")
    ap.add_argument("texto", nargs="?", help="texto livre da lição quando usado com --add")
    try:
        args = ap.parse_args(argv)
    except SystemExit as e:
        return 2 if e.code else 0

    try:
        raiz = validar_caminho(args.dir, "--dir").resolve()
        regras = regras_da_constituicao(raiz)
        ids = parse_rules(args.rule, regras)
        if args.add:
            entrada = formatar_entrada(
                texto=args.texto,
                pattern=args.pattern,
                fix=args.fix,
                ref=args.ref,
            )
            for rid in ids:
                arq = gravar_licao(raiz, rid, entrada)
                print(f"✓ Lição registrada para {rid} em {limpo(arq)}")
            return 0

        blocos = [ler_licao(raiz, rid) for rid in ids]
        print("\n\n".join(blocos))
        return 0
    except EntradaInvalida as e:
        print(f"erro: {limpo(e)}", file=sys.stderr)
        return 2
    except OSError as e:
        print(f"erro: {limpo(e)}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())

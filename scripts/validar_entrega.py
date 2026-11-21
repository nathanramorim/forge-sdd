#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""Validador da entrega (106, task 1.3): confere o que é checável sem LLM e sem rede.

Uso:
  uv run scripts/validar_entrega.py --json --sem-llm [--ponto spec|entrega]
      [--feature ID] [--fixture RAIZ] [--veredito ARQUIVO.json] [--head SHA]

Confere, por spec (sdd/discovery/*/features/*.md): seções obrigatórias, IDs de rule
existentes na Constituição, ADRs citadas e vigentes. No ponto `entrega` com --feature,
a fila da feature não pode ter task em queued/building. Com --veredito, confere o
head_sha do parecer com o HEAD do git (ou --head), o ponto e o veredito `conforme`.
O que sobra para o LLM sai em `julgamento`. Nunca chama LLM nem rede; o JSON do
veredito é entrada não confiável (FLX-08): só é lido, nada dele é executado.

Códigos de saída: 0 tudo conforme; 1 há pendência; 2 uso ou entrada inválida.
"""
import argparse
import json
import re
import subprocess
import sys
from pathlib import Path

RE_SHA = re.compile(r"^[0-9a-f]{7,64}$")
RE_RULE = re.compile(r"\b[A-Z]{3}-\d{2}\b")
RE_ADR = re.compile(r"\bADR-(\d{4})\b")
RE_CTRL = re.compile(r"[\x00-\x1f\x7f-\x9f]")
SECOES = ("## Rules e IDs da Constituição Aplicáveis", "## Critérios de Aceitação Executáveis")
PONTOS = ("spec", "entrega")


class Entrada(Exception):
    """Entrada inválida: saída 2."""


def limpo(valor):
    """Remove caracteres de controle de qualquer texto que vá para a saída."""
    return RE_CTRL.sub("", str(valor))[:300]


def caminho(valor, nome):
    if ".." in Path(valor).parts:
        raise Entrada(f"{nome}: caminho com '..' recusado")
    return Path(valor)


def secao(texto, titulo):
    ini = texto.find(titulo)
    if ini < 0:
        return None
    resto = texto[ini + len(titulo):]
    fim = re.search(r"^## ", resto, re.M)
    return resto[: fim.start()] if fim else resto


def regras_da_constituicao(raiz):
    arq = raiz / "sdd" / "memory" / "constitution.md"
    if not arq.is_file():
        raise Entrada(f"constituição ausente em {limpo(arq)}")
    return set(RE_RULE.findall(arq.read_text(encoding="utf-8")))


def status_da_adr(raiz, numero):
    """Devolve None se a ADR não existe, senão o texto do Status."""
    achados = sorted((raiz / "sdd" / "spec" / "adr").glob(f"adr-{numero}-*.md"))
    if not achados:
        return None
    m = re.search(r"\*\*Status:?\*?\*?:?\s*(.+)", achados[0].read_text(encoding="utf-8"))
    return m.group(1).strip().lower() if m else ""


def checar_spec(raiz, spec, regras, falhas, julgamento):
    nome = limpo(spec.relative_to(raiz))
    texto = spec.read_text(encoding="utf-8")
    for titulo in SECOES:
        if titulo not in texto:
            falhas.append(f"{nome}: seção ausente: {titulo}")
    corpo = secao(texto, SECOES[0]) or ""
    for rid in sorted(set(RE_RULE.findall(corpo))):
        if rid not in regras:
            falhas.append(f"{nome}: rule inexistente na Constituição: {rid}")
    for num in sorted(set(RE_ADR.findall(texto))):
        st = status_da_adr(raiz, num)
        if st is None:
            julgamento.append(f"{nome}: ADR-{num} citada não existe nesta branch (pode nascer em outra feature da iniciativa)")
        elif not (st.startswith("`aceito") or st.startswith("`aprovado") or st.startswith("aceito") or st.startswith("aprovado")):
            falhas.append(f"{nome}: ADR-{num} não vigente (status: {limpo(st)})")


def checar_fila(raiz, feature, falhas):
    for q in sorted((raiz / "sdd" / "discovery").glob("*/queue.json")):
        try:
            dados = json.loads(q.read_text(encoding="utf-8"))
        except (OSError, ValueError):
            continue
        itens = dados.get("items") if isinstance(dados, dict) else None
        for it in itens if isinstance(itens, list) else []:
            if isinstance(it, dict) and it.get("task_id") == feature and it.get("status") in ("queued", "building"):
                falhas.append(f"fila não vazia para {limpo(feature)}: status {limpo(it.get('status'))}")


def escopo_da_branch(raiz):
    """Sem --feature e sem --fixture: deriva da branch (feat/106-13-x -> feat-106-13; feat/106-x -> iniciativa 106)."""
    try:
        r = subprocess.run(["git", "-C", str(raiz), "rev-parse", "--abbrev-ref", "HEAD"], capture_output=True, text=True, timeout=10, check=False)
    except (OSError, subprocess.SubprocessError):
        return None, None
    ramo = r.stdout.strip()
    m = re.match(r"^(?:feat|fix)/(\d+)-(\d+[a-z]?)-", ramo)
    if m:
        return f"feat-{m.group(1)}-{m.group(2)}", None
    m = re.match(r"^(?:feat|fix)/(\d+)-", ramo)
    return (None, m.group(1)) if m else (None, None)


def head_do_git(raiz):
    try:
        r = subprocess.run(["git", "-C", str(raiz), "rev-parse", "HEAD"], capture_output=True, text=True, timeout=10, check=False)
    except (OSError, subprocess.SubprocessError):
        return None
    saida = r.stdout.strip().lower()
    return saida if r.returncode == 0 and RE_SHA.match(saida) else None


def checar_veredito(arq, ponto, head, feature, falhas):
    try:
        dados = json.loads(arq.read_text(encoding="utf-8"))
    except (OSError, ValueError) as e:
        raise Entrada(f"veredito ilegível ou malformado: {limpo(e)}")
    if not isinstance(dados, dict):
        raise Entrada("veredito malformado: a raiz deve ser um objeto")
    sha = dados.get("head_sha")
    if not isinstance(sha, str) or not RE_SHA.match(sha):
        raise Entrada("veredito sem head_sha válido (hex de 7 a 64): recusado")
    if head is None:
        raise Entrada("HEAD indisponível: passe --head ou rode dentro de um repositório git")
    if not head.startswith(sha):
        falhas.append(f"head_sha do veredito ({sha}) não confere com o HEAD ({head})")
    if dados.get("ponto") != ponto:
        falhas.append(f"ponto do veredito ({limpo(dados.get('ponto'))}) diferente de {ponto}")
    if dados.get("veredito") != "conforme":
        falhas.append(f"veredito não é conforme: {limpo(dados.get('veredito'))}")
    if feature and dados.get("feature") not in (None, "", feature):
        falhas.append(f"feature do veredito ({limpo(dados.get('feature'))}) diferente de {limpo(feature)}")


def main(argv):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--json", action="store_true", help="saída em JSON")
    ap.add_argument("--sem-llm", action="store_true", help="aceito por contrato: o script nunca chama LLM")
    ap.add_argument("--ponto", default="spec")
    ap.add_argument("--feature")
    ap.add_argument("--fixture", help="raiz alternativa do projeto (testes)")
    ap.add_argument("--veredito")
    ap.add_argument("--head")
    try:
        a = ap.parse_args(argv)
    except SystemExit as e:
        return 2 if e.code else 0
    falhas, julgamento = [], [
        "spec contra ADRs, rules e padrão do projeto (julgamento semântico)",
        "lições da rule lidas antes de escrever (sem registro verificável local)",
        "forge gate --json não é executado aqui: rode e anexe",
    ]
    try:
        if a.ponto not in PONTOS:
            raise Entrada(f"ponto inválido {limpo(a.ponto)!r}: use spec ou entrega")
        if a.feature is not None and not re.fullmatch(r"[A-Za-z0-9._-]{1,80}", a.feature):
            raise Entrada("feature inválida")
        raiz = caminho(a.fixture or ".", "--fixture").resolve()
        arq_v = caminho(a.veredito, "--veredito") if a.veredito else None
        head = None
        if a.head is not None:
            head = a.head.strip().lower()
            if not RE_SHA.match(head):
                raise Entrada("--head inválido (hex de 7 a 64)")
        elif arq_v:
            head = head_do_git(raiz)
        regras = regras_da_constituicao(raiz)
        feature, inic = a.feature, None
        if not feature and not a.fixture:
            feature, inic = escopo_da_branch(raiz)
            if not feature and not inic:
                julgamento.append("branch sem id de feature ou iniciativa: nenhuma spec conferida (passe --feature)")
        specs = sorted((raiz / "sdd" / "discovery").glob("*/features/*.md"))
        if not a.fixture and not feature:
            specs = [s for s in specs if inic and s.parent.parent.name.startswith(inic + "-")]
        a.feature = feature
        if a.feature:
            specs = [s for s in specs if s.name.startswith(a.feature + "-") or s.name == a.feature + ".md"]
            if not specs:
                raise Entrada(f"nenhuma spec para a feature {limpo(a.feature)}")
        for s in specs:
            checar_spec(raiz, s, regras, falhas, julgamento)
        if a.ponto == "entrega" and a.feature:
            checar_fila(raiz, a.feature, falhas)
        if arq_v:
            checar_veredito(arq_v, a.ponto, head, a.feature, falhas)
    except Entrada as e:
        msg = limpo(e)
        print(json.dumps({"ok": False, "erro": msg}, ensure_ascii=False) if a.json else f"erro: {msg}")
        return 2
    except OSError as e:
        print(f"erro: {limpo(e)}")
        return 2
    ok = not falhas
    if a.json:
        print(json.dumps({"ok": ok, "ponto": a.ponto, "specs": len(specs), "pendencias": falhas, "julgamento": julgamento}, ensure_ascii=False, indent=2))
    else:
        print("ok" if ok else "pendências:\n" + "\n".join("- " + p for p in falhas))
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

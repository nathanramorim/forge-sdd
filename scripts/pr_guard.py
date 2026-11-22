#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""Barreira de PR do Claude Code (106, task 1.4): bloqueia gh pr e az repos pr fora da política.

Uso:
  uv run scripts/pr_guard.py --check '<comando>' [--veredito ARQ.json] [--head SHA] [--raiz DIR]
  uv run scripts/pr_guard.py [--hook]   # lê payload JSON do PreToolUse (Claude Code) via stdin

Regras (RN-7, SEG-01, SEG-03, FLX-08):
- `gh pr create`: permitido somente com `--draft` ou `-d`.
- `gh pr ready` (e `az repos pr update --draft false`) exige também o selo de `scripts/ci-local.sh` do HEAD.
- `gh pr ready` e `gh pr merge`: bloqueados sem veredito `conforme` válido para o HEAD atual.
- `gh pr merge`: bloqueado quando a base do PR não é a principal. Um PR com base em outra branch
  entra por squash e deixa os commits de cima órfãos — o arquivo some da principal (feat-106-24d).
- `az repos pr create`: permitido somente com `--draft true` ou `--draft=true`.
- `az repos pr update`: com `--draft false`/`--draft=false` ou `--status completed`/`--status=completed`,
  bloqueado sem veredito `conforme` válido para o HEAD atual.
- Comandos compostos (`;`, `&&`, `||`, `|`, `&`, quebras de linha, `$()`, crases e subshells)
  são decompostos e cada segmento é validado para impedir bypass.
- Outros comandos passam com código 0.

Nota de cobertura: este hook vale só no Claude Code (PreToolUse); nos outros agentes (Gemini,
Copilot) valem os passos dos comandos e o job "pronto?" da pipeline.

Códigos de saída: 0 permitido; 2 bloqueado ou entrada inválida (convenção de bloqueio de hook).
"""
import argparse
import json
import re
import shlex
import subprocess
import sys
from pathlib import Path

RE_SHA = re.compile(r"^[0-9a-f]{7,64}$")
RE_CTRL = re.compile(r"[\x00-\x1f\x7f-\x9f]")
AVISO_COBERTURA = (
    "Nota: pr_guard.py vale só no Claude Code; nos outros agentes valem os passos dos comandos "
    "e o job 'pronto?' da pipeline."
)
MSG_SELO = (
    "bloqueado (fix-106-10): '{cmd}' exige o selo do CI local para o HEAD atual. "
    "Rode scripts/ci-local.sh (grava .forge/runtime/ci-local.json) e tente de novo."
)
ENV_ASSIGN_RE = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*=.*$")
WRAPPERS_COM_ARGS = {"sudo", "env", "command", "builtin", "nohup", "time", "nice"}
SHELL_EXECS = {"sh", "bash", "zsh", "dash", "ksh"}


def limpo(valor):
    """Remove caracteres de controle de qualquer texto que vá para stderr."""
    return RE_CTRL.sub("", str(valor))[:300]


def extrair_subshells(cmd):
    """Extrai o conteúdo interno de $(...) e `...` para inspeção recursiva."""
    extras = []
    # Crases `...`
    for m in re.finditer(r"`([^`]+)`", cmd):
        extras.append(m.group(1))
    # $(...) com balanceamento de parênteses
    i = 0
    n = len(cmd)
    while i < n - 1:
        if cmd[i] == "$" and cmd[i + 1] == "(":
            prof = 1
            j = i + 2
            while j < n and prof > 0:
                if cmd[j] == "(":
                    prof += 1
                elif cmd[j] == ")":
                    prof -= 1
                j += 1
            interno = cmd[i + 2 : j - 1 if prof == 0 else j]
            if interno.strip():
                extras.append(interno)
            i = i + 2
        else:
            i += 1
    return extras


def segmentar_comando(cmd, profundidade=0):
    """Divide um comando shell em listas de tokens por segmento, incluindo subshells."""
    if profundidade > 8 or not cmd or not cmd.strip():
        return []
    segmentos = []
    for sub in extrair_subshells(cmd):
        segmentos.extend(segmentar_comando(sub, profundidade + 1))

    # Substitui parênteses de subshell fora de aspas por separador ';' usando shlex
    lex = shlex.shlex(cmd, posix=True, punctuation_chars=";&|()")
    lex.whitespace = " \t"
    lex.whitespace_split = False
    try:
        brutos = list(lex)
    except ValueError:
        # Aspas desbalanceadas: fallback conservador por regex para não deixar passar gh/az
        brutos = re.findall(r"[^\s;&|()]+|[;&|()\n\r]+", cmd)

    atual = []
    for tok in brutos:
        if set(tok) <= set(";&|()\n\r"):
            if atual:
                segmentos.append(atual)
                atual = []
        else:
            atual.append(tok)
    if atual:
        segmentos.append(atual)

    finais = []
    for seg in segmentos:
        finais.append(seg)
        # Se o segmento invoca `sh -c '...'` ou `eval ...`, analisa o argumento recursivamente
        desemp = desempacotar_wrappers(seg)
        if not desemp:
            continue
        prog = Path(desemp[0]).name.lower()
        if prog == "eval" and len(desemp) > 1:
            finais.extend(segmentar_comando(" ".join(desemp[1:]), profundidade + 1))
        elif prog in SHELL_EXECS:
            for idx, arg in enumerate(desemp[1:], start=1):
                if arg in ("-c", "-lc", "-ic") and idx + 1 < len(desemp):
                    finais.extend(segmentar_comando(desemp[idx + 1], profundidade + 1))
    return finais


def desempacotar_wrappers(tokens):
    """Remove atribuições VAR=val e prefixos como env/sudo/command do início do segmento."""
    i = 0
    n = len(tokens)
    while i < n:
        t = tokens[i]
        if ENV_ASSIGN_RE.match(t):
            i += 1
            continue
        nome = Path(t).name.lower()
        if nome in WRAPPERS_COM_ARGS:
            i += 1
            while i < n and tokens[i].startswith("-"):
                # sudo -u user / env -i / command -p
                if nome == "sudo" and tokens[i] in ("-u", "-g", "-C", "-D", "-h", "-p", "-r", "-t", "-T", "-U"):
                    i += 2
                elif nome == "env" and tokens[i] in ("-u", "-C", "-S"):
                    i += 2
                else:
                    i += 1
            continue
        break
    return tokens[i:]


def head_do_git(raiz):
    try:
        r = subprocess.run(
            ["git", "-C", str(raiz), "rev-parse", "HEAD"],
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )
    except (OSError, subprocess.SubprocessError):
        return None
    saida = r.stdout.strip().lower()
    return saida if r.returncode == 0 and RE_SHA.match(saida) else None


def veredito_valido(arq, head):
    try:
        dados = json.loads(Path(arq).read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return False
    if not isinstance(dados, dict):
        return False
    if dados.get("veredito") != "conforme":
        return False
    sha = dados.get("head_sha")
    if not isinstance(sha, str) or not RE_SHA.match(sha):
        return False
    return bool(head) and (head.startswith(sha) or sha.startswith(head))


def selo_local_valido(raiz, head_arg=None):
    """Confere o selo de scripts/ci-local.sh (.forge/runtime/ci-local.json) para o HEAD atual."""
    head = head_arg.strip().lower() if isinstance(head_arg, str) and head_arg.strip() else head_do_git(raiz)
    if not head or not RE_SHA.match(head):
        return False
    try:
        dados = json.loads((Path(raiz) / ".forge" / "runtime" / "ci-local.json").read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return False
    if not isinstance(dados, dict):
        return False
    sha = dados.get("head_sha")
    if not isinstance(sha, str) or not RE_SHA.match(sha):
        return False
    return head.startswith(sha) or sha.startswith(head)


def tem_veredito_conforme(raiz, veredito_arg=None, head_arg=None):
    """Confere se existe veredito `conforme` válido para o HEAD atual."""
    head = head_arg.strip().lower() if isinstance(head_arg, str) and head_arg.strip() else head_do_git(raiz)
    if not head or not RE_SHA.match(head):
        return False
    if veredito_arg:
        p = Path(veredito_arg)
        if ".." in p.parts:
            return False
        return veredito_valido(p, head)
    for arq in sorted((Path(raiz) / "sdd" / "discovery").glob("*/verdicts/*.json")):
        if veredito_valido(arq, head):
            return True
    return False


def gh_tem_draft(args):
    """Verifica se args de `gh pr create` contêm `--draft` ou `-d` (inclusive flag curta combinada)."""
    for a in args:
        if a == "--":
            break
        if a in ("--draft", "-d") or a.startswith("--draft="):
            val = a.split("=", 1)[1].lower() if "=" in a else "true"
            if val in ("false", "0", "no"):
                return False
            return True
        if a.startswith("-") and not a.startswith("--") and len(a) > 1:
            # Flags curtas em gh pr create que consomem valor se coladas não começam com 'd' (-B, -H, -R, -t, -b, -F, -l, -a, -m, -p, -r, -T)
            if a[1] not in ("B", "H", "R", "t", "b", "F", "l", "a", "m", "p", "r", "T") and "d" in a[1:]:
                return True
    return False


def gh_base_do_merge(args):
    """Lê o número do PR de `gh pr merge`; None quando o comando não identifica um PR.

    Sem número, o `gh` usa o PR da branch atual, e a base não pode ser conferida aqui: o
    chamador decide (o guard só bloqueia quando consegue ler o número).
    """
    ignorar_valor = {"-b", "--body", "-t", "--subject", "-m", "--body-file", "-F", "--author-email",
                     "-R", "--repo", "--match-head-commit", "--delete-branch"}
    i = 0
    while i < len(args):
        a = args[i]
        if a in ignorar_valor:
            i += 2
            continue
        if a.isdigit():
            return int(a)
        i += 1
    return None


def base_do_pr(raiz, numero):
    """Devolve a base do PR pelo `gh` (somente leitura). None quando não consegue ler."""
    try:
        r = subprocess.run(
            ["gh", "pr", "view", str(numero), "--json", "baseRefName"],
            cwd=raiz or ".", capture_output=True, text=True, timeout=15,
        )
    except (OSError, subprocess.SubprocessError):
        return None
    if r.returncode != 0:
        return None
    try:
        dados = json.loads(r.stdout or "{}")
    except ValueError:
        return None
    base = dados.get("baseRefName")
    return base if isinstance(base, str) and base else None


def principal_do_repo(raiz):
    """Principal do repositório: `main` quando não consegue descobrir."""
    try:
        r = subprocess.run(
            ["gh", "repo", "view", "--json", "defaultBranchRef"],
            cwd=raiz or ".", capture_output=True, text=True, timeout=15,
        )
    except (OSError, subprocess.SubprocessError):
        return "main"
    if r.returncode != 0:
        return "main"
    try:
        dados = json.loads(r.stdout or "{}")
    except ValueError:
        return "main"
    ref = (dados.get("defaultBranchRef") or {}).get("name")
    return ref if isinstance(ref, str) and ref else "main"


def az_valor_op(args, opcoes):
    """Extrai os valores passados para qualquer flag em `opcoes` (ex.: ('--draft',))."""
    valores = []
    i = 0
    n = len(args)
    while i < n:
        a = args[i]
        if a == "--":
            break
        for op in opcoes:
            if a == op:
                if i + 1 < n and not args[i + 1].startswith("-"):
                    valores.append(args[i + 1].lower())
                    i += 1
                else:
                    valores.append("true")
                break
            if a.startswith(op + "="):
                valores.append(a.split("=", 1)[1].lower())
                break
        i += 1
    return valores


def extrair_subcomando_gh_pr(tokens):
    """Se `tokens` for uma chamada `gh ... pr <acao> ...`, devolve (acao, resto_args), senão None."""
    if not tokens:
        return None
    prog = Path(tokens[0]).name.lower()
    if prog not in ("gh", "gh.exe"):
        return None
    # Pula flags globais antes de `pr` (ex.: `gh -R owner/repo pr create ...`)
    i = 1
    n = len(tokens)
    while i < n:
        t = tokens[i]
        if t in ("-R", "--repo", "--hostname"):
            i += 2
            continue
        if t.startswith("-"):
            i += 1
            continue
        break
    if i >= n or tokens[i] != "pr":
        return None
    i += 1
    # Procura o subcomando de `gh pr` (create, ready, merge, etc.)
    acoes = {"create", "ready", "merge", "view", "list", "status", "checks", "checkout", "diff", "edit", "close", "reopen", "review", "comment", "lock", "unlock", "update-branch"}
    resto = tokens[i:]
    for idx, t in enumerate(resto):
        if t in acoes:
            return t, resto[:idx] + resto[idx + 1 :]
    return None


def extrair_subcomando_az_pr(tokens):
    """Se `tokens` for `az repos pr <acao> ...`, devolve (acao, resto_args), senão None."""
    if not tokens:
        return None
    prog = Path(tokens[0]).name.lower()
    if prog not in ("az", "az.cmd", "az.exe"):
        return None
    if len(tokens) >= 4 and tokens[1] == "repos" and tokens[2] == "pr":
        return tokens[3], tokens[4:]
    return None


def avaliar_segmento(seg, veredito_ok, raiz=".", selo_ok=None):
    """Retorna None se permitido, ou mensagem de erro caso bloqueado."""
    tokens = desempacotar_wrappers(seg)
    if not tokens:
        return None

    gh_info = extrair_subcomando_gh_pr(tokens)
    if gh_info is not None:
        acao, args = gh_info
        if acao == "create":
            if not gh_tem_draft(args):
                return (
                    "bloqueado (RN-7): 'gh pr create' exige '--draft' (ou '-d'). "
                    f"Todo PR nasce em draft. {AVISO_COBERTURA}"
                )
            return None
        if acao in ("ready", "merge"):
            if not veredito_ok():
                return (
                    f"bloqueado (RN-7): 'gh pr {acao}' exige veredito 'conforme' válido para o SHA do HEAD atual. "
                    f"{AVISO_COBERTURA}"
                )
            if acao == "ready" and selo_ok is not None and not selo_ok():
                return MSG_SELO.format(cmd="gh pr ready") + f" {AVISO_COBERTURA}"
            if acao == "merge":
                # A base do PR precisa ser a principal: um PR com base em outra branch entra
                # por squash e deixa os commits de cima órfãos — o arquivo some da principal.
                numero = gh_base_do_merge(args)
                if numero is not None:
                    base = base_do_pr(raiz, numero)
                    principal = principal_do_repo(raiz)
                    if base is not None and base != principal:
                        return (
                            f"bloqueado (RN-22): o PR #{numero} tem base '{base}', não '{principal}'. "
                            "Um PR com base em outra branch entra por squash e deixa os commits de cima "
                            f"órfãos. Retargete para '{principal}' antes do merge (gh pr edit --base {principal}). "
                            f"{AVISO_COBERTURA}"
                        )
            return None
        return None

    az_info = extrair_subcomando_az_pr(tokens)
    if az_info is not None:
        acao, args = az_info
        if acao == "create":
            vals = az_valor_op(args, ("--draft",))
            if not vals or vals[-1] not in ("true", "1", "yes"):
                return (
                    "bloqueado (RN-7): 'az repos pr create' exige '--draft true'. "
                    f"Todo PR nasce em draft. {AVISO_COBERTURA}"
                )
            return None
        if acao == "update":
            draft_vals = az_valor_op(args, ("--draft",))
            status_vals = az_valor_op(args, ("--status",))
            sai_do_draft = any(v in ("false", "0", "no") for v in draft_vals)
            completa_pr = any(v == "completed" for v in status_vals)
            if (sai_do_draft or completa_pr) and not veredito_ok():
                return (
                    "bloqueado (RN-7): 'az repos pr update' para tirar de draft ou completar o PR "
                    f"exige veredito 'conforme' válido para o SHA do HEAD atual. {AVISO_COBERTURA}"
                )
            if sai_do_draft and selo_ok is not None and not selo_ok():
                return MSG_SELO.format(cmd="az repos pr update") + f" {AVISO_COBERTURA}"
            return None
        return None

    return None


def avaliar_comando(cmd, raiz=".", veredito_arg=None, head_arg=None):
    """Avalia uma string de comando completa. Retorna (permitido: bool, motivo: str)."""
    if not isinstance(cmd, str):
        return False, "comando inválido"
    if "\x00" in cmd:
        return False, "bloqueado (SEG-01): byte nulo no comando"

    cache_veredito = None

    cache_selo = None

    def selo_ok():
        nonlocal cache_selo
        if cache_selo is None:
            cache_selo = selo_local_valido(raiz, head_arg)
        return cache_selo

    def veredito_ok():
        nonlocal cache_veredito
        if cache_veredito is None:
            cache_veredito = tem_veredito_conforme(raiz, veredito_arg, head_arg)
        return cache_veredito

    for seg in segmentar_comando(cmd):
        motivo = avaliar_segmento(seg, veredito_ok, raiz, selo_ok)
        if motivo:
            return False, motivo
    return True, ""


def comando_do_stdin():
    """Lê payload JSON do PreToolUse (Claude Code) via stdin e devolve (ok, comando_ou_erro)."""
    bruto = sys.stdin.read()
    if not bruto or not bruto.strip():
        return True, ""
    try:
        dados = json.loads(bruto)
    except ValueError as e:
        return False, f"payload JSON inválido no hook PreToolUse: {limpo(e)}"
    if not isinstance(dados, dict):
        return False, "payload do hook PreToolUse deve ser um objeto JSON"
    tool_input = dados.get("tool_input")
    if tool_input is None:
        return True, ""
    if not isinstance(tool_input, dict):
        return False, "tool_input inválido no payload do hook"
    cmd = tool_input.get("command", "")
    if cmd is None:
        return True, ""
    if not isinstance(cmd, str):
        return False, "tool_input.command deve ser string"
    return True, cmd


def main(argv):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--check", metavar="COMANDO", help="comando shell a validar")
    ap.add_argument("--hook", action="store_true", help="lê payload JSON do PreToolUse via stdin")
    ap.add_argument("--veredito", help="caminho de veredito JSON para validar ready/merge")
    ap.add_argument("--head", help="SHA do HEAD a conferir (padrão: git rev-parse HEAD)")
    ap.add_argument("--raiz", default=".", help="raiz do repositório (padrão: .)")
    try:
        a = ap.parse_args(argv)
    except SystemExit as e:
        return 2 if e.code else 0

    raiz_p = Path(a.raiz)
    if ".." in raiz_p.parts:
        print("erro: --raiz com '..' recusado", file=sys.stderr)
        return 2
    if a.veredito and ".." in Path(a.veredito).parts:
        print("erro: --veredito com '..' recusado", file=sys.stderr)
        return 2
    if a.head is not None and not RE_SHA.match(a.head.strip().lower()):
        print("erro: --head inválido (hex de 7 a 64)", file=sys.stderr)
        return 2

    if a.check is not None:
        cmd = a.check
    else:
        ok_in, cmd_ou_erro = comando_do_stdin()
        if not ok_in:
            print(f"pr_guard: {limpo(cmd_ou_erro)}", file=sys.stderr)
            return 2
        cmd = cmd_ou_erro

    ok, motivo = avaliar_comando(cmd, raiz=raiz_p.resolve(), veredito_arg=a.veredito, head_arg=a.head)
    if not ok:
        print(f"pr_guard: {limpo(motivo)}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

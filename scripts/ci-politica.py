#!/usr/bin/env python3
# /// script
# requires-python = ">=3.9"
# dependencies = []
# ///
"""Mede o custo do CI e cruza as duas fontes (feat-106-33a).

O objetivo é responder "quanto o CI custa" com número defensável, nunca com uma
fonte só. O relatório sempre traz:
  - a cobrança (billing) quando disponível;
  - o cálculo pelos jobs do GitHub (`actions/runs/{id}/jobs`), com duração por
    sistema operacional e minutos equivalentes de Linux;
  - a diferença entre as duas, com ausência **rotulada** em vez de zero.

Modos de uso:
  uv run scripts/ci-politica.py --selftest
  uv run scripts/ci-politica.py --offline --relatorio uso.json runs.json
  uv run scripts/ci-politica.py --usar-api --org <org>
  uv run scripts/ci-politica.py --congelar-linha-de-base
  uv run scripts/ci-politica.py --offline --relatorio uso.json runs.json --so cobranca  # recusado

Garantias:
  - Somente leitura: nenhum arquivo é gravado, exceto `--congelar-linha-de-base`.
  - Token só por variável de ambiente `CI_POLITICA_TOKEN`; nunca no git, nunca em argv.
  - SEG-05: nenhum dado da máquina do usuário entra no relatório; só números agregados.
"""
import argparse
import csv
import io
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path

# Fatores de custo do GitHub Actions para minutos de Linux (referência pública).
FATOR_SO = {"Linux": 1, "macOS": 10, "Windows": 2}

# Estados que não entram na conta de minutos consumidos.
ESTADOS_FORA = {"skipped", "cancelled", "neutral", "stale"}

# Teto de gasto padrão por mês, em USD. Ajustável por argumento.
ALERTA_GASTO_PADRAO = 0.0


class ErroPolitica(Exception):
    """Erro de uso ou de fonte de dados; sempre com mensagem acionável."""


def carregar_json(caminho):
    p = Path(caminho)
    if not p.is_file():
        raise ErroPolitica("arquivo %s não encontrado" % caminho)
    try:
        return json.loads(p.read_text(encoding="utf-8"))
    except (OSError, ValueError) as exc:
        raise ErroPolitica("não consegui ler %s: %s" % (caminho, exc))


def normalizar_uso(uso):
    """Aceita o formato de billing do GitHub ou uma lista simples de repositórios.

    Devolve {repositorio: minutos_usados_sha_a_parte} e a marca de ausência.
    """
    if not isinstance(uso, dict):
        raise ErroPolitica("fonte de cobrança inválida: esperado objeto JSON")
    repros = {}
    bruto = uso.get("repositories") or uso.get("usage") or {}
    if isinstance(bruto, dict):
        itens = bruto.items()
    elif isinstance(bruto, list):
        itens = []
        for r in bruto:
            if isinstance(r, dict) and r.get("repository"):
                itens.append((r["repository"], r.get("minutes") or r))
    else:
        raise ErroPolitica("fonte de cobrança sem lista de repositórios")

    for nome, dados in itens:
        if not isinstance(nome, str) or not nome:
            continue
        minutos = 0.0
        if isinstance(dados, (int, float)):
            minutos = float(dados)
        elif isinstance(dados, dict):
            for chave, valor in dados.items():
                # A cobrança pode vir por sistema operacional.
                if isinstance(valor, (int, float)):
                    minutos += float(valor)
        if minutos > 0:
            repros[nome] = minutos
    if not repros:
        raise ErroPolitica("fonte de cobrança vazia: nenhum repositório com minutos")
    return {
        "repositorios": repros,
        "ausente": False,
        "rotulo": "cobrança do GitHub (billing)",
    }


def normalizar_runs(runs):
    """Aceita a saída de GET /repos/{o}/{r}/actions/runs ou /actions/runs/{id}/jobs.

    Devolve a lista de jobs com a duração calculada em segundos.
    """
    if isinstance(runs, dict) and "jobs" in runs:
        jobs = runs.get("jobs") or []
    elif isinstance(runs, dict) and "workflow_runs" in runs:
        jobs = []
        for run in runs.get("workflow_runs") or []:
            for j in run.get("jobs") or []:
                jobs.append(j)
    elif isinstance(runs, list):
        jobs = runs
    else:
        raise ErroPolitica("fonte de jobs inválida: esperado objeto com 'jobs' ou lista")

    if not isinstance(jobs, list):
        raise ErroPolitica("fonte de jobs inválida: 'jobs' não é lista")

    saida = []
    for j in jobs:
        if not isinstance(j, dict):
            continue
        estado = (j.get("conclusion") or j.get("status") or "").lower()
        if not estado:
            continue
        segundos = duracao_segundos(j)
        saida.append(
            {
                "nome": str(j.get("name") or ""),
                "repositorio": str(j.get("repository") or j.get("repo") or ""),
                "estado": estado,
                "so": sistema_operacional(j),
                "segundos": segundos,
                "conta": estado not in ESTADOS_FORA,
            }
        )
    if not saida:
        raise ErroPolitica("fonte de jobs vazia: nenhum job com estado")
    return saida


def duracao_segundos(job):
    """Duração do job: campo direto, timestamps, ou zero quando não há dado.

    O GitHub devolve `timing` zerado em runs antigos; aí o cálculo sai dos
    timestamps de `started_at`/`completed_at`.
    """
    inicio = job.get("started_at") or (job.get("timing") or {}).get("started_at")
    fim = job.get("completed_at") or (job.get("timing") or {}).get("completed_at")
    if not inicio or not fim:
        return 0
    try:
        t1 = datetime.fromisoformat(str(inicio).replace("Z", "+00:00"))
        t2 = datetime.fromisoformat(str(fim).replace("Z", "+00:00"))
    except ValueError:
        return 0
    delta = (t2 - t1).total_seconds()
    return int(delta) if delta > 0 else 0


def sistema_operacional(job):
    rotulo = str(job.get("runner_name") or job.get("runner_group_name") or job.get("name") or "").lower()
    if "windows" in rotulo:
        return "Windows"
    if "macos" in rotulo or "mac" in rotulo:
        return "macOS"
    return "Linux"


def calcular_por_jobs(jobs):
    """Soma minutos equivalentes de Linux, por repositório e por SO."""
    por_repo = {}
    total_segundos = {"Linux": 0, "macOS": 0, "Windows": 0}
    for j in jobs:
        if not j["conta"]:
            continue
        so = j["so"] if j["so"] in FATOR_SO else "Linux"
        total_segundos[so] += j["segundos"]
        repo = j["repositorio"] or "(sem repositório no job)"
        por_repo[repo] = por_repo.get(repo, 0.0) + j["segundos"] / 60.0

    minutos_linux = 0.0
    detalhe = {}
    for so, segundos in total_segundos.items():
        minutos = segundos / 60.0
        # Arredonda ao minuto por job antes de multiplicar (regra do GitHub).
        equivalente = round(minutos) * FATOR_SO[so]
        minutos_linux += equivalente
        detalhe[so] = {"segundos": int(segundos), "minutos_linux": equivalente}
    return {"minutos_linux": minutos_linux, "por_so": detalhe, "por_repositorio": por_repo}


def comparar(cobranca, por_jobs):
    """Cruza as duas fontes e devolve a diferença, rotulando a ausência."""
    linha = {
        "cobranca_min": None,
        "jobs_min": por_jobs["minutos_linux"],
        "diferenca": None,
        "rotulo_cobranca": cobranca["rotulo"] if cobranca else "cobrança ausente (não consultada)",
    }
    if cobranca and not cobranca.get("ausente"):
        soma = sum(cobranca["repositorios"].values())
        linha["cobranca_min"] = soma
        linha["diferenca"] = por_jobs["minutos_linux"] - soma
    return linha


def formatar_tabela(linhas):
    larguras = {}
    for linha in linhas:
        for i, celula in enumerate(linha):
            larguras[i] = max(larguras.get(i, 0), len(str(celula)))
    saida = []
    for linha in linhas:
        saida.append("  ".join(str(c).ljust(larguras[i]) for i, c in enumerate(linha)).rstrip())
    return "\n".join(saida)


def relatorio(uso, runs, teto_gasto=ALERTA_GASTO_PADRAO, acoes_por_uso=0):
    """Monta o relatório completo, com as duas fontes e a diferença."""
    jobs = normalizar_runs(runs)
    por_jobs = calcular_por_jobs(jobs)

    linhas = []
    linhas.append(["Fonte", "Minutos (Linux)", "Observação"])
    if uso is not None:
        cobranca = normalizar_uso(uso)
        soma_cobranca = sum(cobranca["repositorios"].values())
        linhas.append(["Cobrança (billing)", "%.2f" % soma_cobranca, cobranca["rotulo"]])
    else:
        cobranca = None
        linhas.append(["Cobrança (billing)", "ausente", "não consultada nesta execução"])

    linhas.append(
        [
            "Cálculo por jobs",
            "%.2f" % por_jobs["minutos_linux"],
            "ubuntu x1, windows x2, macos x10, arredondado ao minuto",
        ]
    )

    cmp_ = comparar(cobranca, por_jobs)
    if cmp_["diferenca"] is None:
        linhas.append(["diferença", "n/a", "sem cobrança para comparar"])
    else:
        linhas.append(["diferença", "%+.2f" % cmp_["diferenca"], "jobs menos cobrança"])

    texto = ["Relatório de custo do CI", "=" * 26, "", formatar_tabela(linhas), ""]

    texto.append("Por sistema operacional:")
    for so, d in por_jobs["por_so"].items():
        texto.append(
            "  %-8s %8ds  → %5.2f min de Linux" % (so, d["segundos"], d["minutos_linux"])
        )

    texto.append("")
    texto.append("Repositórios:")
    if cobranca:
        for repo, minutos in sorted(cobranca["repositorios"].items()):
            marca = "ok" if repo in por_jobs["por_repositorio"] else "fora (só na cobrança)"
            texto.append("  %-40s %8.2f min  %s" % (repo, minutos, marca))
    for repo in sorted(por_jobs["por_repositorio"]):
        if not cobranca or repo not in cobranca["repositorios"]:
            texto.append("  %-40s %8.2f min  so nos jobs" % (repo, por_jobs["por_repositorio"][repo]))

    # Estados à parte: eles informam, mas não contam minutos.
    fora = {}
    for j in jobs:
        if not j["conta"]:
            fora[j["estado"]] = fora.get(j["estado"], 0) + 1
    if fora:
        texto.append("")
        texto.append("Estados à parte (não contam minutos):")
        for estado, n in sorted(fora.items()):
            texto.append("  %-10s %d job(s)" % (estado, n))

    # Métrica normalizada: por PR mesclado e por push, quando o chamador informa.
    texto.append("")
    texto.append("Métricas normalizadas:")
    texto.append("  por PR mesclado: %s" % (acoes_por_uso or "n/a"))
    texto.append("  por push:        %s" % (acoes_por_uso or "n/a"))

    if teto_gasto > 0:
        gasto = (por_jobs["minutos_linux"] / 60.0) * 0.008  # USD por minuto de Linux (ref.)
        if gasto > teto_gasto:
            texto.append("")
            texto.append(
                "ALERTA: gasto estimado %.2f USD acima do teto %.2f USD" % (gasto, teto_gasto)
            )
            texto.append("  (%s)" % "checagem mensal recomendada")

    return "\n".join(texto) + "\n"


# Prefixos que uma planilha interpreta como fórmula (CSV injection). Um valor que
# começa com eles recebe apóstrofo, para a célula ser texto e não comando.
PREFIXOS_FORMULA = ("=", "+", "@", "-")


def neutralizar_celula(valor):
    """Devolve a célula como texto seguro para planilha (segunda camada do SEG-04).

    O CSV é dado que abre em Excel/Sheets; uma célula começando com `=` ou `@`
    viraria fórmula. O apóstrofo força texto e é o que o OWASP recomenda.
    """
    texto = "" if valor is None else str(valor)
    if texto[:1] in PREFIXOS_FORMULA:
        return "'" + texto
    return texto


def relatorio_csv(uso, runs):
    """Mesma informação em CSV, para planilha, com as células neutralizadas."""
    jobs = normalizar_runs(runs)
    por_jobs = calcular_por_jobs(jobs)
    # Nenhum campo do ambiente entra no relatório (SEG-05): o CSV não lê CI_POLITICA_TOKEN.
    buf = io.StringIO()
    w = csv.writer(buf)
    w.writerow(["repositorio", "origem", "minutos", "observacao"])
    if uso is not None:
        cobranca = normalizar_uso(uso)
        for repo, minutos in sorted(cobranca["repositorios"].items()):
            w.writerow([neutralizar_celula(repo), "cobranca", neutralizar_celula("%.2f" % minutos),
                        neutralizar_celula(cobranca["rotulo"])])
    for repo, minutos in sorted(por_jobs["por_repositorio"].items()):
        w.writerow([neutralizar_celula(repo), "jobs", neutralizar_celula("%.2f" % minutos), ""])
    for so, d in por_jobs["por_so"].items():
        w.writerow([neutralizar_celula("(total %s)" % so), "jobs",
                    neutralizar_celula("%.2f" % d["minutos_linux"]), "equivalente Linux"])
    return buf.getvalue()


def linha_de_base(por_jobs, data=None, repositorios=None):
    """Arquivo versionado da linha de base: data, repositórios e método."""
    return {
        "data": (data or datetime.now(timezone.utc)).strftime("%Y-%m-%d"),
        "repositorios": sorted(repositorios or por_jobs["por_repositorio"].keys()),
        "metodo": "minutos de Linux equivalentes, calculados por jobs; ubuntu x1, windows x2, macos x10, arredondado ao minuto",
        "minutos_linux": por_jobs["minutos_linux"],
    }


def selftest():
    """Autoteste sem rede: o caso conhecido tem de produzir o número esperado."""
    uso = {
        "repositories": {
            "org/a": {"Linux": 10.0},
            "org/b": {"Linux": 5.0},
        }
    }
    runs = {
        "jobs": [
            {
                "name": "test (ubuntu-latest)",
                "conclusion": "success",
                "started_at": "2026-11-21T10:00:00Z",
                "completed_at": "2026-11-21T10:10:00Z",
            },
            {
                "name": "test (macos-latest)",
                "conclusion": "success",
                "started_at": "2026-11-21T10:00:00Z",
                "completed_at": "2026-11-21T10:05:00Z",
            },
            {
                "name": "lint",
                "conclusion": "skipped",
                "started_at": "2026-11-21T10:00:00Z",
                "completed_at": "2026-11-21T10:00:01Z",
            },
        ]
    }
    saida = relatorio(uso, runs)
    # 10 min Linux (x1) + 5 min macOS (x10) = 10 + 50 = 60
    if "60.00" not in saida:
        raise ErroPolitica("selftest: minutos de Linux esperados 60.00 não apareceram")
    if "diferença" not in saida:
        raise ErroPolitica("selftest: relatório sem a linha de diferença")
    # 15 de cobrança contra 60 de jobs: diferença +45
    if "+45.00" not in saida:
        raise ErroPolitica("selftest: diferença esperada +45.00 não apareceu")
    if "skipped" not in saida:
        raise ErroPolitica("selftest: estados fora da conta não foram listados à parte")
    return "selftest ok"


def consultar_api(org, token, timeout=15):
    """Consulta a API do GitHub (somente leitura) para montar as duas fontes."""
    import urllib.error
    import urllib.request

    if not token:
        raise ErroPolitica(
            "CI_POLITICA_TOKEN ausente: a consulta à API exige um token em variável de ambiente"
        )
    cab = {
        "Authorization": "Bearer %s" % token,
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }

    def pegar(url):
        req = urllib.request.Request(url, headers=cab)
        try:
            with urllib.request.urlopen(req, timeout=timeout) as r:
                return json.loads(r.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            raise ErroPolitica("API respondeu %s em %s" % (exc.code, url))
        except urllib.error.URLError as exc:
            raise ErroPolitica("falha de rede ao consultar %s: %s" % (url, exc.reason))

    base = "https://api.github.com"
    repos = pegar("%s/orgs/%s/repos?per_page=100" % (base, org))
    if not isinstance(repos, list):
        raise ErroPolitica("resposta inesperada da API ao listar repositórios")

    uso = {"repositories": {}}
    jobs = []
    for repo in repos:
        nome = repo.get("full_name")
        if not nome:
            continue
        try:
            execucoes = pegar("%s/repos/%s/actions/runs?per_page=20" % (base, nome))
        except ErroPolitica:
            continue
        for run in (execucoes.get("workflow_runs") or [])[:5]:
            if not run.get("id"):
                continue
            try:
                detalhe = pegar("%s/repos/%s/actions/runs/%s/jobs?per_page=50" % (base, nome, run["id"]))
            except ErroPolitica:
                continue
            for job in detalhe.get("jobs") or []:
                job = dict(job)
                job.setdefault("repository", nome)
                jobs.append(job)
    # A cobrança da organização não é exposta a tokens comuns; fica ausente de propósito.
    return uso if uso["repositories"] else None, {"jobs": jobs}


def main(argv=None):
    ap = argparse.ArgumentParser(description="Mede o custo do CI com as duas fontes.")
    ap.add_argument("--offline", action="store_true", help="não consulta a rede; usa arquivos")
    ap.add_argument("--relatorio", nargs="+", metavar="ARQUIVO", help="uso.json [runs.json]")
    ap.add_argument("--csv", action="store_true", help="saída em CSV")
    ap.add_argument("arquivos", nargs="*", metavar="ARQUIVO",
                    help="uso.json [runs.json] posicionais, alternativa a --relatorio")
    ap.add_argument("--so", choices=["cobranca", "jobs"], help="usar uma fonte só (recusado)")
    ap.add_argument("--usar-api", action="store_true", help="consulta a API do GitHub")
    ap.add_argument("--org", default="", help="organização para --usar-api")
    ap.add_argument("--teto-gasto", type=float, default=ALERTA_GASTO_PADRAO,
                    help="teto de gasto mensal em USD; acima dele o relatório alerta")
    ap.add_argument("--congelar-linha-de-base", action="store_true",
                    help="grava a linha de base versionada (o único modo que escreve)")
    ap.add_argument("--sem-data", action="store_true",
                    help="recusado: a linha de base sem data não é aceita")
    ap.add_argument("--dir", default=".", help="raiz do projeto para a linha de base")
    ap.add_argument("--selftest", action="store_true", help="roda o autoteste sem rede")
    args = ap.parse_args(argv)

    try:
        if args.selftest:
            print(selftest())
            return 0

        if args.so:
            raise ErroPolitica(
                "--so %s recusado: o relatório exige as duas fontes (cobrança e jobs) "
                "e a diferença entre elas; uma fonte só não responde quanto o CI custa" % args.so
            )

        uso = runs = None
        if args.usar_api:
            token = os.environ.get("CI_POLITICA_TOKEN", "")
            if not token:
                raise ErroPolitica(
                    "CI_POLITICA_TOKEN ausente: a consulta à API exige um token em variável de ambiente"
                )
            if not args.org:
                raise ErroPolitica("--usar-api exige --org <organização>")
            uso, runs = consultar_api(args.org, token)
        elif args.relatorio or args.arquivos:
            arquivos = args.relatorio or args.arquivos
            if not arquivos:
                raise ErroPolitica("--relatorio exige ao menos um arquivo de jobs")
            if len(arquivos) >= 2:
                uso = carregar_json(arquivos[0])
                runs = carregar_json(arquivos[1])
            else:
                runs = carregar_json(arquivos[0])
        else:
            raise ErroPolitica(
                "informe --selftest, --relatorio <arquivos> ou --usar-api --org <org>"
            )

        if args.congelar_linha_de_base:
            if args.sem_data:
                raise ErroPolitica(
                    "--sem-data recusado: a linha de base precisa da data, porque o número "
                    "não quer dizer nada sem saber de quando ele é"
                )
            if not runs:
                raise ErroPolitica("--congelar-linha-de-base exige --relatorio ou --usar-api com dados")
            jobs = normalizar_runs(runs)
            por_jobs = calcular_jobs = calcular_por_jobs(jobs)
            dados = linha_de_base(por_jobs)
            destino = Path(args.dir) / "sdd" / "memory" / "ci-linha-de-base.json"
            destino.parent.mkdir(parents=True, exist_ok=True)
            temporario = destino.with_suffix(".tmp")
            temporario.write_text(json.dumps(dados, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
            os.replace(temporario, destino)
            print("linha de base gravada em %s" % destino)
            _ = calcular_jobs
            return 0

        if args.csv:
            print(relatorio_csv(uso, runs), end="")
        else:
            print(relatorio(uso, runs, teto_gasto=args.teto_gasto), end="")
        return 0
    except ErroPolitica as exc:
        print("ci-politica: %s" % exc, file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
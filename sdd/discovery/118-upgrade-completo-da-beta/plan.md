# 118 — Plano

Quebra de baixo para cima; máx. 7 tasks por onda.

## Onda 1 — tira o bloqueio (já especificadas, PR #256)

| Task | O quê | Depende de |
|---|---|---|
| fix-106-14a | entregar `scripts/pr_guard.py` e só então o hook; doctor confere o arquivo | — |
| fix-106-14b | wrapper `forge hook pr-guard` com fail-open e migração do comando (emenda da `ARQ-06` + ADR) | 14a |

## Onda 2 — não perder dado

| Task | O quê | Achados |
|---|---|---|
| feat-118-01 | persistir stack e banco no `.sddrc` e repopular nos adaptadores | 2 |
| feat-118-02 | arquivos de instrução (`CLAUDE.md` etc.) com marcador de gerado: editado pelo usuário fica intocado | 3, 10 |

## Onda 3 — o que a 5.2.0 promete

| Task | O quê | Achados |
|---|---|---|
| feat-118-03 | campos de modo na Constituição: template do `init` e acréscimo seguro no `update` (só se faltar) | 5 |
| feat-118-04 | rules e `schema.json`: atualização por marcador/versão (preserva editadas, avisa) | 6 |
| feat-118-05 | workflow próprio `forge-pronto.yml` (arquivo novo, opt-in `--ci`) e texto do `/upgrade-sdd` honesto | 4 |

## Onda 4 — verdade nos diagnósticos

| Task | O quê | Achados |
|---|---|---|
| feat-118-06 | `doctor` e `--dry-run` completos (lista o plano; hook confirmado; sem aviso circular; JSON inválido retorna erro e não grava versão) | 7, 8, 9 |
| feat-118-07 | fixture 5.0 e teste ponta a ponta `TestUpgradeDe50` para os 4 agentes | todos |

Paralelismo: 118-01 e 118-02 juntas; 118-03, 04 e 05 juntas depois da onda 2; 118-06 e 118-07 por último. `/split-features` fatia cada uma em spec própria.

## Rules e IDs por task

`SEG-02` (01, 02, 03, 04), `ARQ-05` (06), `TST-02` e `TST-03` (todas), `GOV-01` (05: se mexer no workflow do repositório, não; a Constituição só muda em 03 com ADR).

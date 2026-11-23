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

**Fronteira com o fix-106-09 (DUP-03):** o fix-106-09 é o **contrato** de atualização (manifesto, rule não editada atualiza, N-1 para N); a 118 é a **implementação**. As features 118-02, 118-04 e 118-07 implementam os furos 1, 3 e 6 do 106-09 e **não** haverá implementação separada do 106-09: ele fica como referência e é marcado "implementado pela 118" no fim.

**Ordem e bloqueios:** a 118-01 vem **antes** da 118-02 (ambas reescrevem os adaptadores de instrução; a 02 usa o `.sddrc` com stack e banco que a 01 grava). A onda 1 depende do aceite do dono para a ADR-0027 (14b); as ondas 2 e 3 **não** dependem dela e podem seguir. A 118-06 mistura três achados e será **fatiada no `/split-features`** em 06a (doctor e confirmação do hook), 06b (`--dry-run` completo) e 06c (JSON inválido e `--yes`).

Ordem da onda 4 (split): 06a depois de 14a e 118-05; 06b depois de 06a e de 02 a 05; 06c depois de 06b (mesmos arquivos); 07 depois de tudo.

Paralelismo: 118-02 depois da 118-01; 118-03, 04 e 05 juntas depois da onda 2; 118-06 e 118-07 por último. `/split-features` fatia cada uma em spec própria.

## Rules e IDs por task

| Task | IDs |
|---|---|
| 14a | `SEG-02`, `TST-02`, `TST-03`, `DUP-02` |
| 14b | `ARQ-06`, `GOV-01`, `P-24`, `TST-03` |
| 118-01 | `SEG-02`, `DUP-03`, `TST-03` |
| 118-02 | `SEG-02`, `DUP-06`, `TST-02`, `TST-03` |
| 118-03 | `SEG-02`, `FLX-05` (ADR curta), `TST-02`, `TST-03` |
| 118-04 | `SEG-02`, `DUP-06`, `TST-03` |
| 118-05 | `SEG-02`, `ARQ-06` (flag, não subcomando), `TST-02` |
| 118-06 | `ARQ-05`, `TST-03` |
| 118-07 | `TST-03`, `TST-02` |

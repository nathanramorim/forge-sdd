# fix-106-08 — Spec e ADR só nascem no lugar certo, e o histórico sai das próprias specs

**Branch:** `docs/106-specs-lugar-certo-e-compat` (PR único das specs 08 e 09)
**Origem:** Iniciativa 106, relato do dono sobre outro projeto que usa o forge-sdd (22/11/2026); `internal/layout` (ADR-0002)
**Tamanho:** `M`
**Status:** `spec` (Gate 1 aprovado em 22/11/2026, `[Seguir]`)

## Sintoma

O dono relatou que, **em outro projeto que usa o forge-sdd**, agentes criaram **arquivos temporários de spec e de ADR em lugar diferente** do previsto. Regra pedida: nenhum arquivo de spec, ADR ou qualquer outro que comprometa a estrutura do forge-sdd e a continuidade de contexto pode ser criado fora do lugar. E mais: o projeto precisa de **histórico de decisão** e de **controle do que foi feito, novo ou fix**.

**Limite honesto da apuração:** o projeto afetado **não foi examinado** (o dono não indicou o caminho). Esta spec parte do que o código deste repositório mostra; a causa no outro projeto fica como **primeira tarefa da implementação** (levantar, só leitura, antes de codar).

## O que o código deste repositório mostra

| Pergunta | Evidência | Conclusão |
|---|---|---|
| Alguma instrução manda gravar spec ou ADR em `/tmp`? | `grep` de `/tmp/`, `mktemp` e `rascunho` em `.agents`, `.claude`, `.gemini`, `.github` e `sdd/memory`: nenhum caso de instrução (só a palavra "rascunho" do estado do `brief`) | **Não** |
| O código cria temporário? | só `os.CreateTemp` de escrita atômica em `internal/queue/atomic.go`, na pasta do alvo | **Técnico, não é spec** |
| O agente cria por conta própria? | o revisor gravou `/tmp/forge-verdict-*.json` (visto em `.forge/runtime/logs`); `internal/pipeline/review.go:308` admite que "o registro se perde junto com o /tmp" | **Sim: o agente improvisa quando nada diz o destino** |
| Algo detecta spec ou ADR fora de lugar? | `forge gate` tem P-01..P-23; nenhum olha o **local** do artefato | **Não**: ninguém avisa nem barra |
| Há histórico de decisão? | 24 arquivos em `sdd/spec/adr/`, **sem índice**; `decisions.md` (tabela D1..); `progress-log.md` (53 linhas, texto livre do Archivist); `forge report` (tokens por discovery, feature e fix) | **Espalhado**, nenhuma visão única |
| Há controle de novo versus fix? | spec de feature em `features/`, de fix em `fixes/`, com cabeçalho `Origem`, `Status`, `Tamanho`; `forge report` separa por tipo **só para métricas** | **Existe na estrutura; falta a visão derivada** |
| ADRs podem colidir? | 159 branches remotas; o `forge next-id` evita colisão de **discovery e fix**, não de **ADR** | **Risco**: dois ADRs com o mesmo número em branches diferentes |

## Feature Correlacionada

- **Arquivo de origem:** `internal/layout` (donos dos caminhos), `internal/gate` (checks P-xx) e `sdd/spec/adr/`.
- **Causa-raiz:** o destino de spec e ADR é uma **convenção**, não uma **garantia**. Sem checagem determinística e sem instrução explícita do destino, o agente improvisa um temporário, e o contexto some quando a sessão acaba. O histórico não é derivado de uma fonte única, então cada vez que se pergunta "o que mudou" é preciso juntar PRs, `status.md`, `progress-log.md` e ADRs à mão.

## Modelo do módulo

**Completo** (`ARQ-11`): um check novo em `internal/gate`, a leitura dos cabeçalhos das specs em `internal/layout`/`internal/gate` e uma opção nova em `forge report`. Sem subcomando novo.

## Decisões do projeto aplicadas

- ADR-0002 (layout e estado por iniciativa): `internal/layout` é o **único** dono de caminho de spec, ADR e estado.
- `DUP-01` — **uma fonte** do histórico: as próprias specs e o git. **Nenhum arquivo novo de "ledger" escrito à mão** (seria mais um lugar para divergir).
- `ARQ-01` / `ARQ-06` — a visão sai de `forge report` (subcomando que já existe), com flag; o check entra no `forge gate` que já existe. Nada de subcomando novo.
- `SEG-01` — caminho vindo do disco é dado externo: normalizado, sem `..`, sem seguir symlink para fora do projeto.
- `COD-04` — nada de temporário sobrevivendo: o que o agente cria como rascunho vive **na branch**, no destino final.
- Regra do dono: pergunta e decisão por componente de ask (`fix-106-07`).

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Dono dos caminhos | `internal/layout` | funções por artefato (`InitiativeQueue`, `Worktrees`, `Constitution`...) | **reusar**; acrescentar `SpecDirs`/`ADRDir` se faltar |
| Checks determinísticos | `internal/gate` (P-01..P-23) | estrutura de check por ID, `forge gate` | **estender** com o próximo ID livre |
| Relatório por tipo | `forge report` | agrega `sdd/.metrics/session-*.json` por discovery, feature e fix | **estender** com `--historico` |
| Numeração sem colisão | `forge next-id` | cobre discovery e fix, olhando todas as branches | **estender** para ADR |
| Cabeçalho das specs | `fixes/*.md` e `features/*.md` | `Origem`, `Tamanho`, `Status`, `Branch` | **reusar** como dado; acrescentar `Tipo` (`feature` ou `fix`) onde faltar |
| Histórico à mão | `progress-log.md` (Archivist) | texto livre | **manter**; não é a fonte da visão |
| Índice de ADR | — | não existe | **gerar** (derivado), não escrever à mão |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-06`, `ARQ-11`, `COD-03`, `COD-04`, `DUP-01`, `DUP-02`, `FLX-01`, `FLX-05`, `GOV-01`, `SEG-01`, `TST-01`, `TST-03`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Escopo

**Dentro:**
1. **Levantar primeiro (só leitura)** o caso do outro projeto: quais arquivos, onde, criados por qual agente e por quê. Sem isso, o check abaixo pode ficar frouxo ou apertado demais.
2. **Check `forge gate` novo** (próximo P-xx livre): falha se existir arquivo com a **assinatura** de spec (`## Critérios de Aceitação Executáveis`) ou de ADR (`# ADR-NNNN`) **fora** de `sdd/discovery/<id>/**` e `sdd/spec/adr/`. Ignora `node_modules`, `.git`, `.forge/runtime/worktrees` e `testdata`.
3. **Aviso no `forge doctor`:** conta e lista os arquivos fora do lugar, com o caminho e o que fazer (mover para o destino e apagar o temporário).
4. **Instrução única** (no `clarify.md`, seção "Onde nasce cada artefato"): spec e ADR são criados **direto no destino**, na branch; rascunho é o próprio arquivo da branch; **nunca** em `/tmp`, na raiz ou em pasta própria. Verdicts e resultados de revisão também vão para a pasta da iniciativa (`review.go` já preserva uma cópia).
5. **`forge report --historico`:** tabela **derivada** (lê as specs e o git), uma linha por item: iniciativa, **tipo (feature ou fix)**, id, status, ADRs e decisões citados, PR e data. Saída em texto e em `--json`. Escrita nenhuma em disco.
6. **Numeração de ADR sem colisão:** `forge next-id --adr` considera todas as branches locais e remotas, como já faz para discovery e fix.
7. **Índice de ADR gerado** pelo mesmo `--historico` (número, título, status, ADRs que substitui).

**Fora:**
- Hook de bloqueio de escrita em tempo de execução (concorre com o agente; mesma lição do `/goal`). A garantia é o check no `gate` e o aviso no `doctor`.
- Reescrever specs antigas para ganhar o campo `Tipo` (o relatório infere pela pasta `features/` ou `fixes/` quando o campo falta).
- Migrar o `progress-log.md`.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. O check acusa spec e ADR fora do lugar e aceita os do lugar certo.
go test ./internal/gate/... -run 'TestSpecForaDoLugarReprova|TestADRForaDoLugarReprova|TestSpecNoLugarCertoPassa'
# 2. Falso positivo controlado: worktrees, node_modules e testdata não contam.
go test ./internal/gate/... -run 'TestForaDoLugarIgnoraWorktreesETestdata'
# 3. O doctor lista o arquivo e o destino correto.
go test ./cmd/forge-sdd/... -run 'TestDoctorAvisaArtefatoForaDoLugar'
# 4. O histórico é derivado, distingue feature de fix e não escreve nada.
go test ./cmd/forge-sdd/... -run 'TestReportHistoricoSeparaFeatureEFix|TestReportHistoricoNaoGrava|TestReportHistoricoJSON'
# 5. ADR sem colisão entre branches.
go test ./cmd/forge-sdd/... -run 'TestNextIDADRConsideraOutrasBranches'
# 6. A instrução existe em um só lugar.
grep -q 'Onde nasce cada artefato' sdd/memory/clarify.md
# 7. Nenhum subcomando novo (ARQ-06 intacto).
! grep -rn 'Use: *"historico' cmd/forge-sdd || exit 1
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** um `/tmp/spec.md` simulado dentro de um projeto de teste (arquivo com a assinatura de spec fora de `sdd/`) **reprova** o `forge gate`; movido para `sdd/discovery/<id>/fixes/`, **passa**. Um caminho com `..` ou symlink para fora do projeto é recusado, sem ler.
- **Prova do histórico:** num projeto de teste com 2 features e 2 fixes, `--historico` devolve 4 linhas com o tipo certo, sem tocar disco.

## Riscos

- **R1 — Falso positivo** (documentação que cita o cabeçalho de spec como exemplo). Mitigação: a assinatura exige o cabeçalho no **início de linha** e o arquivo fora das pastas ignoradas; exceção por marcador `<!-- forge-sdd:exemplo -->`.
- **R2 — O check não pega o caso real** do outro projeto. Mitigação: o levantamento (item 1) vem antes e ajusta a assinatura.
- **R3 — Histórico derivado depende do git**. Mitigação: sem `gh`/git disponível, a coluna de PR fica vazia e o relatório continua, sem inventar dado.

# feat-118-03 — Campos de modo na Constituição do usuário: template do `init` e acréscimo seguro no `update`

**Depende de:** feat-118-01 e feat-118-02 (onda 2 concluída)
**Achado:** 5. **Tamanho:** `M`. **Status:** `spec`

## Objetivo

Projeto novo nasce com a seção `## Ferramentas e Integrações` e os cinco campos de modo; projeto 5.0 recebe os campos **que faltam** no `update`, com padrões seguros, sem alterar nenhum valor existente. Traz a **ADR curta (`FLX-05`)** que registra o acréscimo na Constituição do **usuário** e a exceção no `shouldPreserve`. A `GOV-01` rege só a Constituição **deste** repositório e não se aplica ao arquivo do usuário.

## Entradas

- `templates/sdd/memory/constitution.md.tmpl` (hoje sem os campos: nenhuma ocorrência de atestado, WIP ou início automático).
- Modelo dos campos: tabela `## Ferramentas e Integrações` de `sdd/memory/constitution.md` (linhas 22-30) e `vcs.LeModos`/`modosPadrao` (`internal/vcs/modos.go`).

## Saídas

- Template do `init` com os campos e padrões de `modosPadrao()` (atestado `local`, revisão `com-humano`, WIP `3`, módulos `nenhum`, início automático `desligado`; o VCS segue o já escolhido).
- `update`: função de acréscimo que, **só se a linha do campo não existir**, insere a linha na tabela (cria a seção se faltar); valor existente, mesmo inválido, nunca é tocado; resultado gravado de forma atômica.
- `shouldPreserve` ganha exceção explícita e estreita: `sdd/memory/constitution.md` continua preservado, **exceto** pelo acréscimo de campos ausentes.
- ADR em `sdd/spec/adr/` (próximo número livre; conferir o índice e as ADRs 0027 em aberto antes de numerar).

## Modelo do módulo

**Completo** (`ARQ-11`): template (`scaffold`), acréscimo (`scaffold`), leitura de verificação (`vcs`), ADR. Justificativa: muda o contrato de um arquivo de domínio do usuário.

## Decisões do projeto aplicadas

- `SEG-02`: sem perda; gravação atômica; falha de parse da Constituição = não toca e avisa.
- `FLX-05`: ADR curta no mesmo PR.
- `DUP-06`: os padrões vêm de `vcs.modosPadrao()` (exportar o necessário), não de uma segunda lista.
- Decisão do dono (22/11/2026): acréscimo seguro no `update`, sem reescrita.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Leitura e padrões dos modos | `internal/vcs/modos.go` (`modosPadrao`, `LeModos`, `aplicar`) | já lê os cinco campos da tabela | **reusar** | fonte única dos padrões |
| Preservação de `sdd/` | `shouldPreserve` `scaffold.go:312-319` | tudo sob `sdd/` é preservado | **estender** com exceção estreita | achado 5 |
| Template da Constituição | `templates/sdd/memory/constitution.md.tmpl` | sem os campos | **estender** | projeto novo |
| Acréscimo seguro de JSON | `hooks.go` `mergeClaudeHook*` | padrão de mesclagem sem duplicar | **reusar a ideia** | mesma garantia, em Markdown |
| Constituição deste repo | `sdd/memory/constitution.md:22-30` | modelo da seção | **reusar** como texto | consistência |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-02`, `FLX-05` (ADR curta), `DUP-06`, `DUP-03`, `DUP-02`, `TST-02` (template e golden no mesmo PR), `TST-03`, `ARQ-11`. Fora de alcance: `GOV-01` (só a Constituição deste repo).

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestInitConstituicaoTemCamposDeModo|TestUpdateAcrescentaSoOsCamposQueFaltam|TestUpdateNaoAlteraValorExistenteDeModo|TestUpdateConstituicaoIlegivelNaoETocada|TestShouldPreserveConstituicaoSoAceitaAcrescimo)$' -count=1
go test ./internal/vcs -run '^TestLeModosAposAcrescimoDoUpdate$' -count=1
[ "$(go test ./internal/scaffold ./internal/vcs -list '^(TestInitConstituicaoTemCamposDeModo|TestUpdateAcrescentaSoOsCamposQueFaltam|TestUpdateNaoAlteraValorExistenteDeModo|TestUpdateConstituicaoIlegivelNaoETocada|TestShouldPreserveConstituicaoSoAceitaAcrescimo|TestLeModosAposAcrescimoDoUpdate)$' | grep -c '^Test')" -eq 6 ]
ls sdd/spec/adr/ | grep -c 'constituicao-do-usuario' | grep -qx 1
go test ./internal/scaffold ./internal/vcs -count=1
```

## Fora do escopo

- Mudar valor de campo já escolhido pelo dono.
- Reescrever a Constituição inteira ou as regras `ID` dela.

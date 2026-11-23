# feat-118-03 — Campos de modo na Constituição do usuário: template do `init` e acréscimo seguro no `update`

**Depende de:** feat-118-01 e feat-118-02 (onda 2 concluída)
**Achado:** 5. **Tamanho:** `M`. **Status:** `spec`

## Objetivo

Projeto novo nasce com a seção `## Ferramentas e Integrações` e os cinco campos de modo; projeto 5.0 recebe os campos **que faltam** no `update`, com padrões seguros, sem alterar nenhum valor existente. Traz a **ADR-0028 curta (`FLX-05`)** que registra o acréscimo na Constituição do **usuário** como passo à parte (o `shouldPreserve` segue preservando o arquivo). A `GOV-01` rege só a Constituição **deste** repositório e não se aplica ao arquivo do usuário.

## Entradas

- `templates/sdd/memory/constitution.md.tmpl` (hoje sem os campos: nenhuma ocorrência de atestado, WIP ou início automático).
- Modelo dos campos: tabela `## Ferramentas e Integrações` de `sdd/memory/constitution.md` (linhas 22-30) e `vcs.LeModos`/`modosPadrao` (`internal/vcs/modos.go`).

## Saídas

- Template do `init` com os campos e padrões de `modosPadrao()` (atestado `local`, revisão `com-humano`, WIP `3`, módulos `nenhum`, início automático `desligado`; o VCS segue o já escolhido).
- `shouldPreserve(dest, targetDir) bool` **não muda**: continua `true` para `sdd/memory/constitution.md`, então `Run`/`renderDir` nunca tocam o arquivo.
- O acréscimo é um **passo à parte** (ex.: `AcrescentaCamposDeModo(targetDir)`), chamado pelo `update` **depois** de `Run`: só se a linha do campo não existir, insere a linha na tabela; valor existente, mesmo inválido, nunca é tocado; gravação atômica (`writeFileSafe`).
- **Idempotente:** segunda execução, com todas as linhas presentes, não grava (arquivo e mtime intactos).
- Seção `## Ferramentas e Integrações` ausente: é criada ao fim. Seção presente **sem tabela** (ou tabela ilegível): não grava e avisa em stderr.
- ADR em `sdd/spec/adr/` com o número **0028** (a 0027 é da 14b) e nome contendo `constituicao-do-usuario`; registra o acréscimo como passo à parte, fora do `shouldPreserve`.

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
| Preservação de `sdd/` | `shouldPreserve` `scaffold.go:312-319` (assinatura `bool`) | tudo sob `sdd/` é preservado; não comporta "exceto acréscimo" | **reusar sem mudar**; acréscimo em passo à parte | achado 5 |
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
go test ./internal/scaffold -run '^(TestInitConstituicaoTemCamposDeModo|TestUpdateAcrescentaSoOsCamposQueFaltam|TestUpdateNaoAlteraValorExistenteDeModo|TestUpdateConstituicaoIlegivelNaoETocada|TestAcrescimoDeModoEIdempotente|TestAcrescimoDeModoSecaoSemTabelaNaoGravaEAvisa|TestShouldPreserveConstituicaoSoAceitaAcrescimo)$' -count=1
go test ./internal/vcs -run '^TestLeModosAposAcrescimoDoUpdate$' -count=1
[ "$(go test ./internal/scaffold ./internal/vcs -list '^(TestInitConstituicaoTemCamposDeModo|TestUpdateAcrescentaSoOsCamposQueFaltam|TestUpdateNaoAlteraValorExistenteDeModo|TestUpdateConstituicaoIlegivelNaoETocada|TestAcrescimoDeModoEIdempotente|TestAcrescimoDeModoSecaoSemTabelaNaoGravaEAvisa|TestShouldPreserveConstituicaoSoAceitaAcrescimo|TestLeModosAposAcrescimoDoUpdate)$' | grep -c '^Test')" -eq 8 ]
ls sdd/spec/adr/ | grep -c 'constituicao-do-usuario' | grep -qx 1
ls sdd/spec/adr/ | grep -q '^adr-0028-'
go test ./internal/scaffold ./internal/vcs -count=1
```

`TestShouldPreserveConstituicaoSoAceitaAcrescimo` afirma que `Run`/`renderDir` não tocam `sdd/memory/constitution.md` (só o passo de acréscimo o grava).

## Fora do escopo

- Mudar valor de campo já escolhido pelo dono.
- Reescrever a Constituição inteira ou as regras `ID` dela.

## Nota da Conferência 1

O número da ADR é o próximo livre na `main` quando o Builder começar (a 0027 é da 14b; se a 14b ainda não mesclou, a 0028 deixa a 0027 reservada, o que é inofensivo). Confira com `ls sdd/spec/adr`.

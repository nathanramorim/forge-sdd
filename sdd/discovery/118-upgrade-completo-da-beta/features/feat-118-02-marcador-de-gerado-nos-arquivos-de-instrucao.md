# feat-118-02 — Arquivos de instrução com marcador de gerado: edição do usuário fica intocada

**Depende de:** feat-118-01 (usa `stack` e `db` do `.sddrc`; ambas reescrevem os adaptadores)
**Achados:** 3 e 10 (parte dos adaptadores). **Tamanho:** `M`. **Status:** `spec`

## Objetivo

`CLAUDE.md`, `GEMINI.md`, `OPENAI.md`, `.gemini/system_instructions.md`, `.openai/system_instructions.md` e `.github/copilot-instructions.md` ganham o marcador `<!-- forge-sdd:generated -->` (o `AGENTS.md` já o tem). Com o marcador, o `update` regrava; sem ele, o arquivo é do usuário e fica **intocado**. Esta feature é **extensão própria da 118** (não implementa furo do fix-106-09); o mecanismo de hash da 5.0 que ela cria é reusado pela 118-04.

## Entradas

- Os seis templates de instrução (inclui `OPENAI.md`) e o `shouldPreserve` (`scaffold.go:298-343`).
- `IsGeneratedAgentsMD` (`scaffold.go:348`): cabeçalho + marcador nas três primeiras linhas.

## Saídas

- Marcador nos templates; `shouldPreserve` consulta uma função única `IsGeneratedInstrucao(path)` (a de `AGENTS.md` generalizada, sem duplicar a regra).
- **Adoção do arquivo da 5.0 (sem marcador):** só se o conteúdo for **byte a byte igual** ao template 5.0 renderizado com o `.sddrc` do projeto (com a stack recuperada pela 118-01); aí é regravado com o marcador. Se diferir, é preservado, o `update` avisa e grava `<nome>.novo.md` ao lado. O cabeçalho `# CLAUDE.md — <proj>` **nunca** autoriza regravar.
- Mecanismo: tabela `embed.FS` com os templates 5.0 (ou o hash de cada renderização) em `internal/scaffold`; a 118-04 a estende para rules e `schema.json` (um só mecanismo, `DUP-06`).
- O `update` avisa, em stderr, cada arquivo de instrução preservado.

## Modelo do módulo

**Enxuto** (`ARQ-11`): só `internal/scaffold` (templates e `shouldPreserve`) e uma linha de aviso no `update`. Justificativa: generaliza uma regra que já existe para um arquivo.

## Decisões do projeto aplicadas

- `SEG-02` e `DUP-06`: **uma** função decide "gerado ou do usuário", usada por `AGENTS.md` e pelos demais.
- `TST-02`: templates e goldens mudam no mesmo PR (`TestGoldenInit`).
- Sem heurística de cabeçalho: adoção só por igualdade com o template 5.0 renderizado; a fixture da 118-07 prova.
- `OPENAI.md` existe em `templates/agents/openai/OPENAI.md.tmpl` e entra na lista como os demais.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Regra de arquivo gerado | `scaffold.go:337-339` e `IsGeneratedAgentsMD` | só para `AGENTS.md` | **estender** | generalizar, não copiar (`DUP-06`) |
| Preservação | `shouldPreserve` `scaffold.go:298` | arquivos fora de `sdd/` e `.agents/` sempre regravados (`return false`) | **estender** | causa do achado 3 |
| Templates de instrução | `templates/agents/{claude/CLAUDE.md,gemini/GEMINI.md,gemini/.gemini/system_instructions.md,openai/OPENAI.md,openai/.openai/system_instructions.md}.tmpl`, `templates/.github/copilot-instructions.md.tmpl` | sem marcador | **estender** | marcador nas primeiras linhas |
| Teste de preservação | `scaffold_test.go`, `golden_test.go` | `TestGoldenInit`, sem caso de edição | **criar** casos | `TST-03` |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-02` (preserva domínio do usuário, sem perda), `DUP-06` (uma regra de "gerado"), `DUP-03`, `DUP-02`, `TST-02` (goldens), `TST-03` (edição preservada, gerado atualizado, 5.0 igual adotado, 5.0 divergente com `.novo`, só cabeçalho não adotado; todos incluindo `OPENAI.md`), `ARQ-11`.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**; `TestUpdateNaoPerdeDadoDoUsuario` já existe (criado na 118-01) e ganha o subteste `edicao_em_instrucao` aqui.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestInstrucaoEditadaFicaIntocada|TestInstrucaoGeradaEAtualizada|TestInstrucaoDe50IgualAoTemplateEAdotada|TestInstrucaoDe50DivergenteGeraNovo|TestInstrucaoComSoOCabecalhoNaoEAdotada)$' -count=1
go test ./internal/scaffold -run '^TestUpdateNaoPerdeDadoDoUsuario$' -count=1
go test ./cmd/forge-sdd -run '^TestUpdateAvisaInstrucaoPreservada$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestInstrucaoEditadaFicaIntocada|TestInstrucaoGeradaEAtualizada|TestInstrucaoDe50IgualAoTemplateEAdotada|TestInstrucaoDe50DivergenteGeraNovo|TestInstrucaoComSoOCabecalhoNaoEAdotada|TestUpdateNaoPerdeDadoDoUsuario|TestUpdateAvisaInstrucaoPreservada)$' | grep -c '^Test')" -eq 7 ]
go test ./internal/scaffold ./cmd/forge-sdd -count=1
```

## Fora do escopo

- Skills `sdd-*`, prompts e comandos por agente seguem sempre regenerados; o aviso honesto vai no texto do `/upgrade-sdd` (118-05).
- `--dry-run` listar os arquivos (118-06b).

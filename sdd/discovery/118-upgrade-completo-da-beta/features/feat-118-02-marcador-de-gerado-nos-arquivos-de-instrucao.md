# feat-118-02 — Arquivos de instrução com marcador de gerado: edição do usuário fica intocada

**Depende de:** feat-118-01 (usa `stack` e `db` do `.sddrc`; ambas reescrevem os adaptadores)
**Achados:** 3 e 10 (parte dos adaptadores). **Tamanho:** `M`. **Status:** `spec`

## Objetivo

`CLAUDE.md`, `GEMINI.md`, `.gemini/system_instructions.md`, `.openai/system_instructions.md` e `.github/copilot-instructions.md` ganham o marcador `<!-- forge-sdd:generated -->`. Com o marcador, o `update` regrava; sem ele, o arquivo é do usuário e fica **intocado** (o `AGENTS.md` já funciona assim). Implementa, para os adaptadores, a parte do contrato do fix-106-09 reservada à 118-02 pelo `plan.md` (furos 1, 3 e 6, distribuídos entre 02, 04 e 07); **não** há implementação separada do 106-09 (`DUP-03`).

## Entradas

- Os cinco templates de instrução e o `shouldPreserve` (`scaffold.go:298-343`).
- `IsGeneratedAgentsMD` (`scaffold.go:348`): cabeçalho + marcador nas três primeiras linhas.

## Saídas

- Marcador nos templates; `shouldPreserve` consulta uma função única `IsGeneratedInstrucao(path)` (a de `AGENTS.md` generalizada, sem duplicar a regra).
- Arquivo da 5.0 **sem** marcador mas com o cabeçalho gerado conhecido: adotado uma vez (ganha o marcador na regravação). Sem o cabeçalho: preservado.
- O `update` avisa, em stderr, cada arquivo de instrução preservado por edição do usuário.

## Modelo do módulo

**Enxuto** (`ARQ-11`): só `internal/scaffold` (templates e `shouldPreserve`) e uma linha de aviso no `update`. Justificativa: generaliza uma regra que já existe para um arquivo.

## Decisões do projeto aplicadas

- `SEG-02` e `DUP-06`: **uma** função decide "gerado ou do usuário", usada por `AGENTS.md` e pelos demais.
- `TST-02`: templates e goldens mudam no mesmo PR (`TestGoldenInit`).
- A adoção do arquivo 5.0 por cabeçalho é a única heurística; o Builder a confirma contra a fixture da 118-07 e, se o cabeçalho não bastar, preserva e avisa (nunca sobrescreve).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Regra de arquivo gerado | `scaffold.go:337-339` e `IsGeneratedAgentsMD` | só para `AGENTS.md` | **estender** | generalizar, não copiar (`DUP-06`) |
| Preservação | `shouldPreserve` `scaffold.go:298` | arquivos fora de `sdd/` e `.agents/` sempre regravados (`return false`) | **estender** | causa do achado 3 |
| Templates de instrução | `templates/agents/{claude,gemini,openai}/…`, `templates/.github/copilot-instructions.md.tmpl` | sem marcador | **estender** | marcador nas primeiras linhas |
| Teste de preservação | `scaffold_test.go`, `golden_test.go` | `TestGoldenInit`, sem caso de edição | **criar** casos | `TST-03` |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-02` (preserva domínio do usuário, sem perda), `DUP-06` (uma regra de "gerado"), `DUP-03`, `DUP-02`, `TST-02` (goldens), `TST-03` (edição preservada, gerado atualizado, 5.0 adotado), `ARQ-11`.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**; `TestUpdateNaoPerdeDadoDoUsuario` já existe (criado na 118-01) e ganha o subteste `edicao_em_instrucao` aqui.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestInstrucaoEditadaFicaIntocada|TestInstrucaoGeradaEAtualizada|TestInstrucaoDe50AdotaOMarcador)$' -count=1
go test ./internal/scaffold -run '^TestUpdateNaoPerdeDadoDoUsuario$' -count=1
go test ./cmd/forge-sdd -run '^TestUpdateAvisaInstrucaoPreservada$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestInstrucaoEditadaFicaIntocada|TestInstrucaoGeradaEAtualizada|TestInstrucaoDe50AdotaOMarcador|TestUpdateNaoPerdeDadoDoUsuario|TestUpdateAvisaInstrucaoPreservada)$' | grep -c '^Test')" -eq 5 ]
go test ./internal/scaffold ./cmd/forge-sdd -count=1
```

## Fora do escopo

- Skills `sdd-*`, prompts e comandos por agente seguem sempre regenerados; o aviso honesto vai no texto do `/upgrade-sdd` (118-05).
- `--dry-run` listar os arquivos (118-06b).

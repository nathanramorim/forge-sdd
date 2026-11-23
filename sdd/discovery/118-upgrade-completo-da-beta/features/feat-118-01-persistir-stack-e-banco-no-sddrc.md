# feat-118-01 — Persistir stack e banco no `.sddrc` e repopular nos adaptadores

**Depende de:** nenhuma (onda 2; a 118-02 depende desta)
**Achado:** 2 (`update --upgrade` apaga a stack). **Tamanho:** `M`. **Status:** `spec`

## Objetivo

`forge update` nunca apaga "Stack:" e "Banco:" dos arquivos de instrução. O `.sddrc` passa a guardar `stack` e `db`; o `update` os lê e os devolve aos adaptadores (`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.gemini/` e `.openai/system_instructions.md`, `.github/copilot-instructions.md`).

## Entradas

- `sdd/.sddrc` (5.0, sem `stack`/`db`; ou novo, com eles) e os adaptadores já gerados.
- `config.Config.Stack` e `config.Config.DB` (`internal/config/config.go:159-160`, padrões `go`/`none`, linhas 173-174).

## Saídas

- `.sddrc` com `"stack"` e `"db"`; `Sddrc.ToConfig()` os repassa (hoje não, `config.go:46-55`).
- Projeto 5.0 (sem os campos): o `update` recupera o valor **uma vez** das linhas "Stack:"/"Banco:" dos adaptadores existentes; se não houver valor, deixa vazio. **Nunca** troca por `go`/`none` em silêncio.

## Modelo do módulo

**Enxuto** (`ARQ-11`): toca `internal/config` (campo e leitura) e `internal/scaffold` (template do `.sddrc`, `UpdateSddrc`). Justificativa: uma única direção de dado, sem comando novo.

## Decisões do projeto aplicadas

- `SEG-02`: `UpdateSddrc` já grava de forma atômica (`writeFileSafe`); o valor lido nunca vem de caminho externo.
- `DUP-03`: reusa `UpdateSddrc` e `ReadSddrc`; nenhum leitor novo de config.
- `TST-03`: caso conforme (com valores) e negativo (5.0 sem campos; valor com aspas ou quebra de linha).
- `SEG-01`: o valor recuperado do adaptador é validado na borda (sem quebra de linha, tamanho máximo, sem `{{`).
- Achado verificado: `ToConfig()` não copia `Stack`/`DB`, por isso o `Run` do `update` renderiza em branco.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Campos de stack e banco | `internal/config/config.go:159-160` | `Config.Stack`, `Config.DB` com padrão | **reusar** | só falta persistir |
| Struct do `.sddrc` | `config.go:34-43` | sem `stack`/`db` | **estender** | achado 2 |
| Conversão para Config | `config.go:46-55` `ToConfig` | não copia `Stack`/`DB` | **estender** | causa-raiz |
| Gravação do `.sddrc` | `scaffold.go:280` `UpdateSddrc` e `templates/sdd/.sddrc.tmpl` | só `project`, `version`, `agents`, `telemetry`, `lang`, `naming_convention` | **estender** | template ganha os campos |
| Uso nos adaptadores | `templates/agents/claude/CLAUDE.md.tmpl:3` (`stack {{.Stack}}`) e demais | renderizam `Config` | **reusar** | sem mudança de template |
| Fluxo do `update` | `cmd/forge-sdd/main.go:311-325` | `baseCfg := rc.ToConfig()` | **estender** | recebe os campos |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-02` (gravação atômica, sem perda de dado do usuário), `SEG-01` (valor recuperado validado), `DUP-03` (reusar `UpdateSddrc`/`ReadSddrc`), `DUP-02` (tabela acima), `TST-02` (template do `.sddrc` e golden no mesmo PR), `TST-03` (conforme e negativo), `ARQ-11`.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder** (os testes não existem). `-list` só prova existência; a aprovação é o `-run` com nomes exatos.

```bash
set -o pipefail
go test ./internal/config -run '^(TestSddrcGravaEPreservaStackEBanco|TestSddrcDe50SemStackNaoAssumeDefault)$' -count=1
go test ./internal/scaffold -run '^TestUpdateNaoPerdeDadoDoUsuario$/^stack_e_banco$' -count=1
go test ./internal/scaffold -run '^(TestUpdateRepopulaStackNosAdaptadores|TestRecuperaStackDoAdaptadorRecusaValorInvalido)$' -count=1
go test ./cmd/forge-sdd -run '^TestUpdateUpgradeMantemStackDoSddrc$' -count=1
[ "$(go test ./internal/config ./internal/scaffold ./cmd/forge-sdd -list '^(TestSddrcGravaEPreservaStackEBanco|TestSddrcDe50SemStackNaoAssumeDefault|TestUpdateNaoPerdeDadoDoUsuario|TestUpdateRepopulaStackNosAdaptadores|TestRecuperaStackDoAdaptadorRecusaValorInvalido|TestUpdateUpgradeMantemStackDoSddrc)$' | grep -c '^Test')" -eq 6 ]
go test ./internal/config ./internal/scaffold ./cmd/forge-sdd -count=1
```

`TestUpdateNaoPerdeDadoDoUsuario` é criado aqui com o subteste `stack_e_banco`; a 118-02 acrescenta `edicao_em_instrucao` (o teste fica verde de ponta a ponta só na 118-02).

## Fora do escopo

- Marcador de arquivo gerado e arquivo editado pelo usuário (118-02).
- Detectar a stack lendo o projeto.

# feat-118-06c — JSON inválido devolve erro e não grava versão; terminal sem TTY orienta `--yes`

**Depende de:** feat-118-06b (ambas mexem em `main.go` e no relato do `update`; ordem a, b, c evita conflito)
**Achado:** 9 e o item (e) do escopo. **Tamanho:** `S`. **Status:** `spec`

## Objetivo

Se `settings.json` ou `.mcp.json` for JSON inválido, o arquivo continua intocado, o `update` termina com **código diferente de zero**, **não grava** a versão nova (`sdd/.sdd-version` e `version` do `.sddrc` mudam juntos) e diz o que fazer, uma vez só (sem aviso do MCP duplicado). Sem `--yes` e sem terminal interativo, o comando orienta usar `--yes` em vez de repassar o erro bruto do terminal.

## Entradas

- `MergeClaudePreToolUseHRook`, `MergeClaudeMCPServer` e `installClaudeInboxHook` (`scaffold.go:130-149`), que hoje só imprimem "aviso:" e seguem; `SyncClaudeIntegration` (`hooks.go:190`) devolve `warnings`.
- Fluxo do `update` (`main.go:225-347`): modo interativo chama `survey.RunSmartUpgradePrompt` e `survey.RunUpdate` sem checar TTY.

## Saídas

- Erro tipado de integração (JSON inválido) propagado por `scaffold.Run`; o `update` o trata: não grava a versão, sai com código 1, imprime o arquivo e o conserto manual.
- Mensagem de TTY ausente: "use --yes para rodar sem perguntas".
- Aviso do MCP emitido uma única vez.

## Modelo do módulo

**Completo** (`ARQ-11`): `scaffold` (erro tipado), `cmd/forge-sdd` (ordem de gravação da versão e TTY). Justificativa: muda o contrato de saída do `update`.

## Decisões do projeto aplicadas

- `SEG-01`: JSON de entrada validado na borda; arquivo inválido nunca é reescrito.
- `SEG-02`: a versão só é gravada depois de o resto ter dado certo.
- `DUP-06`: um só ponto decide "integração falhou"; o aviso do MCP sai de `SyncClaudeIntegration` ou do `installClaudeInboxHook`, não dos dois.
- `ARQ-05`: o `--dry-run` continua sem efeito e também relata o JSON inválido, sem falhar.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Mesclagem segura | `hooks.go` `mergeClaudeHookWithMatcher`, `mergeClaudeMCPServer` | não toca em JSON inválido (correto) | **reusar** | só propagar o erro |
| Avisos | `scaffold.go:135-146` e `agent_models.go:44-46` | duas fontes, aviso do MCP duplicado | **estender** (unificar) | achado 9 |
| Gravação da versão | `Run` e `main.go:349-355` | grava e sai 0 mesmo com falha | **estender** | RN-2 |
| Prompt interativo | `main.go:279-309` (`survey`) | sem checagem de TTY | **estender** | item (e) |
| Códigos de saída | `exitCodeErr` (`doctor.go:355`) | padrão existente | **reusar** | `DUP-03` |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`SEG-01`, `SEG-02`, `ARQ-05`, `DUP-06`, `DUP-02`, `ARQ-11`, `TST-03` (JSON inválido, sem TTY, caso conforme).

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestJSONInvalidoNoSettingsRetornaErroTipado|TestJSONInvalidoNaoReescreveOArquivo|TestAvisoDoMCPSaiUmaVez)$' -count=1
go test ./cmd/forge-sdd -run '^(TestUpdateJSONInvalidoSaiComErroENaoGravaVersao|TestUpdateSemTTYOrientaYes|TestUpdateDryRunRelataJSONInvalidoSemFalhar)$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list '^(TestJSONInvalidoNoSettingsRetornaErroTipado|TestJSONInvalidoNaoReescreveOArquivo|TestAvisoDoMCPSaiUmaVez|TestUpdateJSONInvalidoSaiComErroENaoGravaVersao|TestUpdateSemTTYOrientaYes|TestUpdateDryRunRelataJSONInvalidoSemFalhar)$' | grep -c '^Test')" -eq 6 ]
go test ./internal/scaffold ./cmd/forge-sdd -count=1
```

## Fora do escopo

- Reserialização das chaves do `settings.json` (achado 11, cosmético).
- Lista do plano no dry-run (118-06b).

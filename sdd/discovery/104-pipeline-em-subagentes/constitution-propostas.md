# Propostas de mudança da Constituição (iniciativa 104)

Só o dono decide. Nada em `sdd/memory/constitution.md` foi alterado por este alinhamento.

## P-1. ARQ-06: `mcp` e `eval` (já no texto da branch, aguarda aceite)
- **Texto atual (`main`):** lista sem `mcp` e sem `eval`.
- **Texto na branch `feat/104-pipeline-em-subagentes`:** lista com `mcp` (após `acp`) e `eval` (no fim).
- **Motivo:** `forge mcp serve` e `forge eval` existem e são registrados.
- **Evidência:** `cmd/forge-sdd/mcp.go`, `cmd/forge-sdd/eval.go`; `forge --help` do binário compilado lista `mcp` e `eval`; ADR-0015 e ADR-0017.
- **Ação:** o dono aceita ou rejeita no PR da 104.

## P-2. ARQ-06: a lista continua incompleta
- **Texto atual:** omite `brief`, `ps`, `rules`, `version`, `completion` (este último é padrão do cobra).
- **Proposta:** acrescentar `brief`, `ps` e `rules` à lista (ou trocar a lista por "subcomandos determinísticos registrados em `cmd/forge-sdd`", sem enumerar).
- **Evidência:** `forge --help` (binário compilado em 20/11/2026) lista `brief`, `ps`, `rules`; `cmd/forge-sdd/{brief,ps,rules}.go`. `review` não é subcomando de topo: `cmd/forge-sdd/review.go` declara `busReviewCmd` (`forge bus review`); `forge review` responde `unknown command` no binário compilado.

## P-3. SEG-02 e escrita segura (verificado: cobre em parte)
- **Texto atual:** "Escritas respeitam o diretório-alvo e preservam arquivos de domínio do usuário (`shouldPreserve`), sem perda de dados nem path traversal."
- **Lacuna:** não menciona link simbólico nem escrita atômica. Hoje o código recusa alvo ou pai simbólico e grava por temporário e rename (`checkSafeTarget`, `writeFileSafe` em `internal/scaffold/hooks.go`).
- **Proposta (acréscimo a SEG-02, sem ID novo):** "...nem seguem link simbólico para fora do projeto; gravação em arquivo do usuário é atômica (temporário e rename)."
- **Evidência:** `internal/scaffold/hooks.go`; fix de `renderDir` em `fix/104-renderdir-symlink` (`ab61c58`, não mesclado); ADR-0018.

## P-4. Regra nova: saída de agente e de MCP é saneada (sugestão de ID `SEG-08`)
- **Texto atual:** `SEG-05` cobre dado sensível em log; `FLX-08` cobre retorno de agente como dado. Nenhuma exige sanear texto de modelo antes de exibir ou repassar.
- **Proposta:** "SEG-08: texto vindo de modelo, agente ou arquivo externo é saneado (sem ESC, controles nem bidi, com tamanho limitado) antes de ir ao terminal, ao MCP ou a notificações."
- **Evidência:** `mcpSaneia` (`internal/acp/mcp.go`), `notify.Sanitize` (`internal/notify/notify.go`), `sanitizarAchados` (`internal/pipeline/laco.go`). Se aceita, cria ID novo: exige atualizar `.agents/rules/sdd-seguranca.md` (não verificado).

## P-5. Rastreabilidade do `SEG-06`
- `SEG-06` só cita a Governança. O Reviewer do modo chat e seus subagentes de revisão também são somente leitura pela lista de ferramentas (ADR-0015 D-3). Proposta: citar "Reviewer e subagentes de revisão". Evidência: `.claude/agents/reviewer.md` e `governance.md` (`tools: Read, Grep, Glob, Bash`; Bash não é somente leitura por si, verificar).

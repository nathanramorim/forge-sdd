# Plano de Discovery 80 — Status Executivo Tríplice de Baixo Token

Branch única `feat/80-status-executivo-triplice-baixo-token` (`FLX-02`), partindo de `feat/79-clareza-pipeline-e-identificacao-tmux`. Pasta de execução: `sdd/discovery/80-status-executivo-triplice-baixo-token/features/`.

## Ordem e dependências
`80-01 -> 80-02 -> 80-03`

## Quebra de Features
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 80-01 Template Canônico nos Prompts de Agentes | M | — | Atualizar `.gemini/prompts/status.prompt.md`, `.claude/commands/status.md`, `.github/prompts/status.prompt.md`, `.agents/skills/sdd-operacao/references/status.md` e os templates em `internal/scaffold/templates/`. Padronizar a ordem de chamada (capture-pane, queue panel, git log) e o template canônico de saída (< 200 tokens). | `sdd-fluxo-e-portao`, `sdd-codigo-e-testes`; `COD-05`, `DUP-01`, `DUP-10` | — |
| 80-02 Subcomando nativo `forge-sdd status --live` | M | — | Em `cmd/forge-sdd/status.go` e `internal/pipeline`, implementar a flag `--live` no `forge-sdd status`: executa internamente a captura do painel tmux (com fallback para logs de runtime), a checagem da fila e o git log, emitindo o resumo canônico diretamente. Testes em `status_test.go`. | `sdd-arquitetura`, `sdd-codigo-e-testes`; `ARQ-01`, `COD-01`, `COD-03`, `TST-04` | — |
| 80-03 Documentação, LEIA-ME e Validação Final | P | 80-01, 80-02 | Atualizar `.forge/LEIA-ME.md.tmpl`, documentar a flag `--live`, rodar suíte completa (`go test ./...`, `go vet`, `gate`). | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `GOV-01`, `TST-05` | — |

## Critério Final
Aprovação dos critérios CA-01 a CA-05 de `criteria.md`. Resposta idêntica e concisa nas consultas de status em qualquer CLI.

## Pós-Portão
Submissão para aprovação humana (Gate 1 - `FLX-01`). Após aprovação, executar `/split-features` para gerar as tasks em `features/` e entregar ao Builder.

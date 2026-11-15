# Plano de Discovery 79 — Clareza no Pipeline, Identificação no Tmux, Paralelismo de Filas, Monitor de Processos e Menu

Branch única `feat/79-clareza-pipeline-e-identificacao-tmux` (`FLX-02`), partindo de `fix/78-pipeline-gemini-abre-agy`. Pasta de execução: `sdd/discovery/79-clareza-pipeline-e-identificacao-tmux/features/`.

## Ordem e dependências
- **Onda 1 (Paralelo):** `79-01` e `79-02` (independentes, iniciam juntas)
- **Onda 2 (Paralelo após 79-02):** `79-03`, `79-04`, `79-05`, `79-06`
- **Onda 3 (Fechamento):** `79-07`

## Quebra de Features
| Task | Tam. | Depende | Escopo | Rules / IDs | Skills |
|---|---|---|---|---|---|
| 79-01 Diretrizes Pipeline-First | P | — | Atualizar `GEMINI.md`, `CLAUDE.md`, `.agents/skills/sdd-concepcao`, `sdd-operacao` e templates em `internal/scaffold/templates`. Substituir a regra de subagente por inspeção de fila e sugestão de seguir a esteira. | `sdd-fluxo-e-portao`, `sdd-codigo-e-testes`; `COD-05`, `FLX-01`, `DUP-02` | — |
| 79-02 Identificação visual e ergonomia no tmux | M | — | Em `internal/pipeline/tmux.go`, estender `BuildTmuxPlan` para configurar no alvo da sessão (`-t exact`): `status-left` com nome do projeto e branch, `status-left-length 60`, `pane-border-format` com `[ <projeto> ❯ #{pane_title} ]`, janela `forge:<projeto>` e mini-painel compacto de monitoramento. Testes em `tmux_test.go`. | `sdd-arquitetura`, `sdd-codigo-e-testes`; `ARQ-01`, `ARQ-04`, `COD-01`, `COD-03`, `TST-04` | — |
| 79-03 Menu unificado e enxuto (Mapa De/Para) | M | 79-02 | Em `cmd/forge-sdd/menu.go`, eliminar a divisão artificial entre "pipeline" e "estrutura padrão". Implementar a navegação direta com cabeçalho rico e opções promovidas conforme mapa de migração. Atualizar testes de menu. | `sdd-arquitetura`, `sdd-codigo-e-testes`; `ARQ-01`, `COD-01`, `COD-05`, `TST-04` | — |
| 79-04 Alertas visuais e fluxo de aprovação não bloqueante | M | 79-02 | Tratamento do estado de ociosidade sem poluição de loop na tela; badge/borda de alerta no painel do worker quando aprovação humana for demandada (Camada C). | `sdd-codigo-e-testes`, `sdd-seguranca`; `COD-01`, `SEG-01`, `TST-04` | — |
| 79-05 Paralelismo de fila por DAG | M | 79-02 | Em `internal/queue/`, adicionar `depends_on` no item da fila, parser automático do markdown da spec no `enqueue`, e resolvedor topológico de elegibilidade em ondas paralelas para o Builder. | `sdd-arquitetura`, `sdd-codigo-e-testes`; `ARQ-01`, `ARQ-02`, `COD-01`, `COD-03`, `TST-04` | — |
| 79-06 Monitor de processos do projeto (`forge-sdd ps`) | M | 79-02 | Novo pacote `internal/monitor/` e comando `forge-sdd ps`: mapeia painéis tmux, CLIs, workers e subprocessos associados ao projeto, com PID, % CPU, Memória RSS, tempo ativo e `--kill-zombies`. | `sdd-arquitetura`, `sdd-seguranca`; `ARQ-01`, `ARQ-04`, `COD-01`, `SEG-01`, `TST-04` | — |
| 79-07 Documentação, LEIA-ME e validação final | P | 79-01, 79-03, 79-04, 79-05, 79-06 | Atualizar `.forge/LEIA-ME.md.tmpl`, documentação de uso do menu, comando `ps` e paralelismo de fila; rodar suíte completa (`go test ./...`, `go vet`, `gate`). | `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`; `DUP-10`, `FLX-05`, `GOV-01`, `TST-05` | — |

## Critério Final
Aprovação dos critérios CA-01 a CA-07 de `criteria.md`. Verificação manual ou com mocks do comando `forge-sdd ps` e do agendamento DAG de fila.

## Pós-Portão
Submissão para aprovação humana (Gate 1 - `FLX-01`). Após aprovação, executar `/split-features` para atualizar a fila e entregar ao Builder.
